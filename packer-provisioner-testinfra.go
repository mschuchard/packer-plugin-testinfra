//go:generate packer-sdc mapstructure-to-hcl2 -type TestinfraConfig
package main

import (
  "log"
  "context"

  "github.com/hashicorp/hcl/v2/hcldec"
  "github.com/hashicorp/packer-plugin-sdk/packer"
  "github.com/hashicorp/packer-plugin-sdk/template/config"
  "github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

// config data from packer template/config
type TestinfraConfig struct {
  TestFile string `mapstructure:"test_file"`

  ctx interpolate.Context
}

// implements the packer.Provisioner interface
type TestinfraProvisioner struct{
  config TestinfraConfig
}

// implements configspec with hcl2spec helper function
func (provisioner *TestinfraProvisioner) ConfigSpec() hcldec.ObjectSpec {
  return provisioner.config.FlatMapstructure().HCL2Spec()
}


func (provisioner *TestinfraProvisioner) Prepare(raws ...interface{}) error {
  // parse testinfra provisioner config
  err := config.Decode(&provisioner.config, &config.DecodeOpts{
    Interpolate:        true,
    InterpolateContext: &provisioner.config.ctx,
  }, raws...)
  if err != nil {
    return err
  }

  return nil
}

func (provisioner *TestinfraProvisioner) Provision(ctx context.Context, ui packer.Ui, comm packer.Communicator, generatedData map[string]interface{}) error {
  // parse generated data for context
  provisioner.config.ctx.Data = generatedData
  testFile, err := interpolate.Render(provisioner.config.TestFile, &provisioner.config.ctx)
  if err != nil {
    log.Fatalf("Error parsing config for TestFile: %v", err.Error())
    return err
  }

  log.Printf("Testinfra file is: %v", testFile)

  return nil
}
