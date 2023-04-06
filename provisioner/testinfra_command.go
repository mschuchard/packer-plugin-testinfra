package testinfra

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

  "github.com/hashicorp/packer-plugin-sdk/packer"
  "github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

// execute testinfra remotely with *exec.Cmd
func execCmd(cmd *exec.Cmd, ui packer.Ui) error {
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
  } else {
    ui.Say("Testinfra produced no stdout. It is likely something unintended occurred during execution.")
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
func packerRemoteCmd(localCmd *packer.RemoteCmd, installCmd []string, comm packer.Communicator, ui packer.Ui) error {
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
  } else {
    ui.Say("Testinfra produced no stdout. It is likely something unintended occurred during execution.")
  }

  // finish and return
  ui.Say("Testinfra machine image testing is complete")

  return nil
}

// determine and return execution command for testinfra
func (provisioner *Provisioner) determineExecCmd() (*exec.Cmd, *packer.RemoteCmd, error) {
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
