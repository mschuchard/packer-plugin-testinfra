package testinfra

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/tmp"
)

// determine and return appropriate communication string for pytest/testinfra
func (provisioner *Provisioner) determineCommunication(ui packer.Ui) ([]string, error) {
	// declare communication args
	var args []string

	// determine communication string by packer connection type
	connectionType, ok := provisioner.generatedData["ConnType"].(string)
	if !ok {
		ui.Error("packer is unable to resolve the communicator connection type")
		return nil, errors.New("unknown communicator connection type")
	}

	ui.Sayf("testinfra communicating via %s connection type", connectionType)

	// determine communication based on connection type
	switch connectionType {
	case "ssh":
		// assign user and host address
		user, httpAddr, err := provisioner.determineUserAddr(connectionType, ui)
		if err != nil {
			return nil, err
		}

		// assign ssh auth type and string (key file path or password)
		sshAuthType, sshAuthString, err := provisioner.determineSSHAuth(ui)
		if err != nil {
			return nil, err
		}
		log.Print("determined ssh authentication information")

		// determine additional args for ssh based on authentication information
		switch sshAuthType {
		// use ssh private key file
		case privateKey:
			ui.Say("utilizing SSH private key for communicator authentication")
			log.Printf("SSH private key filesystem location is: %s", sshAuthString)

			// append args with ssh connection backend information (user, host, port), private key file, and no strict host key checking
			args = append(args, fmt.Sprintf("--hosts=ssh://%s@%s", user, httpAddr), fmt.Sprintf("--ssh-identity-file=%s", sshAuthString), "--ssh-extra-args=\"-o StrictHostKeyChecking=no\"")
		// use ssh password
		case password:
			ui.Say("utilizing SSH password for communicator authentication")
			ui.Say("warning: this is typically invalid for Python to SSH interfacing, but this plugin will attempt it anyway")
			ui.Say("warning: consider using a passwordless private key or SSH agent instead")

			// append args with ssh connection backend information (user, password, host, port), and no strict host key checking
			args = append(args, fmt.Sprintf("--hosts=ssh://%s:%s@%s", user, sshAuthString, httpAddr), "--ssh-extra-args=\"-o StrictHostKeyChecking=no\"")
		// use ssh agent auth
		case agent:
			ui.Say("utilizing SSH Agent for communicator authentication")

			// append args with ssh connection backend information (user, host, port), and no strict host key checking
			args = append(args, fmt.Sprintf("--hosts=ssh://%s@%s", user, httpAddr), "--ssh-extra-args=\"-o StrictHostKeyChecking=no\"")
		// somehow not in enum
		default:
			ui.Errorf("unsupported ssh authentication type selected: %s", sshAuthType)

			return nil, errors.New("unsupported ssh auth type")
		}
	case "winrm":
		// assign user and host address
		user, httpAddr, err := provisioner.determineUserAddr(connectionType, ui)
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
				ui.Error("winrm communicator password could not be determined from available Packer data")
				return nil, errors.New("unknown winrm password")
			}
		}

		// winrm optional arguments
		optionalArgs, err := provisioner.determineWinRMArgs(ui)
		if err != nil {
			ui.Error("winrm communicator optional arguments could not be determined from available Packer data")
			return nil, err
		}

		// format string for testinfra connection backend setting
		connectionBackend := fmt.Sprintf("--hosts=winrm://%s:%s@%s%s", user, winrmPassword, httpAddr, strings.Join(optionalArgs, "&"))

		// append args with winrm connection backend information (user, password, host, port)
		args = append(args, connectionBackend)
	case "docker", "podman", "lxc":
		// determine instanceid
		instanceID, ok := provisioner.generatedData["ID"].(string)
		if !ok || len(instanceID) == 0 {
			ui.Error("instance id could not be determined")
			return nil, errors.New("unknown instance id")
		}

		// append args with container connection backend information (instanceid)
		args = append(args, fmt.Sprintf("--hosts=%s://%s", connectionType, instanceID))
	default:
		ui.Errorf("communication backend with machine image is not supported, and was resolved to '%s'", connectionType)
		return nil, errors.New("unsupported communication type")
	}

	log.Printf("determined communicator arguments as: %+q", args)

	return args, nil
}

// determine and return user and host address
func (provisioner *Provisioner) determineUserAddr(connectionType string, ui packer.Ui) (string, string, error) {
	// ssh and winrm provisioner generated data maps
	genDataMap := map[string]map[string]string{
		"ssh": {
			"user": "SSHUsername",
			"host": "SSHHost",
			"port": "SSHPort",
		},
		"winrm": {
			"user": "WinRMUser",
			"host": "WinRMHost",
			"port": "WinRMPort",
		},
	}

	// determine user based on connection protocol
	user, ok := provisioner.generatedData[genDataMap[connectionType]["user"]].(string)
	if !ok || len(user) == 0 {
		// fallback to general user (usually packer)
		user, ok = provisioner.generatedData["User"].(string)

		if !ok || len(user) == 0 {
			log.Print("remote user could not be determined from available Packer data")
			return "", "", errors.New("unknown remote user")
		}
	}

	// determine host address and port based on connection protocol
	ipaddress, ok := provisioner.generatedData[genDataMap[connectionType]["host"]].(string)
	if !ok || len(ipaddress) == 0 {
		// fallback to general host information
		ipaddress, ok = provisioner.generatedData["Host"].(string)

		if !ok || len(ipaddress) == 0 {
			log.Print("host address could not be determined from available Packer data")
			return "", "", errors.New("unknown host address")
		}
	}
	// valid ip address so now determine port
	port, ok := provisioner.generatedData[genDataMap[connectionType]["port"]].(int64)
	if !ok || port == int64(0) {
		// fallback to general port
		port, ok = provisioner.generatedData["Port"].(int64)

		if !ok || port == int64(0) {
			log.Print("host port could not be determined from available Packer data")
			return "", "", errors.New("unknown host port")
		}
	}

	// string format connection endpoint
	httpAddr := fmt.Sprintf("%s:%d", ipaddress, port)

	log.Print("determined communication user and connection endpoint")

	return user, httpAddr, nil
}

// determine and return ssh authentication
func (provisioner *Provisioner) determineSSHAuth(ui packer.Ui) (sshAuth, string, error) {
	// assign ssh password
	sshPassword, ok := provisioner.generatedData["SSHPassword"].(string)
	// otherwise retry with general password
	if !ok || len(sshPassword) == 0 {
		sshPassword, ok = provisioner.generatedData["Password"].(string)
	}

	// ssh is being used with password auth and we have a password
	if ok && len(sshPassword) > 0 {
		return password, sshPassword, nil
	} else { // ssh is being used with private key or agent auth so determine that instead
		// parse generated data for ssh private key
		sshPrivateKeyFile, ok := provisioner.generatedData["SSHPrivateKeyFile"].(string)
		// retry with certificate if necessary
		if !ok || len(sshPrivateKeyFile) == 0 {
			sshPrivateKeyFile, ok = provisioner.generatedData["SSHCertificateFile"].(string)
		}

		if ok && len(sshPrivateKeyFile) > 0 {
			// we have a specified private key/cert file so use that
			return privateKey, sshPrivateKeyFile, nil
		} else if provisioner.generatedData["SSHAgentAuth"].(bool) {
			// we can use an empty/automatic private key with ssh agent auth
			return agent, "", nil
		} else { // we have no other options, so create a temp private key file from the packer data
			// write a tmpfile for storing a private key
			tmpSSHPrivateKey, err := tmp.File("testinfra-key")
			if err != nil {
				log.Print("error creating a temp file for the ssh private key")
				return "", "", err
			}

			// attempt to obtain a private key
			SSHPrivateKey, ok := provisioner.generatedData["SSHPrivateKey"].(string)
			if !ok || len("SSHPrivateKey") == 0 {
				log.Print("no SSH authentication information was available in Packer data")
				return "", "", errors.New("no ssh authentication")
			}

			// write the private key to the tmpfile
			if _, err = tmpSSHPrivateKey.WriteString(SSHPrivateKey); err != nil {
				log.Print("failed to write ssh private key to temp file")
				return "", "", err
			}

			// and then close the tmpfile storing the private key
			if err = tmpSSHPrivateKey.Close(); err != nil {
				log.Print("failed to close ssh private key temp file")
				return "", "", err
			}

			return privateKey, tmpSSHPrivateKey.Name(), nil
		}
	}
}

// determine and return winrm optional arguments
func (provisioner *Provisioner) determineWinRMArgs(ui packer.Ui) ([]string, error) {
	// declare optional args slice to contain and later return
	var optionalArgs []string

	// modify pywinrm connection backend for winrm communicator settings
	// check on disable ssl
	if useSSL, ok := provisioner.generatedData["WinRMUseSSL"].(bool); ok && !useSSL {
		optionalArgs = append(optionalArgs, "no_ssl=true")
	}
	// check on do not verify ssl
	if insecure, ok := provisioner.generatedData["WinRMInsecure"].(bool); ok && insecure {
		optionalArgs = append(optionalArgs, "no_verify_ssl=true")
	}
	// check on timeout
	if timeout, ok := provisioner.generatedData["WinRMTimeout"].(string); ok && timeout != "30m" {
		// mathematical conversion of winrmtimeout to seconds
		var timeoutSeconds int

		// cut around hours
		hours, remainder, found := strings.Cut(timeout, "h")
		// parse hours
		if found {
			if quantity, err := strconv.Atoi(hours); err == nil {
				// convert hours
				timeoutSeconds += quantity * 3600
			} else {
				log.Printf("WinRMTimeout Packer data is invalid value and/or format: %s", timeout)
				return nil, errors.New("invalid winrmtimeout")
			}
		}

		// cut around minutes
		minutes, remainder, found := strings.Cut(remainder, "m")
		// parse minutes
		if found {
			if quantity, err := strconv.Atoi(minutes); err == nil {
				// convert minutes
				timeoutSeconds += quantity * 60
			} else {
				log.Printf("WinRMTimeout Packer data is invalid value and/or format: %s", timeout)
				return nil, errors.New("invalid winrmtimeout")
			}
		}

		// cut around seconds
		seconds, found := strings.CutSuffix(remainder, "s")
		// parse seconds
		if found {
			if quantity, err := strconv.Atoi(seconds); err == nil {
				// add seconds
				timeoutSeconds += quantity
			} else {
				log.Printf("WinRMTimeout Packer data is invalid value and/or format: %s", timeout)
				return nil, errors.New("invalid winrmtimeout")
			}
		}

		// if no unit was "found" then data is clearly invalid/malformed
		if timeoutSeconds == 0 {
			log.Printf("WinRMTimeout Packer data is invalid value and/or format: %s", timeout)
			return nil, errors.New("invalid winrmtimeout")
		}

		optionalArgs = append(optionalArgs, fmt.Sprintf("read_timeout_sec=%d", timeoutSeconds))
	}

	// prefix first optional argument with ? character if it exists
	if len(optionalArgs) > 0 {
		optionalArgs[0] = "?" + optionalArgs[0]
	}

	return optionalArgs, nil
}
