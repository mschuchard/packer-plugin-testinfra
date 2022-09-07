packer {
  required_version = "~> 1.8.0"

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
  ssh_username         = "vagrant"
  vm_name              = "fixtures_default_1662571353037_70563"
}

build {
  sources = ["source.docker.ubuntu", "source.virtualbox-vm.ubuntu"]

  provisioner "testinfra" {
    pytest_path = "/usr/local/bin/py.test"
    test_file   = "${path.cwd}/test.py"
  }
}
