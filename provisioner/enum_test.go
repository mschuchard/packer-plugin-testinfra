package testinfra

import (
	"testing"
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

func TestConnectionNew(test *testing.T) {
	connectionTest, err := connectionType("ssh").New()
	if err != nil {
		test.Error(err)
	}
	if connectionTest != ssh {
		test.Error("connection did not type convert correctly")
		test.Errorf("expected: ssh, actual: %s", connectionTest)
	}

	if _, err = connectionType("foo").New(); err == nil || err.Error() != "invalid connection enum" {
		test.Error("connection type conversion did not error expectedly")
		test.Errorf("expected: invalid connection enum, actual: %s", err)
	}
}
