package main

import (
	"log"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
	"github.com/hashicorp/packer-plugin-sdk/version"

	testinfra "github.com/mschuchard/packer-plugin-testinfra/provisioner"
)

func main() {
	// initialize packer plugin set for testinfra
	packerPluginSet := plugin.NewSet()
	// register plugin provisioner
	packerPluginSet.RegisterProvisioner(plugin.DEFAULT_NAME, new(testinfra.Provisioner))
	// set plugin version
	pluginVersion := version.NewPluginVersion("1.5.2", "", "")
	packerPluginSet.SetVersion(pluginVersion)

	// execute packer plugin for testinfra
	if err := packerPluginSet.Run(); err != nil {
		log.Fatalf("Packer Plugin Testinfra failure: %s", err.Error())
	}
}
