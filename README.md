# Packer Plugin Testinfra
The Packer plugin for [Testinfra](https://testinfra.readthedocs.io) (currently and probably also in the future only the provisioner component implemented) is used with [Packer](https://www.packer.io) for validating custom machine images.

## Installation

This plugin requires Packer version `>= 1.7.0` due to the modern SDK usage. A simple Packer config located in the same directory as your other templates and configs utilizing this plugin can automatically manage the plugin:

```hcl
packer {
  required_plugins {
    testinfra = {
      version = "~> 1.5.0"
      source  = "github.com/mschuchard/testinfra"
    }
  }
}
```

Afterwards, `packer init` can automatically manage your plugin as per normal. Note that this plugin currently does not manage your local device's Testinfra installation, and you will need to install that on your local device as a prerequisite for this plugin to function correctly (if not executing local to the instance). If you are using this plugin with `local` enabled, then there is some assistance with managing Testinfra on the temporary Packer instance with the `install_cmd` parameter. The minimum supported version of Testinfra is `6.7.0`. The minimum supported version of Pytest is unknown, but it would be recommended to install a version `>= 7.0.0`.

## Usage

### Basic Example

A basic example for usage of this plugin with selected optional arguments follows below:

```hcl
build {
  provisioner "testinfra" {
    pytest_path = "/usr/local/bin/py.test"
    sudo        = false
    test_files  = ["${path.root}/test.py", "${path.root}/test_other.py"]
    verbose     = 2
  }
}
```

### Arguments

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| **chdir** | Change into this directory before executing `pytest`. Unsupported with `local` test execution. | string | `cwd` | no |
| **compact** | Whether to report in compact form (no header, summary, or warnings). | bool | false | no |
| **destination_dir** | Whether to transfer the `test_files` to the temporary Packer instance used for building the machine image artifact at input value location. Presence of this directory cannot be validated prior to execution. Ignored unless `local` is `true`. The `file` provisioner should normally be preferred instead of this parameter, and this should also be considered a beta feature. | string | "" | no |
| **install_cmd** | Command to execute on the instance used for building the machine image artifact; can be used to e.g. install and configure Testinfra prior to a `local` test execution. Ignored unless `local` is `true`. | list(string) | [] | no |
| **keyword** | PyTest keyword substring expression for selective test execution. | string | "" | no |
| **local** | Execute Testinfra tests locally on the instance used for building the machine image artifact. Most plugin validation is skipped with this option. | bool | false | no |
| **marker** | PyTest marker expression for selective test execution. | string | "" | no |
| **parallel** | Whether to execute the Testinfra tests in parallel across the available physical CPUs. This parameter requires installation of the [pytest-xdist](https://pypi.org/project/pytest-xdist) plugin. | bool | false | no |
| **pytest_path** | The path to the installed `py.test` executable for initiating the Testinfra tests. | string | "py.test" | no |
| **sudo** | Whether or not to execute the tests with `sudo` elevated permissions. | bool | false | no |
| **sudo_user** | User to become when executing the tests. Mutually exclusive with `sudo`, and therefore ignored when `sudo` is input as `true`. | string | "" | no |
| **test_files** | The paths to the files containing the Testinfra tests for execution and validation of the machine image artifact. The default empty value will execute default PyTest behavior of all test files prefixed with `test_` recursively discovered from the current working directory. | list(string) | [] | no |
| **verbose** | The level of Pytest verbose enabled (value corresponds to the number of `v` flags). Maximum value is `4`. | number | 0 | no |

### Communicators

This plugin currently supports the `ssh`, `docker`, `lxc`, and `podman` communicator types. It also supports the `winrm` communicator type, and execution local to the instance used for building the machine image artifact, as beta features. Please ensure that at least one communication type is enabled for the built image (this is also generally a requirement for Packer itself).

## Contributing
Code should pass all unit and acceptance tests. New features should involve new unit tests.

Please consult the GitHub Project for the current development roadmap.
