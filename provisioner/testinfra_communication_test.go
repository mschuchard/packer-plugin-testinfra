package testinfra

import (
  "fmt"
  "os"
  "regexp"
  "testing"
)

// test provisioner determineCommunication properly determines communication strings
func TestProvisionerDetermineCommunication(test *testing.T) {
  var provisioner Provisioner

  // test ssh with httpaddr and password
  generatedData := map[string]interface{}{
    "ConnType": "ssh",
    "User": "me",
    "Password": "password",
    "SSHPrivateKeyFile": "/path/to/sshprivatekeyfile",
    "SSHAgentAuth": false,
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
  if communication != fmt.Sprintf("--hosts=%s:%s@%s:%d --ssh-extra-args=\"-o StrictHostKeyChecking=no\"", generatedData["User"], generatedData["Password"], generatedData["Host"], generatedData["Port"]) {
    test.Errorf("Communication string incorrectly determined: %s", communication)
  }

  // test ssh with private key file
  delete(generatedData, "Password")
  provisioner.generatedData = generatedData

  communication, err = provisioner.determineCommunication()
  if err != nil {
    test.Errorf("determineCommunication function failed to determine ssh: %s", err)
  }
  if communication != fmt.Sprintf("--hosts=%s@%s:%d --ssh-identity-file=%s --ssh-extra-args=\"-o StrictHostKeyChecking=no\"", generatedData["User"], generatedData["Host"], generatedData["Port"], generatedData["SSHPrivateKeyFile"]) {
    test.Errorf("Communication string incorrectly determined: %s", communication)
  }

  // test ssh with no private key but with agent auth
  generatedData["SSHPrivateKeyFile"] = ""
  generatedData["SSHAgentAuth"] =  true

  provisioner.generatedData = generatedData

  communication, err = provisioner.determineCommunication()
  if err != nil {
    test.Errorf("determineCommunication function failed to determine ssh: %s", err)
  }
  if communication != fmt.Sprintf("--hosts=%s@%s:%d --ssh-extra-args=\"-o StrictHostKeyChecking=no\"", generatedData["User"], generatedData["Host"], generatedData["Port"]) {
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
  var provisioner Provisioner

  // dummy up fake ssh data
  generatedData := map[string]interface{}{
    "Password":          "password",
    "SSHPrivateKey":     "abcdefg12345",
    "SSHPrivateKeyFile": "/tmp/sshprivatekeyfile",
    "SSHAgentAuth":      false,
  }

  provisioner.generatedData = generatedData

  // test successfully uses password for ssh auth
  sshAuthType, sshAuthString, err := provisioner.determineSSHAuth()
  if err != nil {
    test.Errorf("determineSSHAuth failed to determine authentication password: %s", err)
  }
  if sshAuthType != "password" {
    test.Errorf("ssh authentication type incorrectly determined: %s", sshAuthType)
  }
  if sshAuthString != generatedData["Password"] {
    test.Errorf("password content incorrectly determined: %s", sshAuthString)
  }

  // remove password from data
  delete(generatedData, "Password")
  provisioner.generatedData = generatedData

  // test successfully returns ssh private key file location
  sshAuthType, sshAuthString, err = provisioner.determineSSHAuth()
  if err != nil {
    test.Errorf("determineSSHAuth failed to determine ssh private key file location: %s", err)
  }
  if sshAuthType != "privateKey" {
    test.Errorf("ssh authentication type incorrectly determined: %s", sshAuthType)
  }
  if sshAuthString != generatedData["SSHPrivateKeyFile"] {
    test.Errorf("ssh private key file location incorrectly determined: %s", sshAuthString)
  }

  // modify to empty ssh key and yes to ssh agent auth
  generatedData["SSHPrivateKeyFile"] = ""
  generatedData["SSHAgentAuth"] = true

  provisioner.generatedData = generatedData

  // test successfully uses empty ssh private key file
  sshAuthType, sshAuthString, err = provisioner.determineSSHAuth()
  if err != nil {
    test.Errorf("determineSSHAuth failed to determine keyless ssh: %s", err)
  }
  if sshAuthType != "agent" {
    test.Errorf("ssh authentication type incorrectly determined: %s", sshAuthType)
  }
  if len(sshAuthString) > 0 {
    test.Errorf("ssh private key file empty location incorrectly determined: %s", sshAuthString)
  }

  // modify to no ssh agent auth
  generatedData["SSHAgentAuth"] = false

  provisioner.generatedData = generatedData

  // test successfully creates tmpfile with expected content for private key
  sshAuthType, sshAuthString, err = provisioner.determineSSHAuth()
  if err != nil {
    test.Errorf("determineSSHAuth failed to create tmpfile and return location of written ssh private key: %s", err)
  }
  if sshAuthType != "privateKey" {
    test.Errorf("ssh authentication type incorrectly determined: %s", sshAuthType)
  }
  if matched, _ := regexp.Match(`/tmp/testinfra-key\d+`, []byte(sshAuthString)); !matched {
    test.Errorf("temporary ssh private key file was not created in the expected location: %s", sshAuthString)
  }
  if sshPrivateKey, _ := os.ReadFile(sshAuthString); string(sshPrivateKey) != generatedData["SSHPrivateKey"] {
    test.Errorf("temporary ssh key file content is not the ssh private key: %s", sshPrivateKey)
  }
}
