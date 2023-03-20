package main

import (
  "os"
  "log"

  "github.com/hashicorp/packer-plugin-sdk/plugin"

  "github.com/mschuchard/packer-plugin-testinfra/version"
  "github.com/mschuchard/packer-plugin-testinfra/provisioner"
)

func main() {
  // initialize packer plugin set for testinfra
  packerPluginSet := plugin.NewSet()
  packerPluginSet.RegisterProvisioner(plugin.DEFAULT_NAME, new(testinfra.Provisioner))
  packerPluginSet.SetVersion(version.PluginVersion)

  // execute packer plugin for testinfra
  err := packerPluginSet.Run()
  if err != nil {
    log.Fatalf("Packer Plugin Testinfra failure: %v", err.Error())
    os.Exit(1)
  }
}
