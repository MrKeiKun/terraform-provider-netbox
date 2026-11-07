package main

import (
	"flag"

	"github.com/MrKeiKun/terraform-provider-netbox/netbox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

// Run "go generate" to format example terraform files and generate the docs for the registry/website

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//go:generate go tool tfplugindocs generate --provider-name terraform-provider-netbox
// The docs tool has no way to set the subcategory from the schema, so the descriptions carry a
// ":meta:subcategory:<name>:" marker. This step moves it into the page header of the generated docs.
//go:generate go run ./scripts/postprocess-docs docs

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	plugin.Serve(&plugin.ServeOpts{
		Debug:        debug,
		ProviderAddr: "registry.terraform.io/mrkeikun/netbox",
		ProviderFunc: func() *schema.Provider {
			return netbox.Provider()
		},
	})
}
