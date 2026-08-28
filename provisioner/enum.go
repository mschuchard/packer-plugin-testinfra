package testinfra

import (
	"errors"
	"log"
	"slices"
)

// ssh auth type with pseudo-enum
type sshAuth string

const (
	password   sshAuth = "password"
	agent      sshAuth = "agent"
	privateKey sshAuth = "privateKey"
)

var sshAuths = []sshAuth{password, agent, privateKey}

// ssh auth type conversion
func (a sshAuth) New() (sshAuth, error) {
	if !slices.Contains(sshAuths, a) {
		log.Printf("string %s could not be converted to sshAuth enum", a)
		return "", errors.New("invalid sshAuth enum")
	}
	return a, nil
}

// connection type with pseudo-enum
type connectionType string

const (
	ssh    connectionType = "ssh"
	winrm  connectionType = "winrm"
	docker connectionType = "docker"
	podman connectionType = "podman"
	lxc    connectionType = "lxc"
)

var connectionTypes = []connectionType{ssh, winrm, docker, podman, lxc}

// connection type conversion
func (a connectionType) New() (connectionType, error) {
	if !slices.Contains(connectionTypes, a) {
		log.Printf("string %s could not be converted to connection enum", a)
		return "", errors.New("invalid connection enum")
	}
	return a, nil
}
