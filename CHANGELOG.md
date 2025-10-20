### 1.5.2 (Next)
- Note in Packer logs when they originate from this plugin.
- Promote `winrm` to officially supported communicator.
- Minor code optimization.

### 1.5.1
- More improvements to Packer communicators' settings interfacing with Testinfra provisioner plugin.
- Improve SSH temporary file cleanup during errors.
- Improve Testinfra command handling.
- Improve path validation.
- Workaround Packer type change for `port` data in new version.

### 1.5.0
- Add `transfer_files` parameter.
- Add `compact` parameter.

### 1.4.2
- Significantly improve Packer SSH and WinRM communicators settings' interfacing with provisioner plugin.

### 1.4.1
- Miscellaneous code, build, and logging improvements

### 1.4.0
- Support multiple levels of verbose (argument bool to int).
- Update parallel execution behavior (argument name change and int to bool).

### 1.3.1
- Improve logging.
- Fix returns for errors during execution command determination.
- Validate `chdir` parameter and fix logging conditional.

### 1.3.0
- Support changing into other directory for test execution.
- Support optional Pytest verbose output instead of former mandatory.
- Support becoming different user after elevated permissions.

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
