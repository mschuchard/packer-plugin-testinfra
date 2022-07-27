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

//go:embed fixtures/docker.pkr.hcl
var testDockerTemplate string

// testinfra basic acceptance testing function
func TestTestinfraProvisioner(test *testing.T) {
  // initialize acceptance test config struct
  testCase := &acctest.PluginTestCase{
		Name:     "testinfra_provisioner_docker_test",
    Init:     true,
    Setup:    func() error { return nil },
    Template: testDockerTemplate,
    Type:     "provisioner",
    Check: func(buildCommand *exec.Cmd, logfile string) error {
      // verify good exit code from packer process
      if buildCommand.ProcessState != nil && buildCommand.ProcessState.ExitCode() != 0 {
        return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
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
        return fmt.Errorf("Unable to read logs in: %s", logfile)
      }
      logsString := string(logsBytes)

      // verify logfile content
      if matched, _ := regexp.MatchString("docker.ubuntu:.*", logsString); !matched {
        test.Fatalf("logs doesn't contain expected testinfra value %q", logsString)
      }

      return nil
    },
  }

  // invoke acceptance test function
  acctest.TestPlugin(test, testCase)
}
