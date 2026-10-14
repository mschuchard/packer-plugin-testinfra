package testinfra

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"slices"
	"strings"
	"sync"

	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

// execute testinfra remotely with *exec.Cmd
func execCmd(cmd *exec.Cmd, ui packer.Ui) error {
	// prepare stdout and stderr pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		ui.Error("unable to prepare the pipe for capturing testinfra stdout")
		return err
	}
	defer stdout.Close()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		ui.Error("unable to prepare the pipe for capturing testinfra stderr")
		return err
	}
	defer stderr.Close()

	// initialize testinfra tests
	ui.Say("Beginning Testinfra validation of machine image")
	if err := cmd.Start(); err != nil {
		ui.Error("initialization of Testinfra py.test command execution returned error")
		return err
	}

	// read stdout and stderr concurrently to avoid deadlock if either pipe buffer fills
	// while the other is still being drained (the child blocks on a full pipe otherwise)
	var outSlurp, errSlurp []byte
	var outErr, errErr error
	var wg sync.WaitGroup
	wg.Add(2)

	// define goroutines to read stdout and stderr concurrently
	go func() {
		defer wg.Done()
		outSlurp, outErr = io.ReadAll(stdout)
	}()
	go func() {
		defer wg.Done()
		errSlurp, errErr = io.ReadAll(stderr)
	}()

	// wait for both goroutines to finish reading stdout and stderr
	wg.Wait()
	if outErr != nil {
		ui.Error("unable to read stdout from Testinfra")
		return outErr
	}
	if errErr != nil {
		ui.Error("unable to read stderr from Testinfra")
		return errErr
	}

	// display testinfra stdout and stderr
	if len(outSlurp) > 0 {
		ui.Say("Testinfra results include the following:")
		ui.Say(string(outSlurp))
	} else {
		ui.Say("Testinfra produced no stdout; it is probable that something unintended occurred during execution")
	}
	if len(errSlurp) > 0 {
		ui.Say("Testinfra stderr output includes the following:")
		ui.Say(string(errSlurp))
	}

	// reap the child process; both pipes are already fully drained above
	if err = cmd.Wait(); err != nil {
		ui.Error("Testinfra returned non-zero exit status")
		return err
	}

	// finish and return
	ui.Say("Testinfra machine image testing is complete")

	return nil
}

// execute testinfra local to temp packer instance with packer.RemoteCmd
func packerRemoteCmd(ctx context.Context, localCmd *packer.RemoteCmd, installCmd []string, comm packer.Communicator, ui packer.Ui) error {
	// install testinfra on temp packer instance
	if len(installCmd) > 0 {
		// cast installCmd to string, log, and init localInstallCmd
		strInstallCmd := strings.Join(installCmd, " ")
		ui.Say("installing Testinfra on instance")
		log.Printf("Testinfra installation command is: %s", strInstallCmd)
		localInstallCmd := &packer.RemoteCmd{Command: strInstallCmd}

		// install testinfra on temp packer instance
		if err := localInstallCmd.RunWithUi(ctx, comm, ui); err != nil {
			ui.Error("Testinfra install command execution failed or was cancelled")
			return err
		}
		// RunWithUi blocks until the exit channel closes, so Wait() here just reads the cached status
		if localInstallCmd.ExitStatus() != 0 {
			ui.Errorf("Testinfra install command returned non-zero exit status: %d", localInstallCmd.ExitStatus())
			return errors.New("testinfra local install non-zero exit code")
		}
	}

	// initialize testinfra tests while streaming stdout/stderr live to ui as the command runs
	ui.Say("beginning Testinfra validation of machine image")
	if err := localCmd.RunWithUi(ctx, comm, ui); err != nil {
		ui.Error("Testinfra py.test command execution failed or was cancelled")
		return err
	}

	// RunWithUi blocks until the exit channel closes, so Wait() here just reads the cached status
	if localCmd.ExitStatus() > 0 {
		ui.Errorf("Testinfra returned non-zero exit status: %d", localCmd.ExitStatus())
		return errors.New("testinfra non-zero exit code")
	}

	// finish and return
	ui.Say("Testinfra machine image testing is complete")

	return nil
}

// determine and return execution command for testinfra
func (provisioner *Provisioner) determineExecCmd(ctx context.Context, ui packer.Ui) (*exec.Cmd, *packer.RemoteCmd, error) {
	// declare args
	var args []string

	// assign determined communication string
	localExec := provisioner.config.Local
	if !localExec {
		communication, err := provisioner.determineCommunication(ui)
		if err != nil {
			ui.Error("could not accurately determine packer communication configuration")
			return nil, nil, err
		}

		args = slices.Concat(args, communication)
	}

	// assign mandatory populated values
	// pytest path
	pytestPath, err := interpolate.Render(provisioner.config.PytestPath, &provisioner.config.ctx)
	if err != nil {
		ui.Errorf("error parsing config for PytestPath: %v", err.Error())
		return nil, nil, err
	}

	// assign optional populated values
	// compact
	if provisioner.config.Compact {
		args = append(args, "--no-header", "--no-summary", "--disable-warnings", "--force-short-summary")
	}
	// junit xml
	if len(provisioner.config.JUnitXML) > 0 {
		args = append(args, fmt.Sprintf("--junit-xml=%s", provisioner.config.JUnitXML))
	}
	// keyword
	keyword, err := interpolate.Render(provisioner.config.Keyword, &provisioner.config.ctx)
	if err != nil {
		ui.Errorf("error parsing config for Keyword: %v", err.Error())
		return nil, nil, err
	}
	if len(keyword) > 0 {
		args = append(args, "-k", keyword)
	}
	// log level
	if len(provisioner.config.LogLevel) > 0 {
		args = append(args, fmt.Sprintf("--log-level=%s", strings.ToUpper(provisioner.config.LogLevel)))
	}
	// marker
	marker, err := interpolate.Render(provisioner.config.Marker, &provisioner.config.ctx)
	if err != nil {
		ui.Errorf("error parsing config for Marker: %v", err.Error())
		return nil, nil, err
	}
	if len(marker) > 0 {
		args = append(args, "-m", marker)
	}
	// parallel
	if provisioner.config.Parallel {
		args = append(args, "-n", "auto")
	}
	// sudo
	if provisioner.config.Sudo {
		args = append(args, "--sudo")
	} else if len(provisioner.config.SudoUser) > 0 { // sudo_user
		args = append(args, fmt.Sprintf("--sudo-user=%s", provisioner.config.SudoUser))
	}

	// verbose
	if provisioner.config.Verbose > 0 {
		// initialize arg
		levelArg := "-"
		// append correct number of "v" based on verbosity level
		levelArg += strings.Repeat("v", provisioner.config.Verbose)

		args = append(args, levelArg)
	}

	// test files
	// initialize interpolated test files
	testFiles := make([]string, len(provisioner.config.TestFiles))

	// iteratively interpolate test files
	for index, testFile := range provisioner.config.TestFiles {
		// interpolate test file
		rendered, err := interpolate.Render(testFile, &provisioner.config.ctx)
		if err != nil {
			ui.Errorf("error parsing config for TestFiles: %v", err.Error())
			return nil, nil, err
		}

		// add interpolated test file to testFiles slice
		testFiles[index] = rendered
	}
	// add testfiles to end of args slice for command string slice
	args = slices.Concat(args, testFiles)

	// return packer remote command for local testing on instance
	if localExec {
		// prepend pytest path to args for command string slice
		command := slices.Insert(args, 0, pytestPath)
		return nil, &packer.RemoteCmd{Command: strings.Join(command, " ")}, nil
	} else { // return exec command for remote testing against instance
		// initialize cmd
		cmd := exec.CommandContext(ctx, pytestPath, args...)
		// determine if user requested execution in different directory
		if len(provisioner.config.Chdir) > 0 {
			cmd.Dir = provisioner.config.Chdir
		}

		// convert interpolated env vars parameter from map to slice
		if len(provisioner.config.EnvVars) > 0 {
			// initialize interpolated env vars
			envVars := make(map[string]string, len(provisioner.config.EnvVars))

			// iteratively interpolate env vars
			for key, value := range provisioner.config.EnvVars {
				// interpolate env var key
				renderedKey, err := interpolate.Render(key, &provisioner.config.ctx)
				if err != nil {
					ui.Errorf("error parsing config for EnvVars key: %s", err.Error())
					return nil, nil, err
				}

				// interpolate env var value
				renderedValue, err := interpolate.Render(value, &provisioner.config.ctx)
				if err != nil {
					ui.Errorf("error parsing config for EnvVars value: %s", err.Error())
					return nil, nil, err
				}

				// add interpolated key-value pair to envVars map
				envVars[renderedKey] = renderedValue
			}

			// initialize envVars string slice
			envVarsSlice := make([]string, 0, len(envVars))

			// convert interpolated envVars map to envVars slice
			for key, value := range envVars {
				envVarsSlice = append(envVarsSlice, fmt.Sprintf("%s=%s", key, value))
			}

			// concat os environment with envVars
			cmd.Env = slices.Concat(os.Environ(), envVarsSlice)
		} else {
			// initialize environment variables from os
			cmd.Env = os.Environ()
		}

		return cmd, nil, nil
	}
}
