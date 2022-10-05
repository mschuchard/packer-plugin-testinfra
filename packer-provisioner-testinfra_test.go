package main

import (
  "os"
  "testing"
  "errors"

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

// test provisioner prepare errors on unspecified test file
func TestProvisionerPrepareEmptyTestFile(test *testing.T) {
  var provisioner TestinfraProvisioner
  var emptyTestFileTestinfraConfig = &TestinfraConfig{
    PytestPath: "/usr/local/bin/py.test",
  }

  err := provisioner.Prepare(emptyTestFileTestinfraConfig)
  if err == nil {
    test.Errorf("Prepare function did not fail on unspecified testfile")
  }
}

// test provisioner prepare errors on nonexistent files
func TestProvisionerPrepareNonExistFiles(test *testing.T) {
  var provisioner TestinfraProvisioner

  // test no pytest
  var noPytestTestinfraConfig = &TestinfraConfig{
    PytestPath: "/home/foo/py.test",
    TestFile:   "fixtures/test.py",
  }

  err := provisioner.Prepare(noPytestTestinfraConfig)
  if !(errors.Is(err, os.ErrNotExist)) {
    test.Errorf("Prepare function did not fail on nonexistent pytest")
  }

  // test no testfile
  var noTestFileTestinfraConfig = &TestinfraConfig{
    PytestPath: "/usr/local/bin/py.test",
    TestFile:   "/home/foo/test.py",
  }

  err = provisioner.Prepare(noTestFileTestinfraConfig)
  if !(errors.Is(err, os.ErrNotExist)) {
    test.Errorf("Prepare function did not fail on nonexistent testfile")
  }
}

// test provisioner determineCommunication properly determines communication strings
func TestProvisionerDetermineCommunication(test *testing.T) {
  var provisioner TestinfraProvisioner

  // test ssh with empty httpaddr
  generatedData := map[string]interface{}{
    "ConnType": "ssh",
    "User": "me",
    "SSHPrivateKeyFile": "/path/to/sshprivatekeyfile",
    "SSHAgentAuth": true,
    "Host": "192.168.0.1",
    "Port": int64(22),
    "PackerHTTPAddr": "",
    "ID": "1234567890",
  }

  provisioner.generatedData = generatedData

  communication, err := provisioner.determineCommunication()
  if err != nil {
    test.Errorf("determineCommunication function failed to determine ssh: %s", err)
  }
  if communication != "--hosts=me@192.168.0.1:22 --ssh-identity-file=/path/to/sshprivatekeyfile --ssh-extra-args=\"-o StrictHostKeyChecking=no\"" {
    test.Errorf("Communication string incorrectly determined: %s", communication)
  }

  // test winrm with httpaddr
  generatedData = map[string]interface{}{
    "ConnType": "winrm",
    "User": "me",
    "Password": "password",
    "Host": "192.168.0.1",
    "Port": int64(5985),
    "PackerHTTPAddr": "192.168.0.1:5986",
    "ID": "1234567890",
  }

  provisioner.generatedData = generatedData

  communication, err = provisioner.determineCommunication()
  if err != nil {
    test.Errorf("determineCommunication function failed to determine winrm: %s", err)
  }
  if communication != "--hosts=winrm://me:password@192.168.0.1:5986" {
    test.Errorf("Communication string incorrectly determined: %s", communication)
  }

  // test docker
  generatedData = map[string]interface{}{
    "ConnType": "docker",
    "User": "me",
    "Host": "192.168.0.1",
    "Port": int64(0),
    "PackerHTTPAddr": "",
    "ID": "1234567890abcdefg",
  }

  provisioner.generatedData = generatedData

  communication, err = provisioner.determineCommunication()
  if err != nil {
    test.Errorf("determineCommunication function failed to determine docker: %s", err)
  }
  if communication != "--hosts=docker://1234567890abcdefg" {
    test.Errorf("Communication string incorrectly determined: %s", communication)
  }

  // test fails on no communication
  generatedData = map[string]interface{}{
    "ConnType": "unknown",
    "User": "me",
    "Host": "192.168.0.1",
    "Port": int64(22),
    "PackerHTTPAddr": "",
    "ID": "1234567890abcdefg",
  }

  provisioner.generatedData = generatedData

  _, err = provisioner.determineCommunication()
  if err == nil {
    test.Errorf("DetermineCommunication function did not fail on unknown connection type")
  }
}

// test provisioner determineSSHAuth properly determines ssh private key file location
func TestProvisionerDetermineSSHAuth(test *testing.T) {
  var provisioner TestinfraProvisioner

  // dummy up fake ssh private key
  generatedData := map[string]interface{}{
    "SSHPrivateKey": "abcdefg12345",
  }

  provisioner.generatedData = generatedData

  // test successfully returns ssh private key file location
  sshPrivateKeyFile, err := provisioner.determineSSHAuth("/tmp/sshprivatekeyfile", false)
  if err != nil {
    test.Errorf("determineSSHAuth failed to determine ssh private key file location: %s", err)
  }
  if sshPrivateKeyFile != "/tmp/sshprivatekeyfile" {
    test.Errorf("ssh private key file location incorrectly determined: %s", sshPrivateKeyFile)
  }

  // test successfully uses empty ssh private key file
  sshPrivateKeyFile, err = provisioner.determineSSHAuth("", true)
  if err != nil {
    test.Errorf("determineSSHAuth failed to determine keyless ssh: %s", err)
  }
  if sshPrivateKeyFile != "" {
    test.Errorf("ssh private key file empty location incorrectly determined: %s", sshPrivateKeyFile)
  }

  // test successfully creates tmpfile for private key
  sshPrivateKeyFile, err = provisioner.determineSSHAuth("", false)
  if err != nil {
    test.Errorf("determineSSHAuth failed to create tmpfile and return location of written ssh private key: %s", err)
  }
}
