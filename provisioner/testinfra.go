//go:generate packer-sdc mapstructure-to-hcl2 -type Config
package testinfra

import (
	"context"
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

// config data deserialized/unmarshalled from packer template/config
type Config struct {
	Chdir          string            `mapstructure:"chdir" required:"false"`
	Compact        bool              `mapstructure:"compact" required:"false"`
	DestinationDir string            `mapstructure:"destination_dir" required:"false"`
	EnvVars        map[string]string `mapstructure:"env_vars" required:"false"`
	InstallCmd     []string          `mapstructure:"install_cmd" required:"false"`
	Keyword        string            `mapstructure:"keyword" required:"false"`
	Local          bool              `mapstructure:"local" required:"false"`
	Marker         string            `mapstructure:"marker" required:"false"`
	Parallel       bool              `mapstructure:"parallel" required:"false"`
	PytestPath     string            `mapstructure:"pytest_path" required:"false"`
	Sudo           bool              `mapstructure:"sudo" required:"false"`
	SudoUser       string            `mapstructure:"sudo_user" required:"false"`
	TestFiles      []string          `mapstructure:"test_files" required:"false"`
	Verbose        int               `mapstructure:"verbose" required:"false"`

	ctx interpolate.Context
}

// implements the packer.Provisioner interface as testinfra.Provisioner
type Provisioner struct {
	config        Config
	generatedData map[string]any
}

// implements configspec with hcl2spec helper function
func (provisioner *Provisioner) ConfigSpec() hcldec.ObjectSpec {
	return provisioner.config.FlatMapstructure().HCL2Spec()
}

// prepares the provisioner plugin
func (provisioner *Provisioner) Prepare(raws ...any) error {
	// parse testinfra provisioner config
	err := config.Decode(&provisioner.config, &config.DecodeOpts{
		PluginType:         "testinfra",
		Interpolate:        true,
		InterpolateContext: &provisioner.config.ctx,
	}, raws...)
	if err != nil {
		log.Print("error decoding the supplied Packer config")
		return err
	}

	// set default executable path for py.test
	if len(provisioner.config.PytestPath) == 0 {
		log.Print("setting PytestPath to default 'py.test'")
		provisioner.config.PytestPath = "py.test"
	} else if info, err := os.Stat(provisioner.config.PytestPath); err != nil || info.IsDir() { // verify valid py.test exists at supplied path
		log.Printf("the Pytest executable does not exist, is not a file, or cannot be accessed at: %s", provisioner.config.PytestPath)

		if err != nil {
			return err
		} else {
			return errors.New("pytest installation issue")
		}
	}

	// log optional arguments
	// local parameter
	if provisioner.config.Local {
		// no validation of testinfra installation
		log.Print("test execution will occur on the temporary Packer instance used for building the machine image artifact")
		log.Print("Testinfra validation does not occur with local execution")

		if len(provisioner.config.InstallCmd) > 0 {
			log.Printf("installation command on the temporary Packer instance prior to Testinfra test execution is: %s", strings.Join(provisioner.config.InstallCmd, " "))
		}

		if len(provisioner.config.DestinationDir) > 0 {
			log.Printf("test files will be copied to '%s' at the temporary Packer instance prior to Testinfra test execution", provisioner.config.DestinationDir)
		}

		// no validation of xdist installation
		if provisioner.config.Parallel {
			log.Print("pytest-xdist validation does not occur with local execution")
			log.Print("Testinfra tests will execute in parallel across the available physical CPUs if possible")
		}

		// environment variables
		if len(provisioner.config.EnvVars) > 0 {
			log.Print("environment variables cannot be set for local execution, and this parameter will be ignored")
		}
	} else { // verify testinfra installed
		// chdir parameter
		if len(provisioner.config.Chdir) > 0 {
			// verify chdir exists and is directory
			if info, err := os.Stat(provisioner.config.Chdir); err != nil || !info.IsDir() {
				log.Printf("the chdir does not exist, is not a directory, or cannot be accessed at: %s", provisioner.config.Chdir)

				if err != nil {
					return err
				} else {
					return errors.New("chdir path issue")
				}
			} else {
				log.Printf("test execution will occur within the following directory: %s", provisioner.config.Chdir)
			}
		}

		// environment variables
		if len(provisioner.config.EnvVars) > 0 {
			log.Printf("environment variables '%v' will be set for the Testinfra execution", provisioner.config.EnvVars)
		}

		log.Print("beginning Testinfra installation verification")

		// initialize testinfra -h command
		cmd := exec.Command(provisioner.config.PytestPath, []string{"-h"}...)

		// prepare stdout pipe
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Print("unable to prepare the pipe for capturing stdout")
			return err
		}
		defer stdout.Close()

		// initialize testinfra installed check
		if err := cmd.Start(); err != nil {
			log.Printf("initialization of Testinfra 'py.test -h' command execution returned non-zero exit status: %s", err.Error())
			return err
		}

		// capture pytest stdout
		outSlurp, err := io.ReadAll(stdout)
		if err != nil {
			log.Printf("unable to read stdout from Pytest: %s", err.Error())
			return err
		}

		// examine pytest stdout
		if len(outSlurp) > 0 {
			// validate minimum version 8.4.0
			if strings.Contains(string(outSlurp), "--force-short-summary") {
				log.Print("pytest version greater than or equal to 8.4.0 verified")
			} else {
				log.Print("the version of the specified pytest is lower than the minimum required version of 8.4.0")
				return errors.New("unsupported version of pytest specified")
			}
			// check for testinfra in stdout
			if strings.Contains(string(outSlurp), "testinfra") {
				log.Print("testinfra installation existence verified")
			} else {
				log.Print("testinfra installation not found by specified Pytest installation")
				return errors.New("testinfra installation not found")
			}

			if provisioner.config.Parallel {
				// check for xdist in pytest usage stdout
				if strings.Contains(string(outSlurp), " -n ") {
					log.Print("Testinfra tests will execute in parallel across the available physical CPUs")
				} else {
					log.Printf("pytest-xdist is not installed, and processes parameter will be reset to default")
					provisioner.config.Parallel = false
				}
			}
		} else {
			// pytest returned no stdout
			log.Print("pytest help command returned no stdout; this indicates an issue with the specified Pytest installation")
			return errors.New("pytest installation issue occurred")
		}
	}

	log.Print("Testinfra installation verified")

	// compact parameter
	if provisioner.config.Compact {
		log.Print("pytest report will be in compact form")
	}

	// keyword parameter
	if len(provisioner.config.Keyword) > 0 {
		log.Printf("executing tests with keyword substring expression: %s", provisioner.config.Keyword)
	}

	// marker parameter
	if len(provisioner.config.Marker) > 0 {
		log.Printf("executing tests with marker expression: %s", provisioner.config.Marker)
	}

	// sudo and sudo_user parameters
	if provisioner.config.Sudo {
		log.Print("testinfra will execute with sudo")

		// warn if sudo_user also specified
		if len(provisioner.config.SudoUser) > 0 {
			log.Print("the 'sudo_user' parameter is ignored when sudo is enabled")
		}
	} else {
		log.Print("testinfra will not execute with sudo")

		// sudo_user mutually exclusive with sudo
		if len(provisioner.config.SudoUser) > 0 {
			log.Printf("testinfra will execute as user: %s", provisioner.config.SudoUser)
		}
	}

	// verbose parameter
	if provisioner.config.Verbose > 0 {
		// validate value does not exceed maximum
		if provisioner.config.Verbose > 4 {
			log.Printf("verbose parameter value was set to value higher than 4 and will therefore be reset to the maximum value of 4")
			provisioner.config.Verbose = 4
		}

		log.Printf("pytest will execute with verbose enabled at level %d", provisioner.config.Verbose)
	}

	// check if testinfra files are specified as inputs
	if len(provisioner.config.TestFiles) == 0 {
		log.Print("all files prefixed with 'test_' recursively discovered from the current working directory will be considered Testinfra test files")
	} else { // verify testinfra files exist
		for _, testFile := range provisioner.config.TestFiles {
			if info, err := os.Stat(testFile); err != nil || info.IsDir() {
				log.Printf("the Testinfra test_file does not exist, is not a file, or cannot be accessed at: %s", testFile)

				if err != nil {
					return err
				} else {
					return errors.New("test file path issue")
				}
			}
		}
	}

	log.Print("packer plugin testinfra validation complete")

	return nil
}

// executes the provisioner plugin
func (provisioner *Provisioner) Provision(ctx context.Context, ui packer.Ui, comm packer.Communicator, generatedData map[string]any) error {
	ui.Say("testing machine image with Testinfra")

	// prepare generated data and context
	provisioner.generatedData = generatedData
	provisioner.config.ctx.Data = generatedData

	// prepare testinfra test command
	cmd, localCmd, err := provisioner.determineExecCmd(ui)
	if cmd != nil {
		log.Printf("complete Testinfra remote command is: %s", cmd.String())
	} else if localCmd != nil {
		log.Printf("complete Testinfra local command is: %s", localCmd.Command)
	}
	if err != nil {
		ui.Error("the Testinfra execution command could not be accurately determined")
		return err
	}

	// execute testinfra remotely with *exec.Cmd
	if localCmd == nil && cmd != nil {
		err = execCmd(cmd, ui)
	} else if localCmd != nil && cmd == nil {
		// testinfra local execution
		if len(provisioner.config.DestinationDir) > 0 {
			// upload testinfra files to temporary packer instance
			if err = uploadFiles(comm, provisioner.config.TestFiles, provisioner.config.DestinationDir); err != nil {
				ui.Error("the test files could not be transferred to the temporary Packer instance")
				return err
			}
		}

		// execute testinfra local to instance with packer.RemoteCmd
		err = packerRemoteCmd(localCmd, provisioner.config.InstallCmd, comm, ui)
	} else {
		// somehow we either returned both commands or neither or something really weird for one or both
		ui.Errorf("incorrectly determined Testinfra remote command (%s) and/or command local to instance (%s); please report as bug with this log information", cmd.String(), localCmd.Command)
		return errors.New("failed pytest command determination")
	}
	if err != nil {
		ui.Error("the Pytest Testinfra execution failed")
		return err
	}

	ui.Say("packer plugin testinfra provisioning complete")

	return nil
}
