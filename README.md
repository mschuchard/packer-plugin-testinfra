# Packer Plugin Testinfra
The Packer plugin for [Testinfra](https://testinfra.readthedocs.io) (currently and probably also in the future only the provisioner component implemented) is used with [Packer](https://www.packer.io) for validating custom machine images.

## Installation

This plugin requires Packer version `>= 1.7.0` due to the modern SDK usage. A simple Packer config located in the same directory as your other templates and configs utilizing this plugin can automatically manage the plugin:

```hcl
packer {
  required_plugins {
    testinfra = {
      version = "~> 1.2.0"
      source  = "github.com/mschuchard/testinfra"
    }
  }
}
```

Afterwards, `packer init` can automatically manage your plugin as per normal. Note that this plugin currently does not manage your local Testinfra installation, and you will need to install that on your local device as a prerequisite for this plugin to function correctly. If you are using this plugin with `local` enabled, then there is some assistance with managing Testinfra on the temporary Packer instance with `install_cmd`. The minimum supported version of Testinfra is `6.7.0`, but a lower version may be possible if you are not using the SSH communicator. The minimum supported version of Pytest is unknown, but it would be recommended to install a version `>= 7.0.0`.

## Usage

### Basic Example

A basic example for usage of this plugin with common optional arguments follows below:

```hcl
build {
  provisioner "testinfra" {
    processes   = 2
    pytest_path = "/usr/local/bin/py.test"
    sudo        = false
    test_files  = ["${path.root}/test.py", "${path.root}/test_other.py"]
  }
}
```

### Arguments

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| **install_cmd** | Command to execute on the instance used for building the machine image artifact; can be used to install and configure Testinfra prior to a `local` test execution. | list(string) | [] | no |
| **keyword** | PyTest keyword substring expression for selective test execution. | string | "" | no |
| **local** | Execute Testinfra tests locally on the instance used for building the machine image artifact. Most plugin validation is skipped with this option. | bool | false | no |
| **marker** | PyTest marker expression for selective test execution. | string | "" | no |
| **processes** | The number of parallel processes for Testinfra test execution. This parameter requires installation of the [pytest-xdist](https://pypi.org/project/pytest-xdist/) plugin. | number | 0 | no |
| **pytest_path** | The path to the installed `py.test` executable for initiating the Testinfra tests. | string | "py.test" | no |
| **sudo** | Whether or not to execute the tests with `sudo` elevated permissions. | bool | false | no |
| **test_files** | The paths to the files containing the Testinfra tests for execution and validation of the machine image artifact. The default empty value will execute default PyTest behavior of all test files prefixed with `test_` recursively discovered from the current working directory. | list(string) | [] | no |

### Communicators

This plugin currently supports the `ssh` and `docker` communicator types. It also supports the `winrm`, `lxc`, and `podman` communicator types, and execution local to the instance used for building the machine image artifact, as beta features. Please ensure that SSH or WinRM is enabled for the built image if it is not a Docker, LXC, or Podman image.

## Contributing
Code should pass all unit and acceptance tests. New features should involve new unit tests.

Please consult the GitHub Project for the current development roadmap.
