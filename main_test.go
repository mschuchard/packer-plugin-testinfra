package main

import (
	_ "embed"
	"log"
	"os"
	"os/exec"
	"regexp"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
)

//go:embed fixtures/test.pkr.hcl
var testTemplate string

// testinfra basic acceptance testing function
func TestProvisioner(test *testing.T) {
	// initialize acceptance test config struct
	testCase := &acctest.PluginTestCase{
		Name: "testinfra_plugin_test",
		Init: false,
		Setup: func() error {
			// inform vbox machine should be running
			log.Print("INFO: ensure the virtualbox virtual machine is running for ssh communicator testing")
			// validate password env var exists
			if password := os.Getenv("PACKER_VAR_password"); len(password) == 0 {
				test.Fatal("environment variable 'PACKER_VAR_password' must be set for acceptance testing")
			}
			return nil
		},
		Template: testTemplate,
		Type:     "provisioner",
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			// verify good exit code from packer process
			if buildCommand.ProcessState != nil && buildCommand.ProcessState.ExitCode() != 0 {
				test.Errorf("unexpected exit code from 'packer build'; logfile: %s", logfile)
				return nil
			}

			// assign logfile from content
			logsBytes, err := os.ReadFile(logfile)
			if err != nil {
				test.Errorf("unable to read logfile at: %s", logfile)
				return err
			}
			// convert log byte slice to string
			logsString := string(logsBytes)

			// verify logfile content for each communicator
			if dockerMatches, _ := regexp.MatchString("docker.ubuntu: packer plugin testinfra provisioning complete.*", logsString); !dockerMatches {
				test.Errorf("logs do not contain expected docker testinfra completion log in logfile: %s", logfile)
			}
			if nullMatches, _ := regexp.MatchString("null.vbox: packer plugin testinfra provisioning complete.*", logsString); !nullMatches {
				test.Errorf("logs do not contain expected ssh testinfra completion log in logfile: %s", logfile)
			}
			//TODO: https://github.com/hashicorp/packer-plugin-virtualbox/issues/77
			/*if vboxMatched, _ := regexp.MatchString("virtualbox-vm.ubuntu: packer plugin testinfra provisioning complete.*", logsString); !vboxMatched {
			  test.Fatalf("logs do not contain expected local testinfra values in logfile: %s", logfile)
			}*/
			// verify testinfra output is as expected
			if testsMatches, _ := regexp.MatchString("2 passed in.*", logsString); !testsMatches {
				test.Errorf("logs do not contain expected testinfra test results: %s", logsString)
			}

			return nil
		},
	}

	// invoke acceptance test function
	acctest.TestPlugin(test, testCase)
}
