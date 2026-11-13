package provider_test

import (
	"context"
	"testing"

	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/MrKeiKun/terraform-provider-netbox/provider"
)

// The public constructor returns the NetBox provider, and it can be served without the Terraform CLI.
func TestNew(t *testing.T) {
	ctx := context.Background()
	p := provider.New("1.2.3")()

	var meta fwprovider.MetadataResponse
	p.Metadata(ctx, fwprovider.MetadataRequest{}, &meta)
	if meta.TypeName != "netbox" {
		t.Errorf("TypeName: want netbox, got %q", meta.TypeName)
	}
	if meta.Version != "1.2.3" {
		t.Errorf("Version: want 1.2.3, got %q", meta.Version)
	}

	server, err := providerserver.NewProtocol6WithError(p)()
	if err != nil {
		t.Fatal(err)
	}
	schema, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range schema.Diagnostics {
		if d.Severity == tfprotov6.DiagnosticSeverityError {
			t.Errorf("%s: %s", d.Summary, d.Detail)
		}
	}
	if len(schema.ResourceSchemas) == 0 || len(schema.DataSourceSchemas) == 0 {
		t.Errorf("got %d resources and %d data sources", len(schema.ResourceSchemas), len(schema.DataSourceSchemas))
	}
}
