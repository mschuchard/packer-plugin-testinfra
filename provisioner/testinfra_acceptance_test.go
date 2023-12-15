package testinfra

import (
	_ "embed"
	"io/ioutil"
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
		Name:     "testinfra_provisioner_test",
		Init:     true,
		Setup:    func() error { return nil },
		Template: testTemplate,
		Type:     "provisioner",
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			// verify good exit code from packer process
			if buildCommand.ProcessState != nil && buildCommand.ProcessState.ExitCode() != 0 {
				test.Errorf("unexpected exit code from Packer build; logfile: %s", logfile)
			}

			// manage logfile
			logs, err := os.Open(logfile)
			if err != nil {
				test.Errorf("unable to open logfile at: %s", logfile)
			}
			defer logs.Close()

			// manage logfile content
			logsBytes, err := ioutil.ReadAll(logs)
			if err != nil {
				test.Errorf("unable to read logfile at: %s", logfile)
			}
			// convert log byte slice to string
			logsString := string(logsBytes)

			// verify logfile content
			if dockerMatches, _ := regexp.MatchString("docker.ubuntu: testing machine image with Testinfra.*", logsString); !dockerMatches {
				test.Errorf("logs do not contain expected docker testinfra value: %s", logsString)
			}
			//TODO: https://github.com/hashicorp/packer-plugin-virtualbox/issues/77
			/*if vbox_matched, _ := regexp.MatchString("virtualbox-vm.ubuntu: Testing machine image with Testinfra.*", logsString); !vbox_matched {
			  test.Fatalf("Logs do not contain expected virtualbox testinfra value %q", logsString)
			}*/
			if testsMatches, _ := regexp.MatchString("2 passed in.*", logsString); !testsMatches {
				test.Errorf("logs do not contain expected testinfra value: %s", logsString)
			}

			return nil
		},
	}

	// invoke acceptance test function
	acctest.TestPlugin(test, testCase)
}
