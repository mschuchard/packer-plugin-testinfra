package testinfra

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

// execute testinfra remotely with *exec.Cmd
func execCmd(cmd *exec.Cmd, ui packer.Ui) error {
	// merge in env settings
	cmd.Env = os.Environ()

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
		ui.Error("initialization of Testinfra py.test command execution returned non-zero exit status")
		return err
	}

	// capture and display testinfra output
	outSlurp, err := io.ReadAll(stdout)
	if err != nil {
		ui.Error("unable to read stdout from Testinfra")
		return err
	}
	if len(outSlurp) > 0 {
		ui.Say("Testinfra results include the following:")
		ui.Say(string(outSlurp))
	} else {
		ui.Say("Testinfra produced no stdout; it is probable that something unintended occurred during execution")
	}

	errSlurp, err := io.ReadAll(stderr)
	if err != nil {
		ui.Error("unable to read stderr from Testinfra")
		return err
	}
	if len(errSlurp) > 0 {
		ui.Error("Testinfra errored internally during execution:")
		ui.Error(string(errSlurp))
		return errors.New("Testinfra internal execution error")
	}

	// wait for testinfra to complete and flush buffers
	if err = cmd.Wait(); err != nil {
		ui.Error("Testinfra returned non-zero exit status")
		return err
	}

	// finish and return
	ui.Say("Testinfra machine image testing is complete")

	return nil
}

// execute testinfra local to temp packer instance with packer.RemoteCmd
func packerRemoteCmd(localCmd *packer.RemoteCmd, installCmd []string, comm packer.Communicator, ui packer.Ui) error {
	// initialize context and log command
	ctx := context.Background()

	// install testinfra on temp packer instance
	if len(installCmd) > 0 {
		// cast installCmd to string, log, and init localInstallCmd
		strInstallCmd := strings.Join(installCmd, " ")
		ui.Say("installing Testinfra on instance")
		log.Printf("Testinfra installation command is: %s", strInstallCmd)
		localInstallCmd := &packer.RemoteCmd{Command: strInstallCmd}

		// install testinfra on temp packer instance
		if err := comm.Start(ctx, localInstallCmd); err != nil {
			ui.Error("Testinfra install command execution returned non-zero exit status")
			return err
		}
	}

	// initialize stdout and stderr as bytes
	var stdout, stderr bytes.Buffer
	localCmd.Stdout = &stdout
	localCmd.Stderr = &stderr

	// initialize testinfra tests
	ui.Say("beginning Testinfra validation of machine image")
	if err := comm.Start(ctx, localCmd); err != nil {
		ui.Error("initialization of Testinfra py.test command execution returned non-zero exit status")
		return err
	}

	// wait for testinfra to complete and flush buffers
	// then check for pytest/testinfra execution issues
	if exitStatus := localCmd.Wait(); exitStatus > 0 || len(stderr.String()) > 0 {
		ui.Error("Testinfra errored internally during execution:")
		ui.Error(stderr.String())
		ui.Errorf("Testinfra returned exit status: %d", exitStatus)
		return errors.New("testinfra non-zero exit code")
	}

	// capture and display testinfra output
	if len(stdout.String()) > 0 {
		ui.Say("Testinfra results include the following:")
		ui.Say(stdout.String())
	} else {
		ui.Say("Testinfra produced no stdout; it is likely something unintended occurred during execution")
	}

	// finish and return
	ui.Say("Testinfra machine image testing is complete")

	return nil
}

// determine and return execution command for testinfra
func (provisioner *Provisioner) determineExecCmd(ui packer.Ui) (*exec.Cmd, *packer.RemoteCmd, error) {
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
		args = append(args, "--no-header", "--no-summary", "--disable-warnings")
	}
	// keyword
	keyword, err := interpolate.Render(provisioner.config.Keyword, &provisioner.config.ctx)
	if err != nil {
		ui.Errorf("error parsing config for Keyword: %v", err.Error())
		return nil, nil, err
	}
	if len(keyword) > 0 {
		args = append(args, "-k", fmt.Sprintf("\"%s\"", keyword))
	}
	// marker
	marker, err := interpolate.Render(provisioner.config.Marker, &provisioner.config.ctx)
	if err != nil {
		ui.Errorf("error parsing config for Marker: %v", err.Error())
		return nil, nil, err
	}
	if len(marker) > 0 {
		args = append(args, "-m", fmt.Sprintf("\"%s\"", marker))
	}
	// parallel
	if provisioner.config.Parallel {
		args = append(args, "-n", "auto")
	}
	// sudo
	if provisioner.config.Sudo {
		args = append(args, "--sudo")
	} else if len(provisioner.config.SudoUser) > 0 { // sudo_user
		args = append(args, fmt.Sprintf("\"--sudo-user=%s\"", provisioner.config.SudoUser))
	}

	// verbose
	if provisioner.config.Verbose > 0 {
		// initialize arg
		levelArg := "-"
		// append correct number of "v" based on verbosity level
		levelArg += strings.Repeat("v", provisioner.config.Verbose)

		args = append(args, levelArg)
	}

	// testfiles
	args = slices.Concat(args, provisioner.config.TestFiles)

	// return packer remote command for local testing on instance
	if localExec {
		// prepend pytest path to args for command string slice
		command := slices.Insert(args, 0, pytestPath)
		return nil, &packer.RemoteCmd{Command: strings.Join(command, " ")}, nil
	} else { // return exec command for remote testing against instance
		// initialize cmd
		cmd := exec.Command(pytestPath, args...)
		// determine if user requested execution in different directory
		if len(provisioner.config.Chdir) > 0 {
			cmd.Dir = provisioner.config.Chdir
		}

		return cmd, nil, nil
	}
}
