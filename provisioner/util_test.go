package testinfra

import (
	"errors"
	"os"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/packer"
)

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
