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

// test provisioner prepare
func TestProvisionerPrepareBasic(test *testing.T) {
  var provisioner TestinfraProvisioner

  err := provisioner.Prepare(basicTestinfraConfig)
  if err != nil {
    test.Errorf("Prepare function failed with basic Testinfra Packer config")
  }
}
