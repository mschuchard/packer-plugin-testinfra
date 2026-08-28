package testinfra

import (
	"errors"
	"os"
	"slices"
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

func TestRedact(test *testing.T) {
	expected := []string{"--hosts=ssh://user:REDACTED@10.0.0.1", "some", "--hosts=winrm://admin:REDACTED@server", "--some-flag", "value", "--hosts=ssh://user:REDACTED@host.com"}
	redacted := redact([]string{"--hosts=ssh://user:pass@10.0.0.1", "some", "--hosts=winrm://admin:secret@server", "--some-flag", "value", "--hosts=ssh://user:p@ss!word@host.com"})

	if !slices.Equal(redacted, expected) {
		test.Error("input string slice was not redacted correctly")
		test.Errorf("expected: %+q, actual: %+q", expected, redacted)
	}
}
