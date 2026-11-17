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
				Sensitive:           true,
				MarkdownDescription: "Netbox API authentication token. Supports both v1 tokens (`Authorization: Token <key>`) and v2 tokens (`Authorization: Bearer nbt_<key>.<token>`). V2 tokens are auto-detected by their `nbt_` prefix. Can be set via the `NETBOX_API_TOKEN` environment variable.",
			},
			"allow_insecure_https": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Flag to set whether to allow https with invalid certificates. Can be set via the `NETBOX_ALLOW_INSECURE_HTTPS` environment variable. Defaults to `false`.",
			},
			"headers": schema.MapAttribute{
				Optional: true,
				// Headers often carry credentials, for example for a proxy in front of Netbox.
				Sensitive:           true,
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
				MarkdownDescription: "How many times a failed request is sent again. GET, PUT and DELETE requests are repeated after a connection error or a 502, 503 or 504 response. A request that creates or changes an object (POST, PATCH) is only repeated when Netbox did not carry it out: no connection could be made, the response is 429 or 503, or Netbox reports a database deadlock. The wait doubles with every attempt and follows a `Retry-After` header. Set to `0` to turn retries off. Can be set via the `NETBOX_MAX_RETRIES` environment variable. Defaults to `3`.",
			},
			"page_size": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "How many objects a data source that lists objects asks Netbox for with each request. Netbox returns at most its `MAX_PAGE_SIZE` (1000 unless the administrator changed it) and the data source reads the next page until it has all objects. Lower it when Netbox is slow to answer a large page and requests time out. Can be set via the `NETBOX_PAGE_SIZE` environment variable. Defaults to `1000`.",
			},
			"batch_reads": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "If true, objects of the same kind that Terraform reads at the same time are read with one request to Netbox in place of one request for each object. That makes `terraform plan` faster and puts less load on Netbox. Set it to `false` when a proxy in front of Netbox does not allow list requests. Can be set via the `NETBOX_BATCH_READS` environment variable. Defaults to `true`.",
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
