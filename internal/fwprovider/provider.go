// Package fwprovider holds the provider, which is built on the Terraform Plugin Framework and
// serves its resources and data sources over protocol version 6.
package fwprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &Provider{}

// Provider is the NetBox provider.
type Provider struct {
	version string
}

// New returns a function that creates the provider.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &Provider{version: version}
	}
}

func (p *Provider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "netbox"
	resp.Version = p.version
}

func (p *Provider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"server_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Location of Netbox server including scheme (http or https) and optional port. Can be set via the `NETBOX_SERVER_URL` environment variable.",
			},
			"api_token": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Netbox API authentication token. Supports both v1 tokens (`Authorization: Token <key>`) and v2 tokens (`Authorization: Bearer nbt_<key>.<token>`). V2 tokens are auto-detected by their `nbt_` prefix. Can be set via the `NETBOX_API_TOKEN` environment variable.",
			},
			"allow_insecure_https": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Flag to set whether to allow https with invalid certificates. Can be set via the `NETBOX_ALLOW_INSECURE_HTTPS` environment variable. Defaults to `false`.",
			},
			"headers": schema.MapAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Set these header on all requests to Netbox. Can be set via the `NETBOX_HEADERS` environment variable.",
			},
			"strip_trailing_slashes_from_url": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "If true, strip trailing slashes from the `server_url` parameter and print a warning when doing so. Note that using trailing slashes in the `server_url` parameter will usually lead to errors. Can be set via the `NETBOX_STRIP_TRAILING_SLASHES_FROM_URL` environment variable. Defaults to `true`.",
			},
			"skip_version_check": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "If true, do not try to determine the running Netbox version at provider startup. Disables warnings about possibly unsupported Netbox version. Also useful for local testing on terraform plans. Can be set via the `NETBOX_SKIP_VERSION_CHECK` environment variable. Defaults to `false`.",
			},
			"request_timeout": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Netbox API HTTP request timeout in seconds. Can be set via the `NETBOX_REQUEST_TIMEOUT` environment variable. Defaults to `10`.",
			},
			"max_retries": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "How many times a failed GET, PUT or DELETE request is sent again after a connection error, or a 502, 503 or 504 response. A request that fails because of a database deadlock in Netbox is sent again whatever its method, since nothing was changed. Other requests that create objects are never repeated. Set to `0` to turn retries off. Can be set via the `NETBOX_MAX_RETRIES` environment variable. Defaults to `3`.",
			},
			"default_tags": schema.SetAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Tags to add to every resource managed by this provider.",
			},
			"ca_cert_file": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Path to a PEM-encoded CA certificate for verifying the Netbox server certificate. Can be set via the `NETBOX_CA_CERT_FILE` environment variable.",
			},
		},
	}
}
