//go:generate packer-sdc mapstructure-to-hcl2 -type Config
package testinfra

import (
	"context"
	"fmt"
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

// config data unmarshalled from packer template/config
type Config struct {
	InstallCmd []string `mapstructure:"install_cmd"`
	Keyword    string   `mapstructure:"keyword"`
	Local      bool     `mapstructure:"local"`
	Marker     string   `mapstructure:"marker"`
	Processes  int      `mapstructure:"processes"`
	PytestPath string   `mapstructure:"pytest_path"`
	Sudo       bool     `mapstructure:"sudo"`
	TestFiles  []string `mapstructure:"test_files"`

	ctx interpolate.Context
}

// implements the packer.Provisioner interface as testinfra.Provisioner
type Provisioner struct {
	config        Config
	generatedData map[string]interface{}
}

// implements configspec with hcl2spec helper function
func (provisioner *Provisioner) ConfigSpec() hcldec.ObjectSpec {
	return provisioner.config.FlatMapstructure().HCL2Spec()
}

// prepares the provisioner plugin
func (provisioner *Provisioner) Prepare(raws ...interface{}) error {
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
	} else { // verify py.test exists at supplied path
		if _, err := os.Stat(provisioner.config.PytestPath); err != nil {
			log.Printf("the Pytest executable does not exist at: %s", provisioner.config.PytestPath)
			return err
		}
	}

	// log optional arguments
	if len(provisioner.config.Keyword) > 0 {
		log.Printf("executing tests with keyword substring expression: %s", provisioner.config.Keyword)
	}

	if provisioner.config.Local {
		// no validation of testinfra installation
		log.Print("test execution will occur on the temporary Packer instance used for building the machine image artifact")

		if len(provisioner.config.InstallCmd) > 0 {
			log.Printf("installation command on the temporary Packer instance prior to Testinfra test execution is: %s", strings.Join(provisioner.config.InstallCmd, " "))
		}

		// no validation of xdist installation
		if provisioner.config.Processes > 0 {
			log.Printf("number of Testinfra processes: %d", provisioner.config.Processes)
		}
	} else { // verify testinfra installed
		log.Print("beginning Testinfra installation verification")
		// initialize testinfra -h command
		cmd := exec.Command(provisioner.config.PytestPath, []string{"-h"}...)

		// prepare stdout pipe
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Print("unable to prepare the pipe for capturing stdout")
			return err
		}

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
			// check for testinfra in stdout
			if strings.Contains(string(outSlurp), "testinfra") {
				log.Print("testinfra installation existence verified")
			} else {
				return fmt.Errorf("testinfra installation not found by specified Pytest installation")
			}

			if provisioner.config.Processes > 0 {
				// check for xdist in pytest usage stdout
				if strings.Contains(string(outSlurp), " -n ") {
					log.Printf("number of Testinfra processes: %d", provisioner.config.Processes)
				} else {
					log.Printf("pytest-xdist is not installed, and processes parameter will be reset to default")
					provisioner.config.Processes = 0
				}
			}
		} else {
			// pytest returned no stdout
			return fmt.Errorf("pytest help command returned no stdout; this indicates an issue with the specified Pytest installation")
		}
	}

	// marker parameter
	if len(provisioner.config.Marker) > 0 {
		log.Printf("executing tests with marker expression: %s", provisioner.config.Marker)
	}

	// sudo parameter
	if provisioner.config.Sudo {
		log.Print("testinfra will execute with sudo")
	} else {
		log.Print("testinfra will not execute with sudo")
	}

	// verify testinfra files are specified as inputs
	if len(provisioner.config.TestFiles) == 0 {
		log.Print("all files prefixed with 'test_' recursively discovered from the current working directory will be considered Testinfra test files")
	} else { // verify testinfra files exist
		for _, testFile := range provisioner.config.TestFiles {
			if _, err := os.Stat(testFile); err != nil {
				log.Printf("the Testinfra test_file does not exist at: %s", testFile)
				return err
			}
		}
	}

	return nil
}

// executes the provisioner plugin
func (provisioner *Provisioner) Provision(ctx context.Context, ui packer.Ui, comm packer.Communicator, generatedData map[string]interface{}) error {
	ui.Say("testing machine image with Testinfra")

	// prepare generated data and context
	provisioner.generatedData = generatedData
	provisioner.config.ctx.Data = generatedData

	// prepare testinfra test command
	cmd, localCmd, err := provisioner.determineExecCmd()
	if err != nil {
		ui.Error("the execution command could not be accurately determined")
		return err
	}

	// execute testinfra remotely with *exec.Cmd
	if localCmd == nil && cmd != nil {
		err = execCmd(cmd, ui)
	} else if localCmd != nil && cmd == nil {
		// execute testinfra local to instance with packer.RemoteCmd
		err = packerRemoteCmd(localCmd, provisioner.config.InstallCmd, comm, ui)
	} else {
		// somehow we either returned both commands or neither or something really weird for one or both
		return fmt.Errorf("incorrectly determined remote command (%s) and/or command local to instance (%s); please report as bug with this log information", cmd.String(), localCmd.Command)
	}
	if err != nil {
		ui.Error("the Pytest Testinfra execution failed")
		return err
	}

	return nil
}
