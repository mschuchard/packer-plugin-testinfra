package testinfra

import (
	"errors"
	"os"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/packer"
)

func TestProvisionerUploadFiles(test *testing.T) {
	comm := new(packer.MockCommunicator)

	err := uploadFiles(comm, []string{}, "/tmp/")
	if err != nil {
		test.Errorf("uploadFiles with empty files slice input returned error")
		test.Error(err)
	}

	err = uploadFiles(comm, []string{"foobar"}, "/tmp/")
	if !errors.Is(err, os.ErrNotExist) {
		test.Errorf("expected nonexistent file to return ErrNotExist error, but instead %s was returned", err)
	}
}
