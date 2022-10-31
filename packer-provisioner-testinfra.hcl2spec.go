// Code generated by "packer-sdc mapstructure-to-hcl2"; DO NOT EDIT.

package main

import (
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"
)

// FlatTestinfraConfig is an auto-generated flat version of TestinfraConfig.
// Where the contents of a field with a `mapstructure:,squash` tag are bubbled up.
type FlatTestinfraConfig struct {
	Processes  *int     `mapstructure:"processes" cty:"processes" hcl:"processes"`
	PytestPath *string  `mapstructure:"pytest_path" cty:"pytest_path" hcl:"pytest_path"`
	TestFiles  []string `mapstructure:"test_files" cty:"test_files" hcl:"test_files"`
}

// FlatMapstructure returns a new FlatTestinfraConfig.
// FlatTestinfraConfig is an auto-generated flat version of TestinfraConfig.
// Where the contents a fields with a `mapstructure:,squash` tag are bubbled up.
func (*TestinfraConfig) FlatMapstructure() interface{ HCL2Spec() map[string]hcldec.Spec } {
	return new(FlatTestinfraConfig)
}

// HCL2Spec returns the hcl spec of a TestinfraConfig.
// This spec is used by HCL to read the fields of TestinfraConfig.
// The decoded values from this spec will then be applied to a FlatTestinfraConfig.
func (*FlatTestinfraConfig) HCL2Spec() map[string]hcldec.Spec {
	s := map[string]hcldec.Spec{
		"processes":   &hcldec.AttrSpec{Name: "processes", Type: cty.Number, Required: false},
		"pytest_path": &hcldec.AttrSpec{Name: "pytest_path", Type: cty.String, Required: false},
		"test_files":  &hcldec.AttrSpec{Name: "test_files", Type: cty.List(cty.String), Required: false},
	}
	return s
}
