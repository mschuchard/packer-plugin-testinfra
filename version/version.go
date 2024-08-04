package version

import "github.com/hashicorp/packer-plugin-sdk/version"

// Version is the main version number that is being run at the moment.
const Version = "1.4.0"

// PluginVersion is used by the plugin set to enable Packer to recognize the plugin version.
// hardcode prerelease to empty string
var PluginVersion = version.InitializePluginVersion(Version, "")
