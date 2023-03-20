package testinfra

import (
  "os"
  "reflect"
  "strings"
  "testing"
  "errors"

  "github.com/hashicorp/packer-plugin-sdk/packer"
)

// global helper vars for tests
var basicConfig = &Config{
  Keyword:    "not slow",
  Marker:     "fast",
  Processes:  4,
  PytestPath: "/usr/local/bin/py.test",
  Sudo:       true,
  TestFiles:  []string{"fixtures/test.py"},
}

var minConfig = &Config{
  TestFiles: []string{"fixtures/test.py"},
}

// test basic config for packer template/config data
func TestProvisionerConfig(test *testing.T) {
  var provisioner = &Provisioner{
    config: *basicConfig,
  }
  configInput := *basicConfig

  if provisioner.config.PytestPath != configInput.PytestPath || !(reflect.DeepEqual(provisioner.config.TestFiles, configInput.TestFiles)) || provisioner.config.Keyword != configInput.Keyword || provisioner.config.Marker != configInput.Marker || provisioner.config.Processes != configInput.Processes || provisioner.config.Sudo != configInput.Sudo {
    test.Errorf("Provisioner config struct not initialized correctly")
  }
}

// test struct for provisioner interface
func TestProvisionerInterface(test *testing.T) {
  var raw interface{} = &Provisioner{}
  if _, ok := raw.(packer.Provisioner); !ok {
    test.Errorf("Testinfra config struct must be a Provisioner")
  }
}

// test provisioner prepare with basic config
func TestProvisionerPrepareBasic(test *testing.T) {
  var provisioner Provisioner

  err := provisioner.Prepare(basicConfig)
  if err != nil {
    test.Errorf("Prepare function failed with basic Testinfra Packer config")
  }
}

// test provisioner prepare with minimal config
func TestProvisionerPrepareMinimal(test *testing.T) {
  var provisioner Provisioner

  err := provisioner.Prepare(minConfig)
  if err != nil {
    test.Errorf("Prepare function failed with minimal Testinfra Packer config")
  }

  if len(provisioner.config.InstallCmd) > 0 {
    test.Errorf("Default empty setting for InstallCmd is incorrect: %s", provisioner.config.InstallCmd)
  }

  if len(provisioner.config.Keyword) > 0 {
    test.Errorf("Default empty setting for Keyword is incorrect: %s", provisioner.config.Keyword)
  }

  if provisioner.config.Local != false {
    test.Errorf("Default false setting for Local is incorrect: %t", provisioner.config.Local)
  }

  if len(provisioner.config.Marker) > 0 {
    test.Errorf("Default empty setting for Marker is incorrect: %s", provisioner.config.Marker)
  }

  if provisioner.config.Processes != 0 {
    test.Errorf("Default empty setting for Processes is incorrect: %d", provisioner.config.Processes)
  }

  if provisioner.config.Sudo != false {
    test.Errorf("Default false setting for Sudo is incorrect: %t", provisioner.config.Sudo)
  }

  if provisioner.config.PytestPath != "py.test" {
    test.Errorf("Default setting for PytestPath is incorrect: %s", provisioner.config.PytestPath)
  }
}

// test provisioner prepare defaults to empty slice for test files
func TestProvisionerPrepareEmptyTestFile(test *testing.T) {
  var provisioner Provisioner
  var emptyTestFileConfig = &Config{
    PytestPath: "/usr/local/bin/py.test",
  }

  err := provisioner.Prepare(emptyTestFileConfig)
  if err != nil {
    test.Errorf("Prepare function failed with no test_files minimal Testinfra Packer config")
  }

  if len(provisioner.config.TestFiles) > 0 {
    test.Errorf("Default setting for TestFiles is incorrect: %s", strings.Join(provisioner.config.TestFiles, ""))
  }
}

// test provisioner prepare errors on nonexistent files
func TestProvisionerPrepareNonExistFiles(test *testing.T) {
  var provisioner Provisioner

  // test no pytest
  var noPytestConfig = &Config{
    PytestPath: "/home/foo/py.test",
    TestFiles:  []string{"fixtures/test.py"},
  }

  err := provisioner.Prepare(noPytestConfig)
  if !(errors.Is(err, os.ErrNotExist)) {
    test.Errorf("Prepare function did not fail correctly on nonexistent pytest")
  }

  // test nonexistent testfile
  var noTestFileConfig = &Config{
    PytestPath: "/usr/local/bin/py.test",
    TestFiles:  []string{"fixtures/test.py", "/home/foo/test.py"},
  }

  err = provisioner.Prepare(noTestFileConfig)
  if !(errors.Is(err, os.ErrNotExist)) {
    test.Errorf("Prepare function did not fail correctly on nonexistent testfile")
  }
}
