//go:generate packer-sdc mapstructure-to-hcl2 -type Config
package testinfra

import (
  "os"
  "os/exec"
  "io"
  "strings"
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

// implements the packer.Provisioner interface
type Provisioner struct{
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
    Interpolate:        true,
    InterpolateContext: &provisioner.config.ctx,
  }, raws...)
  if err != nil {
    log.Print("Error decoding the supplied Packer config.")
    return err
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

  // log optional arguments
  if len(provisioner.config.Keyword) > 0 {
    log.Printf("Executing tests with keyword substring expression: %s", provisioner.config.Keyword)
  }

  if provisioner.config.Local {
    log.Print("Test execution will occur on the temporary Packer instance used for building the machine image artifact")

    if len(provisioner.config.InstallCmd) > 0 {
      log.Printf("Installation command on the temporary Packer instance prior to Testinfra test execution is: %s", strings.Join(provisioner.config.InstallCmd, " "))
    }
  } else { // verify testinfra installed
    // initialize testinfra --help command
    cmd := exec.Command(provisioner.config.PytestPath, []string{"--help"}...)

    // prepare stdout pipe
    stdout, err := cmd.StdoutPipe()
    if err != nil {
      log.Print(err)
      return err
    }

    // initialize testinfra installed check
    if err := cmd.Start(); err != nil {
      log.Printf("Initialization of Testinfra 'py.test --help' command execution returned non-zero exit status: %s", err)
      return err
    }

    // capture pytest stdout
    outSlurp, err := io.ReadAll(stdout)
    if err != nil {
      log.Printf("Unable to read stdout from Pytest: %s", err)
      return err
    }

    // examine pytest stdout
    if len(outSlurp) > 0 {
      // check for testinfra in stdout
    } else {
      // pytest returned no stdout
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
func (provisioner *Provisioner) Provision(ctx context.Context, ui packer.Ui, comm packer.Communicator, generatedData map[string]interface{}) error {
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
    err = execCmd(cmd, ui)
  } else if len(localCmd.Command) > 0 && cmd == nil {
    // execute testinfra local to instance with packer.RemoteCmd
    err = packerRemoteCmd(localCmd, provisioner.config.InstallCmd, comm, ui)
  } else {
    // somehow we either returned both commands or neither or something really weird for one or both
    return fmt.Errorf("Incorrectly determined remote command (%s) and/or command local to instance (%s). Please report as bug with this log information.", cmd.String(), localCmd.Command)
  }
  if err != nil {
    return err
  }

  return nil
}