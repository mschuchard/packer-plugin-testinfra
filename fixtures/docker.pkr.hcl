packer {
  required_version = "~> 1.7.10"

  required_plugins {
    docker = {
      source  = "github.com/hashicorp/docker"
      version = "~> 1.0.0"
    }
  }
}

source "docker" "ubuntu" {
  discard = true
  image   = "ubuntu:latest"
}

build {
  sources = ["source.docker.ubuntu"]

  provisioner "testinfra" {
    pytest_path = "/usr/local/bin/py.test"
    test_file   = "${path.cwd}/fixtures/test.py"
  }
}
