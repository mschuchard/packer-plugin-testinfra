package testinfra

import (
	"fmt"
	"slices"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/packer"
)

// test provisioner determineExecCmd properly determines execution command
func TestProvisionerDetermineExecCmd(test *testing.T) {
	// initialize simple test ui
	ui := packer.TestUi(test)

	// test minimal config with local execution
	var provisioner = &Provisioner{
		config: Config{
			PytestPath: "/usr/local/bin/py.test",
			InstallCmd: []string{"pip", "install", "pytest-testinfra"},
			Local:      true,
		},
	}

	execCmd, localCmd, err := provisioner.determineExecCmd(ui)
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

	provisioner.generatedData = map[string]any{
		"ConnType":          "ssh",
		"User":              "me",
		"SSHPrivateKeyFile": "/path/to/sshprivatekeyfile",
		"SSHAgentAuth":      true,
		"Host":              "192.168.0.1",
		"Port":              22,
		"PackerHTTPAddr":    "192.168.0.1:8200",
		"ID":                "1234567890",
	}

	execCmd, localCmd, err = provisioner.determineExecCmd(ui)
	if err != nil {
		test.Errorf("determineExecCmd function failed to determine execution command for basic config with SSH communicator: %s", err)
	}
	if execCmd.Dir != basicConfig.Chdir {
		test.Error("determineExecCmd function failed to determine execution directory for basic config")
		test.Errorf("actual: %s, expected: %s", execCmd.Dir, basicConfig.Chdir)
	}
	if !slices.Equal(execCmd.Args, slices.Concat([]string{provisioner.config.PytestPath, fmt.Sprintf("--hosts=ssh://%s@%s:%d", provisioner.generatedData["User"], provisioner.generatedData["Host"], provisioner.generatedData["Port"]), fmt.Sprintf("--ssh-identity-file=%s", provisioner.generatedData["SSHPrivateKeyFile"]), "--ssh-extra-args=\"-o StrictHostKeyChecking=no\"", "--no-header", "--no-summary", "--disable-warnings", "-k", fmt.Sprintf("\"%s\"", provisioner.config.Keyword), "-m", fmt.Sprintf("\"%s\"", provisioner.config.Marker), "-n", "auto", "--sudo", "-vv"}, provisioner.config.TestFiles)) {
		test.Errorf("determineExecCmd function failed to properly determine remote execution command for basic config with SSH communicator: %s", execCmd.String())
	}
	if localCmd != nil {
		test.Errorf("determineExecCmd function failed to properly determine empty local execution command for basic config with SSH communicator: %v", localCmd.Command)
	}
}
