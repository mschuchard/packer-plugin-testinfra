packer {
  required_version = "~> 1.9.0"

  required_plugins {
    docker = {
      source  = "github.com/hashicorp/docker"
      version = "~> 1.0.0"
    }
    virtualbox = {
      version = "~> 1.0.0"
      source  = "github.com/hashicorp/virtualbox"
    }
  }
}

source "docker" "ubuntu" {
  discard = true
  image   = "ubuntu:latest"
}

# use ubuntu/jammy64 vagrant box with vbox provider
source "virtualbox-vm" "ubuntu" {
  guest_additions_mode = "disable"
  skip_export          = true
  ssh_host             = "127.0.0.1"
  ssh_username         = "vagrant"
  ssh_certificate_file = "${path.cwd}/fixtures/.vagrant/machines/default/virtualbox/private_key"
  vm_name              = "fixtures_default_1662571353037_70563"
}

# test remote execution
build {
  sources = ["source.docker.ubuntu"]

  provisioner "testinfra" {
    pytest_path = "/usr/local/bin/py.test"
    test_files  = ["${path.cwd}/fixtures/test.py", "${path.cwd}/fixtures/test.py"]
  }
}

# test local execution
# TODO: https://github.com/hashicorp/packer-plugin-virtualbox/issues/77
/*build {
  sources = ["source.virtualbox-vm.ubuntu"]

  provisioner "testinfra" {
    local       = true
    install_cmd = "pip install testinfra"
    # TODO: need file transer feature in 1.3.0 plugin version
    test_files  = ["${path.cwd}/fixtures/test.py", "${path.cwd}/fixtures/test.py"]
  }
}*/
