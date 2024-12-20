// Code generated by "packer-sdc mapstructure-to-hcl2"; DO NOT EDIT.

package testinfra

import (
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"
)

// FlatConfig is an auto-generated flat version of Config.
// Where the contents of a field with a `mapstructure:,squash` tag are bubbled up.
type FlatConfig struct {
	Chdir         *string  `mapstructure:"chdir" required:"false" cty:"chdir" hcl:"chdir"`
	InstallCmd    []string `mapstructure:"install_cmd" required:"false" cty:"install_cmd" hcl:"install_cmd"`
	Keyword       *string  `mapstructure:"keyword" required:"false" cty:"keyword" hcl:"keyword"`
	Local         *bool    `mapstructure:"local" required:"false" cty:"local" hcl:"local"`
	Marker        *string  `mapstructure:"marker" required:"false" cty:"marker" hcl:"marker"`
	Parallel      *bool    `mapstructure:"parallel" required:"false" cty:"parallel" hcl:"parallel"`
	PytestPath    *string  `mapstructure:"pytest_path" required:"false" cty:"pytest_path" hcl:"pytest_path"`
	Sudo          *bool    `mapstructure:"sudo" required:"false" cty:"sudo" hcl:"sudo"`
	SudoUser      *string  `mapstructure:"sudo_user" required:"false" cty:"sudo_user" hcl:"sudo_user"`
	TransferFiles *bool    `mapstructure:"transfer_files" required:"false" cty:"transfer_files" hcl:"transfer_files"`
	TestFiles     []string `mapstructure:"test_files" required:"false" cty:"test_files" hcl:"test_files"`
	Verbose       *int     `mapstructure:"verbose" required:"false" cty:"verbose" hcl:"verbose"`
}

// FlatMapstructure returns a new FlatConfig.
// FlatConfig is an auto-generated flat version of Config.
// Where the contents a fields with a `mapstructure:,squash` tag are bubbled up.
func (*Config) FlatMapstructure() interface{ HCL2Spec() map[string]hcldec.Spec } {
	return new(FlatConfig)
}

// HCL2Spec returns the hcl spec of a Config.
// This spec is used by HCL to read the fields of Config.
// The decoded values from this spec will then be applied to a FlatConfig.
func (*FlatConfig) HCL2Spec() map[string]hcldec.Spec {
	s := map[string]hcldec.Spec{
		"chdir":          &hcldec.AttrSpec{Name: "chdir", Type: cty.String, Required: false},
		"install_cmd":    &hcldec.AttrSpec{Name: "install_cmd", Type: cty.List(cty.String), Required: false},
		"keyword":        &hcldec.AttrSpec{Name: "keyword", Type: cty.String, Required: false},
		"local":          &hcldec.AttrSpec{Name: "local", Type: cty.Bool, Required: false},
		"marker":         &hcldec.AttrSpec{Name: "marker", Type: cty.String, Required: false},
		"parallel":       &hcldec.AttrSpec{Name: "parallel", Type: cty.Bool, Required: false},
		"pytest_path":    &hcldec.AttrSpec{Name: "pytest_path", Type: cty.String, Required: false},
		"sudo":           &hcldec.AttrSpec{Name: "sudo", Type: cty.Bool, Required: false},
		"sudo_user":      &hcldec.AttrSpec{Name: "sudo_user", Type: cty.String, Required: false},
		"transfer_files": &hcldec.AttrSpec{Name: "transfer_files", Type: cty.Bool, Required: false},
		"test_files":     &hcldec.AttrSpec{Name: "test_files", Type: cty.List(cty.String), Required: false},
		"verbose":        &hcldec.AttrSpec{Name: "verbose", Type: cty.Number, Required: false},
	}
	return s
}
