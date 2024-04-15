package testinfra

import (
	"fmt"
	"slices"
	"testing"
)

// test provisioner determineExecCmd properly determines execution command
func TestProvisionerDetermineExecCmd(test *testing.T) {
	// test minimal config with local execution
	var provisioner = &Provisioner{
		config: Config{
			PytestPath: "/usr/local/bin/py.test",
			InstallCmd: []string{"pip", "install", "pytest-testinfra"},
			Local:      true,
		},
	}

	execCmd, localCmd, err := provisioner.determineExecCmd()
	if err != nil {
		test.Errorf("determineExecCmd function failed to determine execution commands for local execution minimal config: %v", err)
	}
	if execCmd != nil {
		test.Errorf("determineExecCmd function failed to properly determine remote execution command for local execution minimal config: %s", execCmd.String())
	}
	if localCmd.Command != provisioner.config.PytestPath {
		test.Errorf("determineExecCmd function failed to properly determine local execution command for local execution minimal config: %s", localCmd.Command)
		test.Error(provisioner.config.PytestPath)
	}

	// test basic config with ssh generated data
	provisioner = &Provisioner{
		config: *basicConfig,
	}

	generatedData := map[string]interface{}{
		"ConnType":          "ssh",
		"User":              "me",
		"SSHPrivateKeyFile": "/path/to/sshprivatekeyfile",
		"SSHAgentAuth":      true,
		"Host":              "192.168.0.1",
		"Port":              int64(22),
		"PackerHTTPAddr":    "192.168.0.1:8200",
		"ID":                "1234567890",
	}

	provisioner.generatedData = generatedData

	execCmd, localCmd, err = provisioner.determineExecCmd()
	if err != nil {
		test.Errorf("determineExecCmd function failed to determine execution command for basic config with SSH communicator: %s", err)
	}
	if execCmd.Dir != basicConfig.Chdir {
		test.Error("determineExecCmd function failed to determine execution directory for basic config")
		test.Errorf("actual: %s, expected: %s", execCmd.Dir, basicConfig.Chdir)
	}
	// 1.22 slices.Concat( , provisioner.config.TestFiles)
	if !slices.Equal(execCmd.Args, append([]string{provisioner.config.PytestPath, fmt.Sprintf("--hosts=ssh://%s@%s:%d", generatedData["User"], generatedData["Host"], generatedData["Port"]), fmt.Sprintf("--ssh-identity-file=%s", generatedData["SSHPrivateKeyFile"]), "--ssh-extra-args=\"-o StrictHostKeyChecking=no\"", "-k", fmt.Sprintf("\"%s\"", provisioner.config.Keyword), "-m", fmt.Sprintf("\"%s\"", provisioner.config.Marker), "-n", fmt.Sprint(provisioner.config.Processes), "--sudo", fmt.Sprintf("\"--sudo-user=%s\"", provisioner.config.SudoUser), "-v"}, provisioner.config.TestFiles...)) {
		test.Errorf("determineExecCmd function failed to properly determine remote execution command for basic config with SSH communicator: %s", execCmd.String())
	}
	if localCmd != nil {
		test.Errorf("determineExecCmd function failed to properly determine empty local execution command for basic config with SSH communicator: %v", localCmd.Command)
	}
}
