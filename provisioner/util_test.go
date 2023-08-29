package testinfra

import (
	"errors"
	"os"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/packer"
)

func TestProvisionerUploadFiles(test *testing.T) {
	var provisioner Provisioner
	comm := new(packer.MockCommunicator)

	if emptyReturn := provisioner.uploadFiles(comm, []string{}, "/tmp/"); emptyReturn != nil {
		test.Errorf("uploadFiles with empty files slice did not return no error, but instead %s was returned", emptyReturn)
	}

	err := provisioner.uploadFiles(comm, []string{"foobar"}, "/tmp/")
	if !errors.Is(err, os.ErrNotExist) {
		test.Errorf("expected nonexistent file to return ErrNotExist error, but instead %s was returned", err)
	}
}
