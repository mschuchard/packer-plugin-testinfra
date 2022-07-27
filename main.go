package main

import (
  "os"
  "log"

  "github.com/hashicorp/packer-plugin-sdk/plugin"
)

func main() {
  // initialize packer plugin set for testinfra
  packerPluginSet := plugin.NewSet()
  packerPluginSet.RegisterProvisioner(plugin.DEFAULT_NAME, new(TestinfraProvisioner))

  // execute packer plugin for testinfra
  err := packerPluginSet.Run()
  if err != nil {
    log.Fatalf("Packer Provisioner Testinfra failure: %v", err.Error())
    os.Exit(1)
  }
}
