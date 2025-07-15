package testinfra

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"testing"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/packer"
)

// test provisioner determineCommunication properly determines communication strings
func TestProvisionerDetermineCommunication(test *testing.T) {
	// initialize simple test ui
	ui := packer.TestUi(test)

	var provisioner Provisioner

	// test ssh with httpaddr and password
	provisioner.generatedData = map[string]any{
		"ConnType":          "ssh",
		"SSHUsername":       "me",
		"SSHPassword":       "password",
		"SSHPrivateKeyFile": "/path/to/sshprivatekeyfile",
		"SSHAgentAuth":      false,
		"SSHHost":           "192.168.0.1",
		"SSHPort":           22,
		"ID":                "1234567890",
	}

	communication, err := provisioner.determineCommunication(ui)
	if err != nil {
		test.Errorf("determineCommunication function failed to determine ssh: %s", err)
	}
	if !slices.Equal(communication, []string{fmt.Sprintf("--hosts=ssh://%s:%s@%s:%d", provisioner.generatedData["SSHUsername"], provisioner.generatedData["SSHPassword"], provisioner.generatedData["SSHHost"], provisioner.generatedData["SSHPort"]), "--ssh-extra-args=\"-o StrictHostKeyChecking=no\""}) {
		test.Errorf("communication string slice for ssh password incorrectly determined: %v", communication)
	}

	// test ssh with private key file
	delete(provisioner.generatedData, "SSHPassword")

	communication, err = provisioner.determineCommunication(ui)
	if err != nil {
		test.Errorf("determineCommunication function failed to determine ssh: %s", err)
	}
	if !slices.Equal(communication, []string{fmt.Sprintf("--hosts=ssh://%s@%s:%d", provisioner.generatedData["SSHUsername"], provisioner.generatedData["SSHHost"], provisioner.generatedData["SSHPort"]), fmt.Sprintf("--ssh-identity-file=%s", provisioner.generatedData["SSHPrivateKeyFile"]), "--ssh-extra-args=\"-o StrictHostKeyChecking=no\""}) {
		test.Errorf("communication string slice for ssh private key incorrectly determined: %v", communication)
	}

	// test ssh with no private key but with agent auth
	provisioner.generatedData["SSHPrivateKeyFile"] = ""
	provisioner.generatedData["SSHAgentAuth"] = true

	communication, err = provisioner.determineCommunication(ui)
	if err != nil {
		test.Errorf("determineCommunication function failed to determine ssh: %s", err)
	}
	if !slices.Equal(communication, []string{fmt.Sprintf("--hosts=ssh://%s@%s:%d", provisioner.generatedData["SSHUsername"], provisioner.generatedData["SSHHost"], provisioner.generatedData["SSHPort"]), "--ssh-extra-args=\"-o StrictHostKeyChecking=no\""}) {
		test.Errorf("communication string slice for ssh agent auth incorrectly determined: %v", communication)
	}

	// test winrm
	provisioner.generatedData = map[string]any{
		"ConnType":      "winrm",
		"WinRMUser":     "me",
		"WinRMPassword": "password",
		"WinRMHost":     "192.168.0.1",
		"WinRMPort":     5986,
		"WinRMUseSSL":   false,
		"WinRMInsecure": true,
	}
	provisioner.generatedData["WinRMTimeout"], _ = time.ParseDuration("20m5s")

	communication, err = provisioner.determineCommunication(ui)
	if err != nil {
		test.Errorf("determineCommunication function failed to determine winrm: %s", err)
	}
	if !slices.Equal(communication, []string{fmt.Sprintf("--hosts=winrm://%s:%s@%s:%d?no_ssl=true&no_verify_ssl=true&read_timeout_sec=1205", provisioner.generatedData["WinRMUser"], provisioner.generatedData["WinRMPassword"], provisioner.generatedData["WinRMHost"], provisioner.generatedData["WinRMPort"])}) {
		test.Errorf("communication string slice for winrm incorrectly determined: %v", communication)
	}

	// test fails on no winrmpassword or password
	delete(provisioner.generatedData, "WinRMPassword")

	_, err = provisioner.determineCommunication(ui)
	if err == nil || err.Error() != "unknown winrm password" {
		test.Error("determineCommunication function did not fail on no available password")
	}

	// test docker
	provisioner.generatedData = map[string]any{
		"ConnType": "docker",
		"ID":       "1234567890abcdefg",
	}

	communication, err = provisioner.determineCommunication(ui)
	if err != nil {
		test.Errorf("determineCommunication function failed to determine docker: %s", err)
	}
	if !slices.Equal(communication, []string{fmt.Sprintf("--hosts=docker://%s", provisioner.generatedData["ID"])}) {
		test.Errorf("communication string slice for docker incorrectly determined: %v", communication)
	}

	// test podman
	provisioner.generatedData = map[string]any{
		"ConnType": "podman",
		"ID":       "1234567890abcdefg",
	}

	communication, err = provisioner.determineCommunication(ui)
	if err != nil {
		test.Errorf("determineCommunication function failed to determine podman: %s", err)
	}
	if !slices.Equal(communication, []string{fmt.Sprintf("--hosts=podman://%s", provisioner.generatedData["ID"])}) {
		test.Errorf("communication string slice for podman incorrectly determined: %v", communication)
	}

	// test lxc
	provisioner.generatedData = map[string]any{
		"ConnType": "lxc",
		"ID":       "1234567890abcdefg",
	}

	communication, err = provisioner.determineCommunication(ui)
	if err != nil {
		test.Errorf("determineCommunication function failed to determine lxc: %s", err)
	}
	if !slices.Equal(communication, []string{fmt.Sprintf("--hosts=lxc://%s", provisioner.generatedData["ID"])}) {
		test.Errorf("communication string slice for lxc incorrectly determined: %v", communication)
	}

	delete(provisioner.generatedData, "ID")
	if _, err = provisioner.determineCommunication(ui); err == nil || err.Error() != "unknown instance id" {
		test.Error("determineCommunication did not fail on unknown instance id")
	}

	// test fails on no communication
	provisioner.generatedData = map[string]any{
		"ConnType": "unknown",
		"User":     "me",
		"Host":     "192.168.0.1",
		"Port":     22,
		"ID":       "1234567890abcdefg",
	}

	_, err = provisioner.determineCommunication(ui)
	if err == nil || err.Error() != "unsupported communication type" {
		test.Error("determineCommunication function did not fail on unsupported connection type")
	}
}

// test provisioner determineUserAddr properly determines user and instance address
func TestDetermineUserAddr(test *testing.T) {
	var provisioner Provisioner
	ui := packer.TestUi(test)

	// dummy up fake user and address data
	provisioner.generatedData = map[string]any{
		"User": "me",
		"Host": "192.168.0.1",
		"Port": 22,
	}

	user, httpAddr, err := provisioner.determineUserAddr("ssh", ui)
	if err != nil {
		test.Error("determineUserAddr failed to determine user and address")
		test.Error(err)
	}
	if user != provisioner.generatedData["User"] {
		test.Error("user was incorrectly determined")
		test.Errorf("expected: %s, actual: %s", provisioner.generatedData["User"], user)
	}
	expectedHttpAddr := fmt.Sprintf("%s:%d", provisioner.generatedData["Host"], provisioner.generatedData["Port"])
	if httpAddr != expectedHttpAddr {
		test.Error("address was incorrectly determined")
		test.Errorf("expected: %s, actual: %s", expectedHttpAddr, httpAddr)
	}

	delete(provisioner.generatedData, "Port")
	if _, _, err = provisioner.determineUserAddr("ssh", ui); err == nil || err.Error() != "unknown host port" {
		test.Error("determineCommunication did not fail on unknown port")
	}

	delete(provisioner.generatedData, "Host")
	if _, _, err = provisioner.determineUserAddr("ssh", ui); err == nil || err.Error() != "unknown host address" {
		test.Error("determineCommunication did not fail on unknown host")
	}

	delete(provisioner.generatedData, "User")
	if _, _, err = provisioner.determineUserAddr("ssh", ui); err == nil || err.Error() != "unknown remote user" {
		test.Error("determineCommunication did not fail on unknown user")
	}
}

// test provisioner determineSSHAuth properly determines authentication information
func TestProvisionerDetermineSSHAuth(test *testing.T) {
	var provisioner Provisioner
	ui := packer.TestUi(test)

	// dummy up fake ssh data
	provisioner.generatedData = map[string]any{
		"Password":          "password",
		"SSHPrivateKey":     "abcdefg12345",
		"SSHPrivateKeyFile": "/tmp/sshprivatekeyfile",
		"SSHAgentAuth":      false,
	}

	// test successfully uses password for ssh auth
	sshAuthType, sshAuthString, err := provisioner.determineSSHAuth(ui)
	if err != nil {
		test.Errorf("determineSSHAuth failed to determine authentication password: %s", err)
	}
	if sshAuthType != password {
		test.Errorf("ssh authentication type incorrectly determined: %s", sshAuthType)
	}
	if sshAuthString != provisioner.generatedData["Password"] {
		test.Errorf("password content incorrectly determined: %s", sshAuthString)
	}

	// remove password from data
	delete(provisioner.generatedData, "Password")

	// test successfully returns ssh private key file location
	sshAuthType, sshAuthString, err = provisioner.determineSSHAuth(ui)
	if err != nil {
		test.Errorf("determineSSHAuth failed to determine ssh private key file location: %s", err)
	}
	if sshAuthType != privateKey {
		test.Errorf("ssh authentication type incorrectly determined: %s", sshAuthType)
	}
	if sshAuthString != provisioner.generatedData["SSHPrivateKeyFile"] {
		test.Errorf("ssh private key file location incorrectly determined: %s", sshAuthString)
	}

	// modify to empty ssh key and yes to ssh agent auth
	provisioner.generatedData["SSHPrivateKeyFile"] = ""
	provisioner.generatedData["SSHAgentAuth"] = true

	// test successfully uses empty ssh private key file
	sshAuthType, sshAuthString, err = provisioner.determineSSHAuth(ui)
	if err != nil {
		test.Errorf("determineSSHAuth failed to determine keyless ssh: %s", err)
	}
	if sshAuthType != agent {
		test.Errorf("ssh authentication type incorrectly determined: %s", sshAuthType)
	}
	if len(sshAuthString) > 0 {
		test.Errorf("ssh private key file empty location incorrectly determined: %s", sshAuthString)
	}

	// modify to no ssh agent auth
	provisioner.generatedData["SSHAgentAuth"] = false

	// test successfully creates tmpfile with expected content for private key
	sshAuthType, sshAuthString, err = provisioner.determineSSHAuth(ui)
	if err != nil {
		test.Errorf("determineSSHAuth failed to create tmpfile and return location of written ssh private key: %s", err)
	}
	if sshAuthType != privateKey {
		test.Errorf("ssh authentication type incorrectly determined: %s", sshAuthType)
	}
	if matched, _ := regexp.Match(`/tmp/testinfra-key\d+`, []byte(sshAuthString)); !matched {
		test.Errorf("temporary ssh private key file was not created in the expected location: %s", sshAuthString)
	}
	if sshPrivateKey, _ := os.ReadFile(sshAuthString); string(sshPrivateKey) != provisioner.generatedData["SSHPrivateKey"] {
		test.Errorf("temporary ssh key file content is not the ssh private key: %s", sshPrivateKey)
	}

	delete(provisioner.generatedData, "SSHPrivateKey")
	if _, _, err = provisioner.determineSSHAuth(ui); err == nil || err.Error() != "no ssh authentication" {
		test.Error("sshauth did not fail on no available ssh authentication information")
	}
}

func TestProvisionerDetermineWinRMArgs(test *testing.T) {
	var provisioner Provisioner
	ui := packer.TestUi(test)

	// test empty data
	args, err := provisioner.determineWinRMArgs(ui)
	if err != nil {
		test.Error(err)
	}
	if len(args) > 0 {
		test.Error("optional arguments were not empty with empty provisioner data")
		test.Errorf("actual: %+q, expected: empty", args)
	}

	// test data resulting in no optional args
	provisioner.generatedData = map[string]any{
		"WinRMUseSSL":   true,
		"WinRMInsecure": false,
	}
	provisioner.generatedData["WinRMTimeout"], _ = time.ParseDuration("30m")

	args, err = provisioner.determineWinRMArgs(ui)
	if err != nil {
		test.Error(err)
	}
	if len(args) > 0 {
		test.Error("optional arguments were not empty with provisioner data causing no optional arguments")
		test.Errorf("actual: %+q, expected: empty", args)
	}

	// test data with all arguments
	provisioner.generatedData = map[string]any{
		"WinRMUseSSL":   false,
		"WinRMInsecure": true,
	}
	provisioner.generatedData["WinRMTimeout"], _ = time.ParseDuration("1h5m2s")
	expectedArgs := []string{"?no_ssl=true", "no_verify_ssl=true", "read_timeout_sec=3902"}

	args, err = provisioner.determineWinRMArgs(ui)
	if err != nil {
		test.Error(err)
	}
	if !slices.Equal(expectedArgs, args) {
		test.Error("optional arguments were not all set according to corresponding provisioner data")
		test.Errorf("actual: %+q, expected: %+q", args, expectedArgs)
	}

	// test malformed winrmtimeout
	provisioner.generatedData["WinRMTimeout"], _ = time.ParseDuration("2a5l1z")
	if _, err = provisioner.determineWinRMArgs(ui); err == nil || err.Error() != "invalid winrmtimeout" {
		test.Error("win rm args did not fail on malformed timeout data")
		test.Error(err)
	}
}
