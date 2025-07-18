package testinfra

import (
	"errors"
	"os"
	"slices"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/packer"
)

// ci boolean; true if either circle or gh actions env var is 'true'
var CI bool = os.Getenv("CIRCLECI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true"

// global helper vars for tests
var basicConfig = &Config{
	Chdir:          "/tmp",
	Compact:        true,
	InstallCmd:     []string{"/bin/false"},
	Keyword:        "not slow",
	Local:          false,
	Marker:         "fast",
	Parallel:       true,
	PytestPath:     "/usr/local/bin/py.test",
	Sudo:           true,
	SudoUser:       "fooman",
	DestinationDir: "/tmp",
	TestFiles:      []string{"../fixtures/test.py"},
	Verbose:        2,
}

// test basic config for packer template/config data
func TestProvisionerConfig(test *testing.T) {
	var provisioner = &Provisioner{
		config: *basicConfig,
	}

	if provisioner.config.PytestPath != basicConfig.PytestPath || provisioner.config.DestinationDir != basicConfig.DestinationDir || !slices.Equal(provisioner.config.TestFiles, basicConfig.TestFiles) || provisioner.config.Chdir != basicConfig.Chdir || provisioner.config.Compact != basicConfig.Compact || !slices.Equal(provisioner.config.InstallCmd, basicConfig.InstallCmd) || provisioner.config.Keyword != basicConfig.Keyword || provisioner.config.Local != basicConfig.Local || provisioner.config.Marker != basicConfig.Marker || provisioner.config.Parallel != basicConfig.Parallel || provisioner.config.Sudo != basicConfig.Sudo || provisioner.config.SudoUser != basicConfig.SudoUser || provisioner.config.Verbose != basicConfig.Verbose {
		test.Errorf("provisioner config struct not initialized correctly")
	}
}

// test struct for provisioner interface
func TestProvisionerInterface(test *testing.T) {
	var raw any = &Provisioner{}
	if _, ok := raw.(packer.Provisioner); !ok {
		test.Errorf("Testinfra config struct must be a Provisioner")
	}
}

// test provisioner prepare with basic config
func TestProvisionerPrepareBasic(test *testing.T) {
	var provisioner Provisioner

	if err := provisioner.Prepare(basicConfig); err != nil {
		test.Errorf("prepare function failed with basic Testinfra Packer config")
		test.Error(err)
	}
}

// test provisioner prepare with minimal config (essentially default settings)
func TestProvisionerPrepareMinimal(test *testing.T) {
	var provisioner Provisioner

	if err := provisioner.Prepare(&Config{}); err != nil {
		test.Errorf("prepare function failed with minimal Testinfra Packer config")
		test.Error(err)
	}

	if len(provisioner.config.Chdir) > 0 {
		test.Errorf("default empty setting for Chdir is incorrect: %s", provisioner.config.Chdir)
	}

	if provisioner.config.Compact {
		test.Errorf("default false setting for compact is incorrect: %t", provisioner.config.Compact)
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

	if provisioner.config.Parallel {
		test.Errorf("default false setting for Parallel is incorrect: %t", provisioner.config.Parallel)
	}

	if provisioner.config.Sudo {
		test.Errorf("default false setting for Sudo is incorrect: %t", provisioner.config.Sudo)
	}

	if len(provisioner.config.SudoUser) > 0 {
		test.Errorf("default empty setting for SudoUser is incorrect: %s", provisioner.config.SudoUser)
	}

	if provisioner.config.Verbose != 0 {
		test.Errorf("default empty setting for Verbose is incorrect: %d", provisioner.config.Verbose)
	}

	if provisioner.config.PytestPath != "py.test" {
		test.Errorf("default setting for PytestPath is incorrect: %s", provisioner.config.PytestPath)
	}

	if len(provisioner.config.TestFiles) > 0 {
		test.Errorf("default empty setting for TestFiles is incorrect: %+q", provisioner.config.TestFiles)
	}

	if len(provisioner.config.DestinationDir) > 0 {
		test.Errorf("default empty setting for DestinationDir is incorrect: %s", provisioner.config.DestinationDir)
	}
}

// test provisioner errors on nonexistent chdir
func TestProvisionerPrepareNonExistChdir(test *testing.T) {
	var provisioner Provisioner
	var noChdirConfig = &Config{
		Chdir: "/tmp/no_one_here",
	}

	if !CI {
		if err := provisioner.Prepare(noChdirConfig); err == nil || !errors.Is(err, os.ErrNotExist) {
			test.Error("prepare function did not fail correctly on nonexistent chdir")
			test.Error(err)
		}
	}
}

// test provisioner prepare errors on nonexistent files
func TestProvisionerPrepareNonExistFiles(test *testing.T) {
	var provisioner Provisioner

	// test no pytest
	var noPytestConfig = &Config{
		PytestPath: "/home/foo/py.test",
		TestFiles:  []string{"../fixtures/test.py"},
	}

	if err := provisioner.Prepare(noPytestConfig); err == nil || !(errors.Is(err, os.ErrNotExist)) {
		test.Error("prepare function did not fail correctly on nonexistent pytest")
		test.Error(err)
	}

	// test nonexistent testfile
	var noTestFileConfig = &Config{
		PytestPath: "/usr/local/bin/py.test",
		TestFiles:  []string{"../fixtures/test.py", "/home/foo/test.py"},
	}

	if err := provisioner.Prepare(noTestFileConfig); err == nil || !(errors.Is(err, os.ErrNotExist)) {
		test.Error("prepare function did not fail correctly on nonexistent testfile")
		test.Error(err)
	}
}

// test provisioner prepare reverts value on processes with no xdist
func TestProvisionerPrepareNoXdist(test *testing.T) {
	var provisioner Provisioner

	// test no xdist with global basic config
	if err := provisioner.Prepare(basicConfig); err != nil {
		test.Error("prepare function failed basic config")
	}
	if provisioner.config.Parallel {
		test.Error("prepare function did not revert parallel member value to default 'false' after determing xdist is not installed")
		test.Errorf("actual value: %t", provisioner.config.Parallel)
	}
}
