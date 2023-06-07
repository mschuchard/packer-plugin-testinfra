package testinfra

import (
	"fmt"
	"log"

	"github.com/hashicorp/packer-plugin-sdk/tmp"
)

// ssh auth type with pseudo-enum
type SSHAuth string

const (
	passwordSSHAuth   SSHAuth = "password"
	agentSSHAuth      SSHAuth = "agent"
	privateKeySSHAuth SSHAuth = "privateKey"
)

// determine and return appropriate communication string for pytest/testinfra
func (provisioner *Provisioner) determineCommunication() (string, error) {
	// parse generated data for required values
	connectionType := provisioner.generatedData["ConnType"].(string)
	user, ok := provisioner.generatedData["SSHUsername"].(string)
	if !ok {
		user = provisioner.generatedData["User"].(string)
	}
	ipaddress := provisioner.generatedData["Host"].(string)
	port, ok := provisioner.generatedData["SSHPort"].(int64)
	if !ok {
		port = provisioner.generatedData["Port"].(int64)
	}
	httpAddr := fmt.Sprintf("%s:%d", ipaddress, port)
	if len(ipaddress) == 0 {
		// this is more likely to be a file upload staging server, but fallback anyway
		httpAddr = provisioner.generatedData["PackerHTTPAddr"].(string)
	}
	instanceID := provisioner.generatedData["ID"].(string)

	// determine communication string by packer connection type
	log.Printf("Testinfra communicating via %s connection type", connectionType)
	communication := ""

	switch connectionType {
	case "ssh":
		// assign ssh auth type and string (key file path or password)
		sshAuthType, sshAuthString, err := provisioner.determineSSHAuth()
		if err != nil {
			return "", err
		}

		// initialize whitespace string as default arg (implied for sshAgentAuth type)
		sshIdentity := " "
		// use ssh private key file
		if sshAuthType == privateKeySSHAuth {
			log.Printf("SSH private key filesystem location is: %s", sshAuthString)
			sshIdentity = fmt.Sprintf(" --ssh-identity-file=%s ", sshAuthString)
		} else if sshAuthType == passwordSSHAuth { // use ssh password
			log.Printf("Utilizing SSH password for communicator authentication.")
			// modify user string to also include password
			user = fmt.Sprintf("%s:%s", user, sshAuthString)
		}

		communication = fmt.Sprintf("--hosts=%s@%s%s--ssh-extra-args=\"-o StrictHostKeyChecking=no\"", user, httpAddr, sshIdentity)
	case "winrm":
		// assign winrm password preferably from winrmpassword
		winRMPassword, ok := provisioner.generatedData["WinRMPassword"].(string)
		// otherwise retry with general password
		if !ok {
			winRMPassword, ok = provisioner.generatedData["Password"].(string)
		}
		// no winrm password available
		if !ok {
			return "", fmt.Errorf("WinRM communicator password could not be determined from available Packer data.")
		}

		communication = fmt.Sprintf("--hosts=winrm://%s:%s@%s", user, winRMPassword, httpAddr)
	case "docker":
		communication = fmt.Sprintf("--hosts=docker://%s", instanceID)
	case "podman":
		communication = fmt.Sprintf("--hosts=podman://%s", instanceID)
	case "lxc":
		communication = fmt.Sprintf("--hosts=lxc://%s", instanceID)
	}
	if len(communication) == 0 {
		return "", fmt.Errorf("Communication with machine image could not be properly determined")
	}

	return communication, nil
}

// determine and return ssh authentication
func (provisioner *Provisioner) determineSSHAuth() (SSHAuth, string, error) {
	// assign ssh password preferably from sshpassword
	sshPassword, ok := provisioner.generatedData["SSHPassword"].(string)

	// otherwise retry with general password
	if !ok {
		sshPassword, ok = provisioner.generatedData["Password"].(string)
	}

	// ssh is being used with password auth and we have a password
	if ok {
		return passwordSSHAuth, sshPassword, nil
	} else { // ssh is being used with private key or agent auth so determine that instead
		// parse generated data for ssh private key and agent auth info
		sshPrivateKeyFile := provisioner.generatedData["SSHPrivateKeyFile"].(string)
		sshAgentAuth := provisioner.generatedData["SSHAgentAuth"].(bool)

		if len(sshPrivateKeyFile) > 0 {
			// we have a legitimate private key file so use that
			return privateKeySSHAuth, sshPrivateKeyFile, nil
		} else if sshAgentAuth {
			// we can use an empty private key with ssh agent auth
			return agentSSHAuth, sshPrivateKeyFile, nil
		} else { // create a private key file instead
			// write a tmpfile for storing a private key
			tmpSSHPrivateKey, err := tmp.File("testinfra-key")
			if err != nil {
				return "", "", fmt.Errorf("Error creating a temp file for the ssh private key: %v", err.Error())
			}

			// attempt to obtain a private key
			SSHPrivateKey := provisioner.generatedData["SSHPrivateKey"].(string)

			// write the private key to the tmpfile
			_, err = tmpSSHPrivateKey.WriteString(SSHPrivateKey)
			if err != nil {
				return "", "", fmt.Errorf("Failed to write ssh private key to temp file")
			}

			// and then close the tmpfile storing the private key
			err = tmpSSHPrivateKey.Close()
			if err != nil {
				return "", "", fmt.Errorf("Failed to close ssh private key temp file")
			}

			return privateKeySSHAuth, tmpSSHPrivateKey.Name(), nil
		}
	}
}
