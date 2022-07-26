//go:generate packer-sdc mapstructure-to-hcl2 -type TestinfraConfig
package main

import (
  "os"
  "os/exec"
  "fmt"
  "log"
  "errors"
  "context"

  "github.com/hashicorp/hcl/v2/hcldec"
  "github.com/hashicorp/packer-plugin-sdk/packer"
  "github.com/hashicorp/packer-plugin-sdk/template/config"
  "github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

// config data from packer template/config
type TestinfraConfig struct {
  PytestPath string `mapstructure:"pytest_path"`
  TestFile   string `mapstructure:"test_file"`

  ctx interpolate.Context
}

// implements the packer.Provisioner interface
type TestinfraProvisioner struct{
  config TestinfraConfig
}

// implements configspec with hcl2spec helper function
func (provisioner *TestinfraProvisioner) ConfigSpec() hcldec.ObjectSpec {
  return provisioner.config.FlatMapstructure().HCL2Spec()
}

// prepares the provisioner plugin
func (provisioner *TestinfraProvisioner) Prepare(raws ...interface{}) error {
  // parse testinfra provisioner config
  err := config.Decode(&provisioner.config, &config.DecodeOpts{
    Interpolate:        true,
    InterpolateContext: &provisioner.config.ctx,
  }, raws...)
  if err != nil {
    log.Fatal("Error decoding the supplied Packer config.")
    return err
  }

  // set default executable path for py.test
  if provisioner.config.PytestPath == "" {
    provisioner.config.PytestPath = "py.test"
  } else { // verify py.test exists at supplied path
    if _, err := os.Stat(provisioner.config.PytestPath); errors.Is(err, os.ErrNotExist) {
      log.Fatalf("The Pytest executable does not exist at: %s", provisioner.config.PytestPath)
      return err
    }
  }

  // verify testinfra file exists
  if _, err := os.Stat(provisioner.config.TestFile); errors.Is(err, os.ErrNotExist) {
    log.Fatalf("The Testinfra file does not exist at: %s", provisioner.config.TestFile)
    return err
  }

  return nil
}

// executes the provisioner plugin
func (provisioner *TestinfraProvisioner) Provision(ctx context.Context, ui packer.Ui, comm packer.Communicator, generatedData map[string]interface{}) error {
  ui.Say("Testing machine image with Testinfra")

  // parse generated data for context and required values
  provisioner.config.ctx.Data = generatedData
  hostname := generatedData["Host"].(string)
  user := generatedData["User"].(string)
  port := generatedData["Port"].(string)
  communication := fmt.Sprintf("--hosts=%s@%s:%s", user, hostname, port)

  // pyest path
  pytestPath, err := interpolate.Render(provisioner.config.PytestPath, &provisioner.config.ctx)
  if err != nil {
    log.Fatalf("Error parsing config for PytestPath: %v", err.Error())
    return err
  }
  // testfile
  testFile, err := interpolate.Render(provisioner.config.TestFile, &provisioner.config.ctx)
  if err != nil {
    log.Fatalf("Error parsing config for TestFile: %v", err.Error())
    return err
  }

  // prepare testinfra test command
  log.Printf("Complete command is: %s -v %s %s", pytestPath, communication, testFile)
  cmd := exec.Command(pytestPath, "-v", communication, testFile)
  cmd.Env = os.Environ()

  // execute testinfra tests
  if err := cmd.Start(); err != nil {
    return err
  }
  err = cmd.Wait()
  if err != nil {
    log.Fatalf("Non-zero exit status: %s", err)
    return err
  }

  // finish
  ui.Say("Testinfra machine image testing is complete")

  return nil
}
