package testinfra

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/packer"
)

// test provisioner determineCommunication properly determines communication strings
func TestProvisionerDetermineCommunication(test *testing.T) {
	// initialize simple test ui
	ui := packer.TestUi(test)

	var provisioner Provisioner

	// test ssh with httpaddr and password
	generatedData := map[string]interface{}{
		"ConnType":          "ssh",
		"SSHUsername":       "me",
		"SSHPassword":       "password",
		"SSHPrivateKeyFile": "/path/to/sshprivatekeyfile",
		"SSHAgentAuth":      false,
		"SSHHost":           "192.168.0.1",
		"SSHPort":           int64(22),
		"ID":                "1234567890",
	}

	provisioner.generatedData = generatedData

	communication, err := provisioner.determineCommunication(ui)
	if err != nil {
		test.Errorf("determineCommunication function failed to determine ssh: %s", err)
	}
	if !slices.Equal(communication, []string{fmt.Sprintf("--hosts=ssh://%s:%s@%s:%d", generatedData["SSHUsername"], generatedData["SSHPassword"], generatedData["SSHHost"], generatedData["SSHPort"]), "--ssh-extra-args=\"-o StrictHostKeyChecking=no\""}) {
		test.Errorf("communication string slice for ssh password incorrectly determined: %v", communication)
	}

	// test ssh with private key file
	delete(generatedData, "SSHPassword")
	provisioner.generatedData = generatedData

	communication, err = provisioner.determineCommunication(ui)
	if err != nil {
		test.Errorf("determineCommunication function failed to determine ssh: %s", err)
	}
	if !slices.Equal(communication, []string{fmt.Sprintf("--hosts=ssh://%s@%s:%d", generatedData["SSHUsername"], generatedData["SSHHost"], generatedData["SSHPort"]), fmt.Sprintf("--ssh-identity-file=%s", generatedData["SSHPrivateKeyFile"]), "--ssh-extra-args=\"-o StrictHostKeyChecking=no\""}) {
		test.Errorf("communication string slice for ssh private key incorrectly determined: %v", communication)
	}

	// test ssh with no private key but with agent auth
	generatedData["SSHPrivateKeyFile"] = ""
	generatedData["SSHAgentAuth"] = true
	provisioner.generatedData = generatedData

	communication, err = provisioner.determineCommunication(ui)
	if err != nil {
		test.Errorf("determineCommunication function failed to determine ssh: %s", err)
	}
	if !slices.Equal(communication, []string{fmt.Sprintf("--hosts=ssh://%s@%s:%d", generatedData["SSHUsername"], generatedData["SSHHost"], generatedData["SSHPort"]), "--ssh-extra-args=\"-o StrictHostKeyChecking=no\""}) {
		test.Errorf("communication string slice for ssh agent auth incorrectly determined: %v", communication)
	}

	// test winrm
	generatedData = map[string]interface{}{
		"ConnType":      "winrm",
		"WinRMUser":     "me",
		"WinRMPassword": "password",
		"WinRMHost":     "192.168.0.1",
		"WinRMPort":     int64(5986),
		"WinRMUseSSL":   false,
		"WinRMInsecure": true,
	}

	provisioner.generatedData = generatedData

	communication, err = provisioner.determineCommunication(ui)
	if err != nil {
		test.Errorf("determineCommunication function failed to determine winrm: %s", err)
	}
	if !slices.Equal(communication, []string{fmt.Sprintf("--hosts=winrm://%s:%s@%s:%d?no_ssl=true&no_verify_ssl=true", generatedData["WinRMUser"], generatedData["WinRMPassword"], generatedData["WinRMHost"], generatedData["WinRMPort"])}) {
		test.Errorf("communication string slice for winrm incorrectly determined: %v", communication)
	}

	// test fails on no winrmpassword or password
	delete(generatedData, "WinRMPassword")
	provisioner.generatedData = generatedData

	_, err = provisioner.determineCommunication(ui)
	if err == nil {
		test.Errorf("determineCommunication function did not fail on no available password")
	}

	// test docker
	generatedData = map[string]interface{}{
		"ConnType": "docker",
		"ID":       "1234567890abcdefg",
	}

	provisioner.generatedData = generatedData

	communication, err = provisioner.determineCommunication(ui)
	if err != nil {
		test.Errorf("determineCommunication function failed to determine docker: %s", err)
	}
	if !slices.Equal(communication, []string{fmt.Sprintf("--hosts=docker://%s", generatedData["ID"])}) {
		test.Errorf("communication string slice for docker incorrectly determined: %v", communication)
	}

	// test podman
	generatedData = map[string]interface{}{
		"ConnType": "podman",
		"ID":       "1234567890abcdefg",
	}

	provisioner.generatedData = generatedData

	communication, err = provisioner.determineCommunication(ui)
	if err != nil {
		test.Errorf("determineCommunication function failed to determine podman: %s", err)
	}
	if !slices.Equal(communication, []string{fmt.Sprintf("--hosts=podman://%s", generatedData["ID"])}) {
		test.Errorf("communication string slice for podman incorrectly determined: %v", communication)
	}

	// test lxc
	generatedData = map[string]interface{}{
		"ConnType": "lxc",
		"ID":       "1234567890abcdefg",
	}

	provisioner.generatedData = generatedData

	communication, err = provisioner.determineCommunication(ui)
	if err != nil {
		test.Errorf("determineCommunication function failed to determine lxc: %s", err)
	}
	if !slices.Equal(communication, []string{fmt.Sprintf("--hosts=lxc://%s", generatedData["ID"])}) {
		test.Errorf("communication string slice for lxc incorrectly determined: %v", communication)
	}

	// test fails on no communication
	generatedData = map[string]interface{}{
		"ConnType": "unknown",
		"User":     "me",
		"Host":     "192.168.0.1",
		"Port":     int64(22),
		"ID":       "1234567890abcdefg",
	}

	provisioner.generatedData = generatedData

	_, err = provisioner.determineCommunication(ui)
	if err == nil {
		test.Errorf("determineCommunication function did not fail on unknown connection type")
	}
}

// test provisioner determineUserAddr properly determines user and instance address
func TestDetermineUserAddr(test *testing.T) {
	var provisioner Provisioner

	// dummy up fake user and address data
	generatedData := map[string]interface{}{
		"User": "me",
		"Host": "192.168.0.1",
		"Port": int64(22),
	}

	provisioner.generatedData = generatedData

	user, httpAddr, err := provisioner.determineUserAddr("ssh")
	if err != nil {
		test.Error("determineUserAddr failed to determine user and address")
		test.Error(err)
	}
	if user != generatedData["User"] {
		test.Error("user was incorrectly determined")
		test.Errorf("expected: %s, actual: %s", generatedData["User"], user)
	}
	expectedHttpAddr := fmt.Sprintf("%s:%d", generatedData["Host"], generatedData["Port"])
	if httpAddr != expectedHttpAddr {
		test.Error("address was incorrectly determined")
		test.Errorf("expected: %s, actual: %s", expectedHttpAddr, httpAddr)
	}
}

// test provisioner determineSSHAuth properly determines authentication information
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
	if sshAuthType != passwordSSHAuth {
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
	if sshAuthType != privateKeySSHAuth {
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
	if sshAuthType != agentSSHAuth {
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
	if sshAuthType != privateKeySSHAuth {
		test.Errorf("ssh authentication type incorrectly determined: %s", sshAuthType)
	}
	if matched, _ := regexp.Match(`/tmp/testinfra-key\d+`, []byte(sshAuthString)); !matched {
		test.Errorf("temporary ssh private key file was not created in the expected location: %s", sshAuthString)
	}
	if sshPrivateKey, _ := os.ReadFile(sshAuthString); string(sshPrivateKey) != generatedData["SSHPrivateKey"] {
		test.Errorf("temporary ssh key file content is not the ssh private key: %s", sshPrivateKey)
	}
}

func TestProvisionerDetermineWinRMArgs(test *testing.T) {
	var provisioner Provisioner

	// test empty data
	args := provisioner.determineWinRMArgs()
	if len(args) > 0 {
		test.Error("optional arguments were not empty with empty provisioner data")
		test.Errorf("actual: %+q, expected: empty", args)
	}

	// test data resulting in no optional args
	provisioner.generatedData = map[string]interface{}{
		"WinRMUseSSL":   true,
		"WinRMInsecure": false,
	}
	args = provisioner.determineWinRMArgs()
	if len(args) > 0 {
		test.Error("optional arguments were not empty with provisioner data causing no optional arguments")
		test.Errorf("actual: %+q, expected: empty", args)
	}

	// test data with all arguments
	provisioner.generatedData = map[string]interface{}{
		"WinRMUseSSL":   false,
		"WinRMInsecure": true,
	}
	expectedArgs := []string{"?no_ssl=true", "no_verify_ssl=true"}

	args = provisioner.determineWinRMArgs()
	if !slices.Equal(expectedArgs, args) {
		test.Error("optional arguments were not all set according to corresponding provisioner data")
		test.Errorf("actual: %+q, expected: %+q", args, expectedArgs)
	}
}
