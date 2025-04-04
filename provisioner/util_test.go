package testinfra

import (
	"errors"
	"os"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/packer"
)

func TestSSHAuthNew(test *testing.T) {
	sshAuthTest, err := sshAuth("password").New()
	if err != nil {
		test.Error(err)
	}
	if sshAuthTest != password {
		test.Error("sshauth did not type convert correctly")
		test.Errorf("expected: password, actual: %s", sshAuthTest)
	}

	if _, err = sshAuth("foo").New(); err == nil || err.Error() != "invalid sshAuth enum" {
		test.Error("sshauth type conversion did not error expectedly")
		test.Errorf("expected: invalid sshAuth enum, actual: %s", err)
	}
}

func TestProvisionerUploadFiles(test *testing.T) {
	comm := &packer.MockCommunicator{}

	err := uploadFiles(comm, []string{"../.gitignore"}, "/dafdfsad")
	if err != nil {
		test.Errorf("generic inputs returned error: %s", err)
	}

	err = uploadFiles(comm, []string{"foobar"}, "/tmp")
	if !errors.Is(err, os.ErrNotExist) {
		test.Errorf("expected nonexistent file to return ErrNotExist error, but instead %s was returned", err)
	}
}
