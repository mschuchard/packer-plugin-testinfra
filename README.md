# Packer Plugin Testinfra
The Packer plugin for [Testinfra](https://testinfra.readthedocs.io) (currently and probably also in the future only the provisioner component implemented) is used with [Packer](https://www.packer.io) for validating custom machine images.

## Installation

This plugin requires Packer version >= 1.7.0 due to the modern SDK usage. A simple Packer config located in the same directory as your other templates and configs utilizing this plugin can automatically manage the plugin:

```hcl
packer {
  required_plugins {
    testinfra = {
      version = "~> 1.1.0"
      source  = "github.com/mschuchard/testinfra"
    }
  }
}
```

Afterwards, `packer init` can automatically manage your plugin as per normal. Note that this plugin currently does not manage your Testinfra installation, and you will need to install that as a prerequisite for this plguin to function correctly.

## Usage

### Basic Example

A basic example for usage of this plugin follows below:

```hcl
build {
  provisioner "testinfra" {
    processes   = 2
    pytest_path = "/usr/local/bin/py.test"
    sudo        = false
    test_files  = ["${path.root}/test.py"]
  }
}
```

### Arguments

- **processes**: The number of parallel processes for Testinfra test execution.  
default: `null`
- **pytest_path**: The path to the installed `py.test` executable for initiating the Testinfra tests.  
default: `py.test`
- **sudo**: Whether or not to execute the tests with `sudo` elevated permissions.  
default: false
- **test_files**: The paths to the files containing the Testinfra tests for execution and validation of the machine image artifact.  
default: `null`

### Communicators

This plugin currently supports the `ssh` and `docker` communicator types. It also supports the `winrm`, `lxc`, and `podman` communicator types as a beta feature. Please ensure that SSH or WinRM is enabled for the built image if it is not a Docker, LXC, or Podman image.

## Contributing
Code should pass all unit and acceptance tests. New features should involve new unit tests.

Please consult the GitHub Project for the current development roadmap.
