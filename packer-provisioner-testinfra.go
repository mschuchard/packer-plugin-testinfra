//go:generate packer-sdc mapstructure-to-hcl2 -type TestinfraConfig
package main

import (
  "os"
  "context"
  "log"

  "github.com/hashicorp/hcl/v2/hcldec"
  "github.com/hashicorp/packer-plugin-sdk/plugin"
  "github.com/hashicorp/packer-plugin-sdk/packer"
)

// config data from packer template/config
type TestinfraConfig struct {}

// implements the packer.Provisioner interface
type TestinfraProvisioner struct{
  config TestinfraConfig
}

func main() {
  // initialize packer plugin set for testinfra
  packerPluginSet := plugin.NewSet()
  packerPluginSet.RegisterProvisioner(plugin.DEFAULT_NAME, new(TestinfraProvisioner))

  // execute packer plugin for testinfra
  err := packerPluginSet.Run()
  if err != nil {
    log.Fatalf("Packer Provisioner Testinfra failure: %v", err.Error())
  }
}

func (provisioner *TestinfraProvisioner) ConfigSpec() hcldec.ObjectSpec {
  return provisioner.config.FlatMapstructure().HCL2Spec()
}


func (provisioner *TestinfraProvisioner) Prepare(...interface{}) error {
  return nil
}

func (provisioner *TestinfraProvisioner) Provision(ctx context.Context, ui packer.Ui, comm packer.Communicator, data map[string]interface{}) error {
  return nil
}
