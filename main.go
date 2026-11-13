package main

import (
	"context"
	"flag"
	"log"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwprovider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// Run "go generate" to format example terraform files and generate the docs for the registry/website

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//go:generate go tool tfplugindocs generate --provider-name terraform-provider-netbox
// The docs tool has no way to set the subcategory from the schema, so the descriptions carry a
// ":meta:subcategory:<name>:" marker. This step moves it into the page header of the generated docs.
//go:generate go run ./scripts/postprocess-docs docs

// version is set at build time.
var version = "dev"

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/mrkeikun/netbox",
		Debug:   debug,
	}
	if err := providerserver.Serve(context.Background(), fwprovider.New(version), opts); err != nil {
		log.Fatal(err)
	}
}
