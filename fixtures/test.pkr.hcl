packer {
  required_version = "~> 1.13.0"

  required_plugins {
    docker = {
      source  = "github.com/hashicorp/docker"
      version = "~> 1.1.0"
    }
    /*virtualbox = {
      version = "~> 1.0.0"
      source  = "github.com/hashicorp/virtualbox"
    }*/
  }
}

variable "password" {}

# use local device with null provider
source "null" "vbox" {
  ssh_host       = "127.0.0.1"
  ssh_port       = "10022"
  ssh_username   = "vagrant"
  ssh_password   = var.password
}

# use ubuntu:latest docker image with docker provider
source "docker" "ubuntu" {
  discard = true
  image   = "ubuntu:latest"
}

# test remote execution docker and ssh communicators
build {
  sources = ["source.docker.ubuntu", "source.null.vbox"]

  provisioner "testinfra" {
    pytest_path = "/usr/local/bin/py.test"
    test_files  = ["${path.root}/fixtures/test.py", "${path.root}/fixtures/test.py"]
  }
}

# use ubuntu/jammy64 vagrant box with vbox provider
/*source "virtualbox-vm" "ubuntu" {
  guest_additions_mode = "disable"
  skip_export          = true
  ssh_host             = "127.0.0.1"
  ssh_username         = "vagrant"
  ssh_certificate_file = "${path.cwd}/fixtures/.vagrant/machines/default/virtualbox/private_key"
  vm_name              = "fixtures_default_1662571353037_70563"
}

# test local execution
# TODO: https://github.com/hashicorp/packer-plugin-virtualbox/issues/77
build {
  sources = ["source.virtualbox-vm.ubuntu"]

  provisioner "testinfra" {
    local       = true
    install_cmd = "pip install testinfra"
    # TODO: need file transfer feature in 1.3.0 plugin version
    test_files  = ["${path.cwd}/fixtures/test.py", "${path.cwd}/fixtures/test.py"]
  }
}*/
