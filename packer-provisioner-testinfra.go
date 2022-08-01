//go:generate packer-sdc mapstructure-to-hcl2 -type TestinfraConfig
package main

import (
  "os"
  "os/exec"
  "io"
  "fmt"
  "log"
  "context"
  "errors"

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
  config        TestinfraConfig
  generatedData map[string]interface{}
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

  // prepare generated data and context
  provisioner.generatedData = generatedData
  provisioner.config.ctx.Data = generatedData

  // parse generated data for required values
  connectionType := provisioner.generatedData["ConnType"].(string)
  ipaddress := provisioner.generatedData["Host"].(string)
  user := provisioner.generatedData["User"].(string)
  port := provisioner.generatedData["Port"].(int64)
  instanceId := provisioner.generatedData["ID"].(string)

  // determine communication string by packer connection type
  communication := ""
  if connectionType == "ssh" {
    communication = fmt.Sprintf("--hosts=%s@%s:%d", user, ipaddress, port)
  }
  if connectionType == "docker" {
    communication = fmt.Sprintf("--hosts=docker://%s", instanceId)
  }

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
  cmd := exec.Command(pytestPath, "-v", communication, testFile)
  log.Printf("Complete Testinfra command is: %s", cmd.String())
  cmd.Env = os.Environ()

  // prepare stdout and stderr pipes
  stdout, err := cmd.StdoutPipe()
  if err != nil {
    log.Fatal(err)
    return err
  }
  stderr, err := cmd.StderrPipe()
  if err != nil {
    log.Fatal(err)
    return err
  }

  // initialize testinfra tests
  if err := cmd.Start(); err != nil {
    log.Fatalf("Initialization of Testinfra py.test command execution returned non-zero exit status: %s", err)
    return err
  }

  // display testinfra results
  outSlurp, err := io.ReadAll(stdout)
  if err != nil {
    log.Fatalf("Unable to read stdout from Testinfra: %s", err)
    return err
  }
  if len(outSlurp) > 0 {
    ui.Message(string(outSlurp))
  }

  errSlurp, err := io.ReadAll(stderr)
  if err != nil {
    log.Fatalf("Unable to read stderr from Testinfra: %s", err)
    return err
  }
  if len(errSlurp) > 0 {
    ui.Error(string(errSlurp))
  }

  // wait for testinfra to complete
  err = cmd.Wait()
  if err != nil {
    log.Fatalf("Testinfra returned non-zero exit status: %s", err)
    return err
  }

  // finish
  ui.Say("Testinfra machine image testing is complete")

  return nil
}
