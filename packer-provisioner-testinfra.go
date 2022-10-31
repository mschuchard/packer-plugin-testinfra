//go:generate packer-sdc mapstructure-to-hcl2 -type TestinfraConfig
package main

import (
  "os"
  "os/exec"
  "strings"
  "io"
  "fmt"
  "log"
  "context"
  "errors"

  "github.com/hashicorp/hcl/v2/hcldec"
  "github.com/hashicorp/packer-plugin-sdk/packer"
  "github.com/hashicorp/packer-plugin-sdk/template/config"
  "github.com/hashicorp/packer-plugin-sdk/template/interpolate"
  "github.com/hashicorp/packer-plugin-sdk/tmp"
)

// config data from packer template/config
type TestinfraConfig struct {
  Processes  int      `mapstructure:"processes"`
  PytestPath string   `mapstructure:"pytest_path"`
  TestFiles  []string `mapstructure:"test_files"`

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
    log.Print("Error decoding the supplied Packer config.")
    return err
  }

  // set default number of parallel processes
  if provisioner.config.Processes == 0 {
    log.Print("Not overriding Testinfra default number of parallel processes.")
  }

  // set default executable path for py.test
  if provisioner.config.PytestPath == "" {
    log.Print("Setting PytestPath to default 'py.test'")
    provisioner.config.PytestPath = "py.test"
  } else { // verify py.test exists at supplied path
    if _, err := os.Stat(provisioner.config.PytestPath); errors.Is(err, os.ErrNotExist) {
      log.Printf("The Pytest executable does not exist at: %s", provisioner.config.PytestPath)
      return err
    }
  }

  // verify testinfra file exists
  if len(provisioner.config.TestFiles) == 0 {
    return fmt.Errorf("The Testinfra test_files were not specified")
  } else {
    for _, testFile := range provisioner.config.TestFiles {
      if _, err := os.Stat(testFile); errors.Is(err, os.ErrNotExist) {
        log.Printf("The Testinfra test_file does not exist at: %s", testFile)
        return err
      }
    }
  }

  return nil
}

// executes the provisioner plugin
func (provisioner *TestinfraProvisioner) Provision(ctx context.Context, ui packer.Ui, comm packer.Communicator, generatedData map[string]interface{}) error {
  ui.Say("Testing machine image with Testinfra")

  // prepare generated data and context
  provisioner.generatedData = generatedData
  provisioner.config.ctx.Data = generatedData

  // assign communication string
  communication, err := provisioner.determineCommunication()
  if err != nil {
    return err
  }

  // pytest path
  pytestPath, err := interpolate.Render(provisioner.config.PytestPath, &provisioner.config.ctx)
  if err != nil {
    log.Printf("Error parsing config for PytestPath: %v", err.Error())
    return err
  }
  // testfile
  testFiles := provisioner.config.TestFiles
  if err != nil {
    log.Printf("Error parsing config for TestFiles: %v", err.Error())
    return err
  }

  // prepare testinfra test command
  cmd := exec.Command(pytestPath, "-v", communication, strings.Join(testFiles, " "))
  log.Printf("Complete Testinfra command is: %s", cmd.String())
  cmd.Env = os.Environ()

  // prepare stdout and stderr pipes
  stdout, err := cmd.StdoutPipe()
  if err != nil {
    log.Print(err)
    return err
  }
  stderr, err := cmd.StderrPipe()
  if err != nil {
    log.Print(err)
    return err
  }

  // initialize testinfra tests
  ui.Say("Beginning Testinfra validation of machine image")
  if err := cmd.Start(); err != nil {
    log.Printf("Initialization of Testinfra py.test command execution returned non-zero exit status: %s", err)
    return err
  }

  // capture and display testinfra output
  outSlurp, err := io.ReadAll(stdout)
  if err != nil {
    log.Printf("Unable to read stdout from Testinfra: %s", err)
    return err
  }
  if len(outSlurp) > 0 {
    ui.Message(string(outSlurp))
  }

  errSlurp, err := io.ReadAll(stderr)
  if err != nil {
    log.Printf("Unable to read stderr from Testinfra: %s", err)
    return err
  }
  if len(errSlurp) > 0 {
    ui.Error(string(errSlurp))
  }

  // wait for testinfra to complete and flush buffers
  err = cmd.Wait()
  if err != nil {
    log.Printf("Testinfra returned non-zero exit status: %s", err)
    return err
  }

  // finish and return
  ui.Say("Testinfra machine image testing is complete")

  return nil
}

// determine and return appropriate communication string for pytest/testinfra
func (provisioner *TestinfraProvisioner) determineCommunication() (string, error) {
  // parse generated data for required values
  connectionType := provisioner.generatedData["ConnType"].(string)
  user := provisioner.generatedData["User"].(string)
  ipaddress := provisioner.generatedData["Host"].(string)
  port := provisioner.generatedData["Port"].(int64)
  httpAddr := fmt.Sprintf("%s:%d", ipaddress, port)
  if len(ipaddress) == 0 {
    httpAddr = provisioner.generatedData["PackerHTTPAddr"].(string)
  }
  instanceID := provisioner.generatedData["ID"].(string)

  // parse generated data for optional values
  //uuid := provisioner.generatedData["PackerRunUUID"].(string)
  //password := provisioner.generatedData["Password"].(string)
  //sshPubKey := provisioner.generatedData["SSHPublicKey"].(string)
  //sshPrivKey := provisioner.generatedData["SSHPrivateKey"].(string)
  //winRMPassword := provisioner.generatedData["WinRMPassword"].(string)

  // determine communication string by packer connection type
  log.Printf("Testinfra communicating via %s connection type", connectionType)
  communication := ""

  switch connectionType {
  case "ssh":
    // assign ssh private key file
    sshPrivateKeyFile := provisioner.generatedData["SSHPrivateKeyFile"].(string)
    sshAgentAuth := provisioner.generatedData["SSHAgentAuth"].(bool)
    sshPrivateKeyFile, err := provisioner.determineSSHAuth(sshPrivateKeyFile, sshAgentAuth)
    if err != nil {
      return "", err
    }
    log.Printf("SSH private key filesystem location is: %s", sshPrivateKeyFile)

    communication = fmt.Sprintf("--hosts=%s@%s --ssh-identity-file=%s --ssh-extra-args=\"-o StrictHostKeyChecking=no\"", user, httpAddr, sshPrivateKeyFile)
  case "winrm":
    // assign winrm password
    winRMPassword := provisioner.generatedData["Password"]

    communication = fmt.Sprintf("--hosts=winrm://%s:%s@%s", user, winRMPassword, httpAddr)
  case "docker":
    communication = fmt.Sprintf("--hosts=docker://%s", instanceID)
  }
  if len(communication) == 0 {
    return "", fmt.Errorf("Communication with machine image could not be properly determined")
  }

  return communication, nil
}

// determine ssh authentication
func (provisioner *TestinfraProvisioner) determineSSHAuth(sshPrivateKeyFile string, sshAgentAuth bool) (string, error) {
  if sshPrivateKeyFile != "" || sshAgentAuth {
    // we have a legitimate private key file, so use that
    return sshPrivateKeyFile, nil
  } else {
    // write a tmpfile for storing a private key
    tmpSSHPrivateKey, err := tmp.File("testinfra-key")
    if err != nil {
      return "", fmt.Errorf("Error creating a temp file for the ssh private key: %v", err)
    }
    // attempt to obtain a private key
    SSHPrivateKey := provisioner.generatedData["SSHPrivateKey"].(string)
    // write the private key to the tmpfile
    _, err = tmpSSHPrivateKey.WriteString(SSHPrivateKey)
    if err != nil {
      return "", fmt.Errorf("Failed to write ssh private key to temp file")
    }
    // and then close the tmpfile storing the private key
    err = tmpSSHPrivateKey.Close()
    if err != nil {
      return "", fmt.Errorf("Failed to close ssh private key temp file")
    }
    return tmpSSHPrivateKey.Name(), nil
  }
}
