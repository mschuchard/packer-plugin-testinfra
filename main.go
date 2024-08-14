package main

import (
	"log"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
	"github.com/hashicorp/packer-plugin-sdk/version"

	"github.com/mschuchard/packer-plugin-testinfra/provisioner"
)

func main() {
	// initialize plugin version
	pluginVersion := version.InitializePluginVersion("1.4.0", "")

	// initialize packer plugin set for testinfra
	packerPluginSet := plugin.NewSet()
	packerPluginSet.RegisterProvisioner(plugin.DEFAULT_NAME, new(testinfra.Provisioner))
	packerPluginSet.SetVersion(pluginVersion)

	// execute packer plugin for testinfra
	if err := packerPluginSet.Run(); err != nil {
		log.Fatalf("Packer Plugin Testinfra failure: %s", err.Error())
		os.Exit(1)
	}
}
