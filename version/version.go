package version

import "github.com/hashicorp/packer-plugin-sdk/version"

const (
	// Version is the main version number that is being run at the moment.
	Version = "1.3.0"

	// VersionPrerelease is A pre-release marker for the Version. If this is ""
	// (empty string) then it means that it is a final release. Otherwise, this
	// is a pre-release such as "dev" (in development), "beta", "rc1", etc.
	VersionPrerelease = ""
)

// PluginVersion is used by the plugin set to enable Packer to recognize the plugin version.
var PluginVersion = version.InitializePluginVersion(Version, VersionPrerelease)
