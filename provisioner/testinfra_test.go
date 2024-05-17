package testinfra

import (
	"errors"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/packer"
)

// ci boolean; true if either circle or gh actions env var is 'true'
var CI bool = os.Getenv("CIRCLECI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true"

// global helper vars for tests
var basicConfig = &Config{
	Chdir:      "/tmp",
	Keyword:    "not slow",
	Marker:     "fast",
	Processes:  4,
	PytestPath: "/usr/local/bin/py.test",
	Sudo:       true,
	SudoUser:   "fooman",
	TestFiles:  []string{"fixtures/test.py"},
	Verbose:    true,
}

// test basic config for packer template/config data
func TestProvisionerConfig(test *testing.T) {
	var provisioner = &Provisioner{
		config: *basicConfig,
	}

	if provisioner.config.PytestPath != basicConfig.PytestPath || !slices.Equal(provisioner.config.TestFiles, basicConfig.TestFiles) || provisioner.config.Chdir != basicConfig.Chdir || provisioner.config.Keyword != basicConfig.Keyword || provisioner.config.Marker != basicConfig.Marker || provisioner.config.Processes != basicConfig.Processes || provisioner.config.Sudo != basicConfig.Sudo || provisioner.config.SudoUser != basicConfig.SudoUser || provisioner.config.Verbose != basicConfig.Verbose {
		test.Errorf("provisioner config struct not initialized correctly")
	}
}

// test struct for provisioner interface
func TestProvisionerInterface(test *testing.T) {
	var raw interface{} = &Provisioner{}
	if _, ok := raw.(packer.Provisioner); !ok {
		test.Errorf("Testinfra config struct must be a Provisioner")
	}
}

// test provisioner prepare with basic config
func TestProvisionerPrepareBasic(test *testing.T) {
	var provisioner Provisioner

	if !CI {
		err := provisioner.Prepare(basicConfig)
		if err != nil {
			test.Errorf("prepare function failed with basic Testinfra Packer config")
		}
	}
}

// test provisioner prepare with minimal config (essentially default setting)
func TestProvisionerPrepareMinimal(test *testing.T) {
	var provisioner Provisioner

	err := provisioner.Prepare(&Config{})
	if err != nil && !CI {
		test.Errorf("prepare function failed with minimal Testinfra Packer config")
	}

	if len(provisioner.config.Chdir) > 0 {
		test.Errorf("default empty setting for Chdir is incorrect: %s", provisioner.config.Chdir)
	}

	if len(provisioner.config.InstallCmd) > 0 {
		test.Errorf("default empty setting for InstallCmd is incorrect: %s", provisioner.config.InstallCmd)
	}

	if len(provisioner.config.Keyword) > 0 {
		test.Errorf("default empty setting for Keyword is incorrect: %s", provisioner.config.Keyword)
	}

	if provisioner.config.Local != false {
		test.Errorf("default false setting for Local is incorrect: %t", provisioner.config.Local)
	}

	if len(provisioner.config.Marker) > 0 {
		test.Errorf("default empty setting for Marker is incorrect: %s", provisioner.config.Marker)
	}

	if provisioner.config.Processes != 0 {
		test.Errorf("default empty setting for Processes is incorrect: %d", provisioner.config.Processes)
	}

	if provisioner.config.Sudo == true {
		test.Errorf("default false setting for Sudo is incorrect: %t", provisioner.config.Sudo)
	}

	if len(provisioner.config.SudoUser) > 0 {
		test.Errorf("default empty setting for SudoUser is incorrect: %s", provisioner.config.SudoUser)
	}

	if provisioner.config.Verbose == true {
		test.Errorf("default false setting for Verbose is incorrect: %t", provisioner.config.Verbose)
	}

	if provisioner.config.PytestPath != "py.test" {
		test.Errorf("default setting for PytestPath is incorrect: %s", provisioner.config.PytestPath)
	}

	if len(provisioner.config.TestFiles) != 0 {
		test.Errorf("default setting for TestFiles is incorrect: %+q", provisioner.config.TestFiles)
	}
}

// test provisioner prepare defaults to empty slice for test files
func TestProvisionerPrepareEmptyTestFile(test *testing.T) {
	var provisioner Provisioner
	var emptyTestFileConfig = &Config{
		PytestPath: "/usr/local/bin/py.test",
	}

	if !CI {
		err := provisioner.Prepare(emptyTestFileConfig)
		if err != nil {
			test.Error("prepare function failed with no test_files minimal Testinfra Packer config")
		}

		if len(provisioner.config.TestFiles) > 0 {
			test.Errorf("default setting for TestFiles is incorrect: %s", strings.Join(provisioner.config.TestFiles, ""))
		}
	}
}

// test provisioner prepare errors on nonexistent files
func TestProvisionerPrepareNonExistFiles(test *testing.T) {
	var provisioner Provisioner

	// test no pytest
	var noPytestConfig = &Config{
		PytestPath: "/home/foo/py.test",
		TestFiles:  []string{"fixtures/test.py"},
	}

	err := provisioner.Prepare(noPytestConfig)
	if !(errors.Is(err, os.ErrNotExist)) {
		test.Error("prepare function did not fail correctly on nonexistent pytest")
		test.Error(err)
	}

	// test nonexistent testfile
	var noTestFileConfig = &Config{
		PytestPath: "/usr/local/bin/py.test",
		TestFiles:  []string{"fixtures/test.py", "/home/foo/test.py"},
	}

	if !CI {
		err = provisioner.Prepare(noTestFileConfig)
		if !(errors.Is(err, os.ErrNotExist)) {
			test.Error("prepare function did not fail correctly on nonexistent testfile")
			test.Error(err)
		}
	}
}

// test provisioner prepare errors on processes with no xdist
func TestProvisionerPrepareNoXdist(test *testing.T) {
	var provisioner Provisioner

	// test no xdist with global basic config
	if err := provisioner.Prepare(basicConfig); err != nil && !CI {
		test.Error("prepare function failed basic config")
	}
	if CI && provisioner.config.Processes != basicConfig.Processes {
		test.Error("prepare function reverted processes member value to default 0 after determing xdist is installed")
		test.Errorf("actual value: %d", provisioner.config.Processes)
	}
	if !CI && provisioner.config.Processes != 0 {
		test.Error("prepare function did not revert processes member value to default 0 after determing xdist is not installed")
		test.Errorf("actual value: %d", provisioner.config.Processes)
	}
}
