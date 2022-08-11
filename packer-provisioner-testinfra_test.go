package main

import (
  "testing"

  "github.com/hashicorp/packer-plugin-sdk/packer"
)

// global helper vars for tests
var basicTestinfraConfig = &TestinfraConfig{
  PytestPath: "/usr/local/bin/py.test",
  TestFile:   "fixtures/test.py",
}

var minTestinfraConfig = &TestinfraConfig{
  TestFile: "fixtures/test.py",
}

// test config for packer template/config data
func TestProvisionerConfig(test *testing.T) {
  var provisioner = &TestinfraProvisioner{
    config: *basicTestinfraConfig,
  }

  if provisioner.config.PytestPath != "/usr/local/bin/py.test" || provisioner.config.TestFile != "fixtures/test.py" {
    test.Errorf("Provisioner config struct not initialized correctly")
  }
}

// test struct for provisioner interface
func TestProvisionerInterface(test *testing.T) {
  var raw interface{} = &TestinfraProvisioner{}
  if _, ok := raw.(packer.Provisioner); !ok {
    test.Errorf("Testinfra config struct must be a Provisioner")
  }
}

// test provisioner prepare with basic config
func TestProvisionerPrepareBasic(test *testing.T) {
  var provisioner TestinfraProvisioner

  err := provisioner.Prepare(basicTestinfraConfig)
  if err != nil {
    test.Errorf("Prepare function failed with basic Testinfra Packer config")
  }
}

// test provisioner prepare with minimal config
func TestProvisionerPrepareMinimal(test *testing.T) {
  var provisioner TestinfraProvisioner

  err := provisioner.Prepare(minTestinfraConfig)
  if err != nil {
    test.Errorf("Prepare function failed with minimal Testinfra Packer config")
  }

  if provisioner.config.PytestPath != "py.test" {
    test.Errorf("Default setting for PytestPath is incorrect: %s", provisioner.config.PytestPath)
  }
}

// test provisioner prepare errors on nonexistent pytest
func TestProvisionerPrepareNoPytest(test *testing.T) {
  var provisioner TestinfraProvisioner
  var noPytestTestinfraConfig = &TestinfraConfig{
    PytestPath: "/home/foo/py.test",
    TestFile:   "fixtures/test.py",
  }

  err := provisioner.Prepare(noPytestTestinfraConfig)
  if err == nil {
    test.Errorf("Prepare function did not fail on nonexistent pytest")
  }
}

// test provisioner prepare errors on nonexistent test file
func TestProvisionerPrepareNoTestFile(test *testing.T) {
  var provisioner TestinfraProvisioner
  var noTestFileTestinfraConfig = &TestinfraConfig{
    PytestPath: "/usr/local/bin/py.test",
    TestFile:   "/home/foo/test.py",
  }

  err := provisioner.Prepare(noTestFileTestinfraConfig)
  if err == nil {
    test.Errorf("Prepare function did not fail on nonexistent testfile")
  }
}
