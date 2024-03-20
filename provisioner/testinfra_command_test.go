package testinfra

import (
	"fmt"
	"strings"
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
		test.Errorf("determineExecCmd function failed to determine execution command for minimal config with local execution: %s", err)
	}
	if execCmd != nil {
		test.Errorf("determineExecCmd function failed to properly determine remote execution command for minimal config with local execution: %s", execCmd.String())
	}
	if localCmd.Command != fmt.Sprintf("%s -v", provisioner.config.PytestPath) {
		test.Errorf("determineExecCmd function failed to properly determine local execution command for minimal config with local execution: %s", localCmd.Command)
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
	if execCmd.String() != fmt.Sprintf("%s -v --hosts=ssh://%s@%s:%d --ssh-identity-file=%s --ssh-extra-args=\"-o StrictHostKeyChecking=no\" -k \"%s\" -m \"%s\" -n %d --sudo %s", provisioner.config.PytestPath, generatedData["User"], generatedData["Host"], generatedData["Port"], generatedData["SSHPrivateKeyFile"], provisioner.config.Keyword, provisioner.config.Marker, provisioner.config.Processes, strings.Join(provisioner.config.TestFiles, "")) {
		test.Errorf("determineExecCmd function failed to properly determine remote execution command for basic config with SSH communicator: %s", execCmd.String())
	}
	if localCmd != nil {
		test.Errorf("determineExecCmd function failed to properly determine empty local execution command for basic config with SSH communicator: %v", localCmd.Command)
	}
}
