### 1.3.0 (Next)
- Support changing into other directory for test execution.

### 1.2.3
- Only allow `processes` input parameter if `pytest-xdist` installed.
- Switch paramiko connection backend to pure ssh.
- Completely revise remote communication integration.

### 1.2.2
- Logging updates, code optimization, go version increments, and dependency updates.

### 1.2.1
- Pass marker and keyword expressions as singular strings to command executors.
- Reorganize module into packages.
- Add testinfra installation verification.

### 1.2.0
- Support executing tests by keyword substring expression.
- Allow empty `test_files` for default PyTest behavior.
- Temporary Packer instance local execution support (beta).
- Intrinsic support for Testinfra installation on Packer instance (beta).

### 1.1.1
- Improve WinRM password determination.
- Ignore empty SSH private key in SSH backend arguments.
- Utilize SSH Password for authentication when provided.
- Prefer SSHPort and SSHUsername Data when communicator is SSH.

### 1.1.0
- Support multiple TestInfra `test_files`.
- Support parallel test execution.
- Support executing tests with sudo.
- Support executing tests by marker expression.
- Testinfra Podman and LXC backend support (beta).

### 1.0.1
- Testinfra WinRM backend support (beta).
- Testinfra SSH backend improved support.
- Prefer generated Host and Port over generated httpAddr.

### 1.0.0
- Initial release.
