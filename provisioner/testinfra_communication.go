package testinfra

import (
	"errors"
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
func (provisioner *Provisioner) determineCommunication() ([]string, error) {
	// declare communication args
	var args []string

	// parse generated data for required values
	connectionType := provisioner.generatedData["ConnType"].(string)

	// determine communication string by packer connection type
	log.Printf("testinfra communicating via %s connection type", connectionType)

	// determine communication based on connection type
	switch connectionType {
	case "ssh":
		// assign user and host address
		user, httpAddr, err := provisioner.determineUserAddr()
		if err != nil {
			return nil, err
		}

		// assign ssh auth type and string (key file path or password)
		sshAuthType, sshAuthString, err := provisioner.determineSSHAuth()
		if err != nil {
			return nil, err
		}

		// determine additional args for ssh based on authentication information
		switch sshAuthType {
		// use ssh private key file
		case privateKeySSHAuth:
			log.Printf("SSH private key filesystem location is: %s", sshAuthString)

			// append args with ssh connection backend information (user, host, port), private key file, and no strict host key checking
			args = append(args, fmt.Sprintf("--hosts=ssh://%s@%s", user, httpAddr), fmt.Sprintf("--ssh-identity-file=%s", sshAuthString), "--ssh-extra-args=\"-o StrictHostKeyChecking=no\"")
		// use ssh password
		case passwordSSHAuth:
			log.Print("utilizing SSH password for communicator authentication")

			// append args with ssh connection backend information (user, password, host, port), and no strict host key checking
			args = append(args, fmt.Sprintf("--hosts=ssh://%s:%s@%s", user, sshAuthString, httpAddr), "--ssh-extra-args=\"-o StrictHostKeyChecking=no\"")
		// use ssh agent auth
		default:
			log.Print("utilizing SSH Agent auth for communicator authentication")

			// append args with ssh connection backend information (user, host, port), and no strict host key checking
			args = append(args, fmt.Sprintf("--hosts=ssh://%s@%s", user, httpAddr), "--ssh-extra-args=\"-o StrictHostKeyChecking=no\"")
		}
	case "winrm":
		// assign user and host address
		user, httpAddr, err := provisioner.determineUserAddr()
		if err != nil {
			return nil, err
		}

		// assign winrm password preferably from winrmpassword
		winrmPassword, ok := provisioner.generatedData["WinRMPassword"].(string)
		// otherwise retry with general password
		if !ok || len(winrmPassword) == 0 {
			winrmPassword, ok = provisioner.generatedData["Password"].(string)

			// no winrm password available
			if !ok || len(winrmPassword) == 0 {
				log.Print("winrm communicator password could not be determined from available Packer data")
				return nil, errors.New("unknown winrm password")
			}
		}

		// append args with winrm connection backend information (user, password, host, port)
		args = append(args, fmt.Sprintf("--hosts=winrm://%s:%s@%s", user, winrmPassword, httpAddr))
	case "docker", "podman", "lxc":
		// determine instanceid
		instanceID, ok := provisioner.generatedData["ID"].(string)
		if !ok || len(instanceID) == 0 {
			log.Print("instance id could not be determined")
			return nil, errors.New("unknown instance id")
		}

		// append args with container connection backend information (instanceid)
		args = append(args, fmt.Sprintf("--hosts=%s://%s", connectionType, instanceID))
	default:
		log.Printf("communication backend with machine image is not supported, and was resolved to '%s'", connectionType)
		return nil, errors.New("unsupported communication type")
	}

	return args, nil
}

// determine and return user and host address
func (provisioner *Provisioner) determineUserAddr() (string, string, error) {
	// determine user
	user, ok := provisioner.generatedData["SSHUsername"].(string)
	if !ok || len(user) == 0 {
		// fallback to general user (usually packer)
		user, ok = provisioner.generatedData["User"].(string)

		if !ok || len(user) == 0 {
			log.Print("remote user could not be determined from available Packer data")
			return "", "", errors.New("unknown remote user")
		}
	}

	// determine host address and port
	var httpAddr string
	ipaddress, ok := provisioner.generatedData["Host"].(string)
	if !ok || len(ipaddress) == 0 {
		// this is more likely to be a file upload staging server, but fallback anyway
		httpAddr, ok = provisioner.generatedData["PackerHTTPAddr"].(string)

		if !ok || len(httpAddr) == 0 {
			log.Print("host address and port could not be determined")
			return "", "", errors.New("unknown host")
		}
	} else {
		// valid ip address so now determine port
		port, ok := provisioner.generatedData["SSHPort"].(int64)

		// fall back to general port
		if !ok || port == int64(0) {
			port = provisioner.generatedData["Port"].(int64)

			//if !ok || port == int64(0) {
			if port == int64(0) {
				log.Print("host port could not be determined")
				return "", "", errors.New("unknown host port")
			}
		}

		httpAddr = fmt.Sprintf("%s:%d", ipaddress, port)
	}

	return user, httpAddr, nil
}

// determine and return ssh authentication
func (provisioner *Provisioner) determineSSHAuth() (SSHAuth, string, error) {
	// assign ssh password preferably from sshpassword
	sshPassword, ok := provisioner.generatedData["SSHPassword"].(string)

	// otherwise retry with general password
	if !ok || len(sshPassword) == 0 {
		sshPassword, ok = provisioner.generatedData["Password"].(string)
	}

	// ssh is being used with password auth and we have a password
	if ok && len(sshPassword) > 0 {
		return passwordSSHAuth, sshPassword, nil
	} else { // ssh is being used with private key or agent auth so determine that instead
		// parse generated data for ssh private key and agent auth info
		sshPrivateKeyFile := provisioner.generatedData["SSHPrivateKeyFile"].(string)
		sshAgentAuth := provisioner.generatedData["SSHAgentAuth"].(bool)

		if len(sshPrivateKeyFile) > 0 {
			// we have a specified private key file so use that
			return privateKeySSHAuth, sshPrivateKeyFile, nil
		} else if sshAgentAuth {
			// we can use an empty/automatic private key with ssh agent auth
			return agentSSHAuth, sshPrivateKeyFile, nil
		} else { // create a private key file instead from the privatekey data
			// write a tmpfile for storing a private key
			tmpSSHPrivateKey, err := tmp.File("testinfra-key")
			if err != nil {
				log.Print("error creating a temp file for the ssh private key")
				return "", "", err
			}

			// attempt to obtain a private key
			SSHPrivateKey := provisioner.generatedData["SSHPrivateKey"].(string)

			// write the private key to the tmpfile
			_, err = tmpSSHPrivateKey.WriteString(SSHPrivateKey)
			if err != nil {
				log.Print("failed to write ssh private key to temp file")
				return "", "", err
			}

			// and then close the tmpfile storing the private key
			err = tmpSSHPrivateKey.Close()
			if err != nil {
				log.Print("failed to close ssh private key temp file")
				return "", "", err
			}

			return privateKeySSHAuth, tmpSSHPrivateKey.Name(), nil
		}
	}
}
