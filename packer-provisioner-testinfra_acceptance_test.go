package main

import (
  _ "embed"
  "fmt"
  "os"
  "os/exec"
  "io/ioutil"
  "regexp"
  "testing"

  "github.com/hashicorp/packer-plugin-sdk/acctest"
)

//go:embed fixtures/test.pkr.hcl
var testTemplate string

// testinfra basic acceptance testing function
func TestTestinfraProvisioner(test *testing.T) {
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
        return fmt.Errorf("Bad exit code from Packer build. Logfile: %s", logfile)
      }

      // manage logfile
      logs, err := os.Open(logfile)
      if err != nil {
        return fmt.Errorf("Unable to find logfile at: %s", logfile)
      }
      defer logs.Close()

      // manage logfile content
      logsBytes, err := ioutil.ReadAll(logs)
      if err != nil {
        return fmt.Errorf("Unable to read logs at: %s", logfile)
      }
      logsString := string(logsBytes)

      // verify logfile content
      if docker_matched, _ := regexp.MatchString("docker.ubuntu: Testing machine image with Testinfra.*", logsString); !docker_matched {
        test.Fatalf("Logs do not contain expected testinfra value %q", logsString)
      }
      if vbox_matched, _ := regexp.MatchString("virtualbox-vm.ubuntu: Testing machine image with Testinfra.*", logsString); !vbox_matched {
        test.Fatalf("Logs doesn't contain expected testinfra value %q", logsString)
      }
      if tests_matched, _ := regexp.MatchString("2 passed in.*", logsString); !tests_matched {
        test.Fatalf("Logs doesn't contain expected testinfra value %q", logsString)
      }

      return nil
    },
  }

  // invoke acceptance test function
  acctest.TestPlugin(test, testCase)
}
