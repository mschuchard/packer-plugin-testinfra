package main

import (
  "fmt"
  "os"
  "reflect"
  "strings"
  "regexp"
  "testing"
  "errors"

  "github.com/hashicorp/packer-plugin-sdk/packer"
)

// global helper vars for tests
var basicTestinfraConfig = &TestinfraConfig{
  Marker:     "fast",
  Processes:  4,
  PytestPath: "/usr/local/bin/py.test",
  Sudo:       true,
  TestFiles:  []string{"fixtures/test.py"},
}

var minTestinfraConfig = &TestinfraConfig{
  TestFiles: []string{"fixtures/test.py"},
}

// test basic config for packer template/config data
func TestProvisionerConfig(test *testing.T) {
  var provisioner = &TestinfraProvisioner{
    config: *basicTestinfraConfig,
  }
  configInput := *basicTestinfraConfig

  if provisioner.config.PytestPath != configInput.PytestPath || !(reflect.DeepEqual(provisioner.config.TestFiles, configInput.TestFiles)) || provisioner.config.Marker != configInput.Marker || provisioner.config.Processes != configInput.Processes || provisioner.config.Sudo != configInput.Sudo {
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
    TestFiles:  []string{"fixtures/test.py"},
  }

  err := provisioner.Prepare(noPytestTestinfraConfig)
  if !(errors.Is(err, os.ErrNotExist)) {
    test.Errorf("Prepare function did not fail correctly on nonexistent pytest")
  }

  // test nonexistent testfile
  var noTestFileTestinfraConfig = &TestinfraConfig{
    PytestPath: "/usr/local/bin/py.test",
    TestFiles:  []string{"fixtures/test.py", "/home/foo/test.py"},
  }

  err = provisioner.Prepare(noTestFileTestinfraConfig)
  if !(errors.Is(err, os.ErrNotExist)) {
    test.Errorf("Prepare function did not fail correctly on nonexistent testfile")
  }
}

// test provisioner determineExecCmd properly determines execution command
func TestProvisionerCmdExec(test *testing.T) {
  // test basic config with ssh generated data
  var provisioner = &TestinfraProvisioner{
    config: *basicTestinfraConfig,
  }

  generatedData := map[string]interface{}{
    "ConnType": "ssh",
    "User": "me",
    "SSHPrivateKeyFile": "/path/to/sshprivatekeyfile",
    "SSHAgentAuth": true,
    "Host": "192.168.0.1",
    "Port": int64(22),
    "PackerHTTPAddr": "192.168.0.1:8200",
    "ID": "1234567890",
  }

  provisioner.generatedData = generatedData

  execCmd, err := provisioner.determineExecCmd()
  if err != nil {
    test.Errorf("determineExecCmd function failed to determine execution command for basic config with SSH communicator: %s", err)
  }
  if execCmd.String() != fmt.Sprintf("%s -v --hosts=%s@%s:%d --ssh-identity-file=%s --ssh-extra-args=\"-o StrictHostKeyChecking=no\" %s -m %s -n %d --sudo", provisioner.config.PytestPath, generatedData["User"], generatedData["Host"], generatedData["Port"], generatedData["SSHPrivateKeyFile"], strings.Join(provisioner.config.TestFiles, ""), provisioner.config.Marker, provisioner.config.Processes) {
    test.Errorf("determineExecCmd function failed to properly determine execution command for basic config with SSH communicator: %s", execCmd.String())
  }
}

// test provisioner determineCommunication properly determines communication strings
func TestProvisionerDetermineCommunication(test *testing.T) {
  var provisioner TestinfraProvisioner

  // test ssh with httpaddr
  generatedData := map[string]interface{}{
    "ConnType": "ssh",
    "User": "me",
    "SSHPrivateKeyFile": "/path/to/sshprivatekeyfile",
    "SSHAgentAuth": true,
    "Host": "192.168.0.1",
    "Port": int64(22),
    "PackerHTTPAddr": "192.168.0.1:8200",
    "ID": "1234567890",
  }

  provisioner.generatedData = generatedData

  communication, err := provisioner.determineCommunication()
  if err != nil {
    test.Errorf("determineCommunication function failed to determine ssh: %s", err)
  }
  if communication != fmt.Sprintf("--hosts=%s@%s:%d --ssh-identity-file=%s --ssh-extra-args=\"-o StrictHostKeyChecking=no\"", generatedData["User"], generatedData["Host"], generatedData["Port"], generatedData["SSHPrivateKeyFile"]) {
    test.Errorf("Communication string incorrectly determined: %s", communication)
  }

  // test winrm with empty host, port, and winrmpassword
  generatedData = map[string]interface{}{
    "ConnType": "winrm",
    "User": "me",
    "Password": "password",
    "Host": "",
    "Port": int64(0),
    "PackerHTTPAddr": "192.168.0.1:5986",
    "ID": "1234567890",
  }

  provisioner.generatedData = generatedData

  communication, err = provisioner.determineCommunication()
  if err != nil {
    test.Errorf("determineCommunication function failed to determine winrm: %s", err)
  }
  if communication != fmt.Sprintf("--hosts=winrm://%s:%s@%s", generatedData["User"], generatedData["Password"], generatedData["PackerHTTPAddr"]) {
    test.Errorf("Communication string incorrectly determined: %s", communication)
  }

  // test fails on no winrmpassword or password
  delete(generatedData, "Password")
  provisioner.generatedData = generatedData

  _, err = provisioner.determineCommunication()
  if err == nil {
    test.Errorf("DetermineCommunication function did not fail on no available password")
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
  if communication != fmt.Sprintf("--hosts=docker://%s", generatedData["ID"]) {
    test.Errorf("Communication string incorrectly determined: %s", communication)
  }

  // test podman
  generatedData = map[string]interface{}{
    "ConnType": "podman",
    "User": "me",
    "Host": "192.168.0.1",
    "Port": int64(0),
    "PackerHTTPAddr": "",
    "ID": "1234567890abcdefg",
  }

  provisioner.generatedData = generatedData

  communication, err = provisioner.determineCommunication()
  if err != nil {
    test.Errorf("determineCommunication function failed to determine podman: %s", err)
  }
  if communication != fmt.Sprintf("--hosts=podman://%s", generatedData["ID"]) {
    test.Errorf("Communication string incorrectly determined: %s", communication)
  }

  // test lxc
  generatedData = map[string]interface{}{
    "ConnType": "lxc",
    "User": "me",
    "Host": "192.168.0.1",
    "Port": int64(0),
    "PackerHTTPAddr": "",
    "ID": "1234567890abcdefg",
  }

  provisioner.generatedData = generatedData

  communication, err = provisioner.determineCommunication()
  if err != nil {
    test.Errorf("determineCommunication function failed to determine lxc: %s", err)
  }
  if communication != fmt.Sprintf("--hosts=lxc://%s", generatedData["ID"]) {
    test.Errorf("Communication string incorrectly determined: %s", communication)
  }

  // test fails on no communication
  generatedData = map[string]interface{}{
    "ConnType": "unknown",
    "User": "me",
    "Host": "192.168.0.1",
    "Port": int64(22),
    "PackerHTTPAddr": "192.168.0.1:22",
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
    "SSHPrivateKey":     "abcdefg12345",
    "SSHPrivateKeyFile": "/tmp/sshprivatekeyfile",
    "SSHAgentAuth":      false,
  }

  provisioner.generatedData = generatedData

  // test successfully returns ssh private key file location
  sshPrivateKeyFile, err := provisioner.determineSSHAuth()
  if err != nil {
    test.Errorf("determineSSHAuth failed to determine ssh private key file location: %s", err)
  }
  if sshPrivateKeyFile != generatedData["SSHPrivateKeyFile"] {
    test.Errorf("ssh private key file location incorrectly determined: %s", sshPrivateKeyFile)
  }

  // modify to empty ssh key and yes to ssh agent auth
  generatedData["SSHPrivateKeyFile"] = ""
  generatedData["SSHAgentAuth"] = true

  provisioner.generatedData = generatedData

  // test successfully uses empty ssh private key file
  sshPrivateKeyFile, err = provisioner.determineSSHAuth()
  if err != nil {
    test.Errorf("determineSSHAuth failed to determine keyless ssh: %s", err)
  }
  if len(sshPrivateKeyFile) > 0 {
    test.Errorf("ssh private key file empty location incorrectly determined: %s", sshPrivateKeyFile)
  }

  // modify to no ssh agent auth
  generatedData["SSHAgentAuth"] = false

  provisioner.generatedData = generatedData

  // test successfully creates tmpfile with expected content for private key
  sshPrivateKeyFile, err = provisioner.determineSSHAuth()
  if err != nil {
    test.Errorf("determineSSHAuth failed to create tmpfile and return location of written ssh private key: %s", err)
  }
  if matched, _ := regexp.Match(`/tmp/testinfra-key\d+`, []byte(sshPrivateKeyFile)); !matched {
    test.Errorf("temporary ssh private key file was not created in the expected location: %s", sshPrivateKeyFile)
  }
  if sshPrivateKey, _ := os.ReadFile(sshPrivateKeyFile); string(sshPrivateKey) != generatedData["SSHPrivateKey"] {
    test.Errorf("temporary ssh key file content is not the ssh private key: %s", sshPrivateKey)
  }
}
