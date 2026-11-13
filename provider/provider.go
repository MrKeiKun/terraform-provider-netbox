// Package provider exposes the NetBox Terraform provider as a Go library, for tools that serve it in
// process instead of through the Terraform CLI.
//
// The implementation lives in internal/fwprovider, which other modules cannot import. This package
// only passes on its constructor.
package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/provider"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwprovider"
)

// New returns the constructor of the NetBox provider, as passed to providerserver.Serve by main.go.
// The version is reported as the version of the provider.
func New(version string) func() provider.Provider {
	return fwprovider.New(version)
}
