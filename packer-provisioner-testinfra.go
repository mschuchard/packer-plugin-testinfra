//go:generate packer-sdc mapstructure-to-hcl2 -type TestinfraConfig
package main

import (
  "os"
  "os/exec"
  "strconv"
  "strings"
  "bytes"
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

// implements the packer.Provisioner interface
type TestinfraProvisioner struct{
  config        TestinfraConfig
  generatedData map[string]interface{}
}

// ssh auth type with pseudo-enum
type SSHAuth string

const (
  passwordSSHAuth   SSHAuth = "password"
  agentSSHAuth      SSHAuth = "agent"
  privateKeySSHAuth SSHAuth = "privateKey"
)

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

  // log optional arguments
  if len(provisioner.config.Keyword) > 0 {
    log.Printf("Executing tests with keyword substring expression: %s", provisioner.config.Keyword)
  }

  if provisioner.config.Local {
    log.Print("Test execution will occur on the temporary Packer instance used for building the machine image artifact")

    if len(provisioner.config.InstallCmd) > 0 {
      log.Printf("Installation command on the temporary Packer instance prior to Testinfra test execution is: %s", strings.Join(provisioner.config.InstallCmd, " "))
    }
  }

  if len(provisioner.config.Marker) > 0 {
    log.Printf("Executing tests with marker expression: %s", provisioner.config.Marker)
  }

  log.Printf("Number of Testinfra processes: %d.", provisioner.config.Processes)

  if provisioner.config.Sudo {
    log.Print("Testinfra will execute with sudo.")
  } else {
    log.Print("Testinfra will not execute with sudo.")
  }

  // set default executable path for py.test
  if len(provisioner.config.PytestPath) == 0 {
    log.Print("Setting PytestPath to default 'py.test'")
    provisioner.config.PytestPath = "py.test"
  } else { // verify py.test exists at supplied path
    if _, err := os.Stat(provisioner.config.PytestPath); errors.Is(err, os.ErrNotExist) {
      log.Printf("The Pytest executable does not exist at: %s", provisioner.config.PytestPath)
      return err
    }
  }

  // verify testinfra files are specified as inputs
  if len(provisioner.config.TestFiles) == 0 {
    log.Print("All files prefixed with 'test_' recursively discovered from the current working directory will be considered Testinfra test files")
  } else { // verify testinfra files exist
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

  // prepare testinfra test command
  cmd, localCmd, err := provisioner.determineExecCmd()
  if err != nil {
    return err
  }

  // execute testinfra remotely with *exec.Cmd
  if len(localCmd.Command) == 0 && cmd != nil {
    err = execCmdTestinfra(cmd, ui)
  } else if len(localCmd.Command) > 0 && cmd == nil {
    // execute testinfra local to instance with packer.RemoteCmd
    err = packerRemoteCmdTestinfra(localCmd, provisioner.config.InstallCmd, comm, ui)
  } else {
    // somehow we either returned both commands or neither or something really weird for one or both
    return fmt.Errorf("Incorrectly determined remote command (%s) and/or command local to instance (%s). Please report as bug with this log information.", cmd.String(), localCmd.Command)
  }
  if err != nil {
    return err
  }

  return nil
}

// execute testinfra remotely with *exec.Cmd
func execCmdTestinfra(cmd *exec.Cmd, ui packer.Ui) error {
  // merge in env settings
  cmd.Env = os.Environ()
  log.Printf("Complete Testinfra remote command is: %s", cmd.String())

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
    ui.Say("Testinfra results include the following:")
    ui.Message(string(outSlurp))
  }

  errSlurp, err := io.ReadAll(stderr)
  if err != nil {
    log.Printf("Unable to read stderr from Testinfra: %s", err)
    return err
  }
  if len(errSlurp) > 0 {
    ui.Error("Testinfra errored internally during execution:")
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

// execute testinfra local to temp packer instance with packer.RemoteCmd
func packerRemoteCmdTestinfra(localCmd *packer.RemoteCmd, installCmd []string, comm packer.Communicator, ui packer.Ui) error {
  // initialize context and log command
  ctx := context.TODO()
  log.Printf("Complete Testinfra local command is: %s", localCmd.Command)

  // install testinfra on temp packer instance
  if len(installCmd) > 0 {
    // cast installCmd to string, log, and init localInstallCmd
    strInstallCmd := strings.Join(installCmd, " ")
    ui.Say("Installing Testinfra on instance")
    log.Printf("Testinfra installation command is: %s", strInstallCmd)
    localInstallCmd := &packer.RemoteCmd{Command: strInstallCmd}

    // install testinfra on temp packer instance
    if err := comm.Start(ctx, localInstallCmd); err != nil {
      log.Printf("Testinfra install command execution returned non-zero exit status: %s", err)
      return err
    }
  }

  // initialize stdout and stderr
  var stdout bytes.Buffer
  localCmd.Stdout = &stdout
  var stderr bytes.Buffer
  localCmd.Stderr = &stderr

  // initialize testinfra tests
  ui.Say("Beginning Testinfra validation of machine image")
  if err := comm.Start(ctx, localCmd); err != nil {
    log.Printf("Initialization of Testinfra py.test command execution returned non-zero exit status: %s", err)
    return err
  }

  // wait for testinfra to complete and flush buffers
  // then check for pytest/testinfra execution issues
  if exitStatus := localCmd.Wait(); exitStatus > 0 || len(stderr.String()) > 0 {
    ui.Error("Testinfra errored internally during execution:")
    ui.Error(stderr.String())
    return fmt.Errorf("Testinfra returned exit status: %d", exitStatus)
  }

  // capture and display testinfra output
  if len(stdout.String()) > 0 {
    ui.Say("Testinfra results include the following:")
    ui.Message(stdout.String())
  }

  // finish and return
  ui.Say("Testinfra machine image testing is complete")

  return nil
}

// determine and return execution command for testinfra
func (provisioner *TestinfraProvisioner) determineExecCmd() (*exec.Cmd, *packer.RemoteCmd, error) {
  // initialize args with base argument
  args := []string{"-v"}

  // assign determined communication string
  localExec := provisioner.config.Local
  if localExec == false {
    communication, err := provisioner.determineCommunication()
    if err != nil {
      return nil, &packer.RemoteCmd{}, err
    }
    args = append(args, communication)
  }

  // assign mandatory populated values
  // pytest path
  pytestPath, err := interpolate.Render(provisioner.config.PytestPath, &provisioner.config.ctx)
  if err != nil {
    log.Printf("Error parsing config for PytestPath: %v", err.Error())
    return nil, &packer.RemoteCmd{}, err
  }

  // assign optional populated values
  // testfiles
  args = append(args, provisioner.config.TestFiles...)
  // keyword
  keyword, err := interpolate.Render(provisioner.config.Keyword, &provisioner.config.ctx)
  if err != nil {
    log.Printf("Error parsing config for Keyword: %v", err.Error())
    return nil, &packer.RemoteCmd{}, err
  }
  if len(keyword) > 0 {
    args = append(args, "-k", fmt.Sprintf("\"%s\"", keyword))
  }
  // marker
  marker, err := interpolate.Render(provisioner.config.Marker, &provisioner.config.ctx)
  if err != nil {
    log.Printf("Error parsing config for Marker: %v", err.Error())
    return nil, &packer.RemoteCmd{}, err
  }
  if len(marker) > 0 {
    args = append(args, "-m", fmt.Sprintf("\"%s\"", marker))
  }
  // processes
  if provisioner.config.Processes != 0 {
    args = append(args, "-n", strconv.Itoa(provisioner.config.Processes))
  }
  // sudo
  if provisioner.config.Sudo == true {
    args = append(args, "--sudo")
  }

  // return packer remote command for local testing on instance
  if localExec == true {
    return nil, &packer.RemoteCmd{Command: fmt.Sprintf("%s %s", pytestPath, strings.Join(args, " "))}, nil
  } else { // return exec command for remote testing against instance
    return exec.Command(pytestPath, args...), &packer.RemoteCmd{Command: ""}, nil
  }
}

// determine and return appropriate communication string for pytest/testinfra
func (provisioner *TestinfraProvisioner) determineCommunication() (string, error) {
  // parse generated data for required values
  connectionType := provisioner.generatedData["ConnType"].(string)
  user, ok := provisioner.generatedData["SSHUsername"].(string)
  if !ok {
    user = provisioner.generatedData["User"].(string)
  }
  ipaddress := provisioner.generatedData["Host"].(string)
  port, ok := provisioner.generatedData["SSHPort"].(int64)
  if !ok {
    port = provisioner.generatedData["Port"].(int64)
  }
  httpAddr := fmt.Sprintf("%s:%d", ipaddress, port)
  if len(ipaddress) == 0 {
    httpAddr = provisioner.generatedData["PackerHTTPAddr"].(string)
  }
  instanceID := provisioner.generatedData["ID"].(string)

  // parse generated data for optional values
  //uuid := provisioner.generatedData["PackerRunUUID"].(string)

  // determine communication string by packer connection type
  log.Printf("Testinfra communicating via %s connection type", connectionType)
  communication := ""

  switch connectionType {
  case "ssh":
    // assign ssh auth type and string (key file path or password)
    sshAuthType, sshAuthString, err := provisioner.determineSSHAuth()
    if err != nil {
      return "", err
    }

    // initialize whitespace string as default arg (implied for sshAgentAuth type)
    sshIdentity := " "
    // use ssh private key file
    if sshAuthType == privateKeySSHAuth {
      log.Printf("SSH private key filesystem location is: %s", sshAuthString)
      sshIdentity = fmt.Sprintf(" --ssh-identity-file=%s ", sshAuthString)
    } else if sshAuthType == passwordSSHAuth { // use ssh password
      log.Printf("Utilizing SSH password for communicator authentication.")
      // modify user string to also include password
      user = fmt.Sprintf("%s:%s", user, sshAuthString)
    }

    communication = fmt.Sprintf("--hosts=%s@%s%s--ssh-extra-args=\"-o StrictHostKeyChecking=no\"", user, httpAddr, sshIdentity)
  case "winrm":
    // assign winrm password preferably from winrmpassword
    winRMPassword, ok := provisioner.generatedData["WinRMPassword"].(string)
    // otherwise retry with general password
    if !ok {
      winRMPassword, ok = provisioner.generatedData["Password"].(string)
    }
    // no winrm password available
    if !ok {
      return "", fmt.Errorf("WinRM communicator password could not be determined from available Packer data.")
    }

    communication = fmt.Sprintf("--hosts=winrm://%s:%s@%s", user, winRMPassword, httpAddr)
  case "docker":
    communication = fmt.Sprintf("--hosts=docker://%s", instanceID)
  case "podman":
    communication = fmt.Sprintf("--hosts=podman://%s", instanceID)
  case "lxc":
    communication = fmt.Sprintf("--hosts=lxc://%s", instanceID)
  }
  if len(communication) == 0 {
    return "", fmt.Errorf("Communication with machine image could not be properly determined")
  }

  return communication, nil
}

// determine and return ssh authentication
func (provisioner *TestinfraProvisioner) determineSSHAuth() (SSHAuth, string, error) {
  // assign ssh password preferably from sshpassword
  sshPassword, ok := provisioner.generatedData["SSHPassword"].(string)

  // otherwise retry with general password
  if !ok {
    sshPassword, ok = provisioner.generatedData["Password"].(string)
  }

  // ssh is being used with password auth and we have a password
  if ok {
    return passwordSSHAuth, sshPassword, nil
  } else { // ssh is being used with private key or agent auth so determine that instead
    // parse generated data for ssh private key and agent auth info
    sshPrivateKeyFile := provisioner.generatedData["SSHPrivateKeyFile"].(string)
    sshAgentAuth := provisioner.generatedData["SSHAgentAuth"].(bool)

    if len(sshPrivateKeyFile) > 0 {
      // we have a legitimate private key file so use that
      return privateKeySSHAuth, sshPrivateKeyFile, nil
    } else if sshAgentAuth {
      // we can use an empty private key with ssh agent auth
      return agentSSHAuth, sshPrivateKeyFile, nil
    } else { // create a private key file instead
      // write a tmpfile for storing a private key
      tmpSSHPrivateKey, err := tmp.File("testinfra-key")
      if err != nil {
        return "", "", fmt.Errorf("Error creating a temp file for the ssh private key: %v", err)
      }

      // attempt to obtain a private key
      SSHPrivateKey := provisioner.generatedData["SSHPrivateKey"].(string)

      // write the private key to the tmpfile
      _, err = tmpSSHPrivateKey.WriteString(SSHPrivateKey)
      if err != nil {
        return "", "", fmt.Errorf("Failed to write ssh private key to temp file")
      }

      // and then close the tmpfile storing the private key
      err = tmpSSHPrivateKey.Close()
      if err != nil {
        return "", "", fmt.Errorf("Failed to close ssh private key temp file")
      }

      return privateKeySSHAuth, tmpSSHPrivateKey.Name(), nil
    }
  }
}
