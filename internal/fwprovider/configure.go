package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwresource"
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/MrKeiKun/terraform-provider-netbox/internal/providerstate"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type providerModel struct {
	ServerURL                   types.String `tfsdk:"server_url"`
	APIToken                    types.String `tfsdk:"api_token"`
	AllowInsecureHTTPS          types.Bool   `tfsdk:"allow_insecure_https"`
	Headers                     types.Map    `tfsdk:"headers"`
	StripTrailingSlashesFromURL types.Bool   `tfsdk:"strip_trailing_slashes_from_url"`
	SkipVersionCheck            types.Bool   `tfsdk:"skip_version_check"`
	RequestTimeout              types.Int64  `tfsdk:"request_timeout"`
	MaxRetries                  types.Int64  `tfsdk:"max_retries"`
	DefaultTags                 types.Set    `tfsdk:"default_tags"`
	CACertFile                  types.String `tfsdk:"ca_cert_file"`
}

func envString(v types.String, env string) string {
	if !v.IsNull() && !v.IsUnknown() {
		return v.ValueString()
	}
	return os.Getenv(env)
}

func envBool(v types.Bool, env string, def bool, diags *diag.Diagnostics) bool {
	if !v.IsNull() && !v.IsUnknown() {
		return v.ValueBool()
	}
	if s := os.Getenv(env); s != "" {
		b, err := strconv.ParseBool(s)
		if err != nil {
			diags.AddError("Invalid "+env, fmt.Sprintf("%q is not true or false", s))
			return def
		}
		return b
	}
	return def
}

// envInt returns a whole number from 0 to math.MaxInt32, taken from the configuration or the
// environment variable.
func envInt(v types.Int64, env string, def int, diags *diag.Diagnostics) int {
	n := int64(def)
	if !v.IsNull() && !v.IsUnknown() {
		n = v.ValueInt64()
	} else if s := os.Getenv(env); s != "" {
		parsed, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			diags.AddError("Invalid "+env, fmt.Sprintf("%q is not a whole number", s))
			return def
		}
		n = parsed
	}
	if n < 0 || n > math.MaxInt32 {
		diags.AddError("Invalid "+env, fmt.Sprintf("%d is not a whole number from 0 to %d", n, math.MaxInt32))
		return def
	}
	return int(n)
}

// Configure creates the client the resources and data sources of the framework provider use. The
// settings fall back to the same environment variables as the SDKv2 provider.
func (p *Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Values that are not known yet (for example computed by another resource) cannot be used.
	if cfg.ServerURL.IsUnknown() || cfg.APIToken.IsUnknown() || cfg.AllowInsecureHTTPS.IsUnknown() ||
		cfg.Headers.IsUnknown() || cfg.StripTrailingSlashesFromURL.IsUnknown() || cfg.SkipVersionCheck.IsUnknown() ||
		cfg.RequestTimeout.IsUnknown() || cfg.MaxRetries.IsUnknown() || cfg.DefaultTags.IsUnknown() || cfg.CACertFile.IsUnknown() {
		return
	}

	settings := providerstate.Settings{}
	settings.ServerURL = envString(cfg.ServerURL, "NETBOX_SERVER_URL")
	settings.APIToken = envString(cfg.APIToken, "NETBOX_API_TOKEN")
	settings.AllowInsecureHTTPS = envBool(cfg.AllowInsecureHTTPS, "NETBOX_ALLOW_INSECURE_HTTPS", false, &resp.Diagnostics)
	settings.StripTrailingSlashesFromURL = envBool(cfg.StripTrailingSlashesFromURL, "NETBOX_STRIP_TRAILING_SLASHES_FROM_URL", true, &resp.Diagnostics)
	settings.SkipVersionCheck = envBool(cfg.SkipVersionCheck, "NETBOX_SKIP_VERSION_CHECK", false, &resp.Diagnostics)
	settings.RequestTimeout = envInt(cfg.RequestTimeout, "NETBOX_REQUEST_TIMEOUT", 10, &resp.Diagnostics)
	settings.MaxRetries = envInt(cfg.MaxRetries, "NETBOX_MAX_RETRIES", netboxapi.DefaultMaxRetries, &resp.Diagnostics)
	settings.CACertFile = envString(cfg.CACertFile, "NETBOX_CA_CERT_FILE")

	if settings.ServerURL == "" {
		resp.Diagnostics.AddAttributeError(path.Root("server_url"), "Missing server_url", "Set `server_url` in the provider configuration or the `NETBOX_SERVER_URL` environment variable.")
	}
	if settings.APIToken == "" {
		resp.Diagnostics.AddAttributeError(path.Root("api_token"), "Missing api_token", "Set `api_token` in the provider configuration or the `NETBOX_API_TOKEN` environment variable.")
	}
	if resp.Diagnostics.HasError() {
		return
	}

	settings.Headers = map[string]interface{}{}
	if !cfg.Headers.IsNull() && !cfg.Headers.IsUnknown() {
		for k, v := range cfg.Headers.Elements() {
			if s, ok := v.(types.String); ok {
				settings.Headers[k] = s.ValueString()
			}
		}
	} else if raw := os.Getenv("NETBOX_HEADERS"); raw != "" {
		var h map[string]string
		if err := json.Unmarshal([]byte(raw), &h); err != nil {
			resp.Diagnostics.AddError("Invalid NETBOX_HEADERS", fmt.Sprintf("the variable must hold a JSON object of header names and values: %s", err))
			return
		}
		for k, v := range h {
			settings.Headers[k] = v
		}
	}

	if !cfg.DefaultTags.IsNull() && !cfg.DefaultTags.IsUnknown() {
		for _, e := range cfg.DefaultTags.Elements() {
			if s, ok := e.(types.String); ok {
				settings.DefaultTags = append(settings.DefaultTags, s.ValueString())
			}
		}
	}

	state, diags := providerstate.Connect(ctx, settings)
	for _, d := range diags {
		if d.Error {
			resp.Diagnostics.AddError(d.Summary, d.Detail)
		} else {
			resp.Diagnostics.AddWarning(d.Summary, d.Detail)
		}
	}
	if state == nil || resp.Diagnostics.HasError() {
		return
	}

	resp.ResourceData = state
	resp.DataSourceData = state
}

func (p *Provider) Resources(_ context.Context) []func() resource.Resource {
	out := make([]func() resource.Resource, 0, len(standardResources))
	for _, spec := range standardResources {
		out = append(out, fwresource.New(spec))
	}
	for _, spec := range manualResources {
		out = append(out, fwresource.New(spec))
	}
	out = append(out, customResources()...)
	return out
}

func (p *Provider) DataSources(_ context.Context) []func() datasource.DataSource {
	out := make([]func() datasource.DataSource, 0, len(standardDataSources)+len(customDataSources))
	for _, spec := range standardDataSources {
		out = append(out, fwresource.NewDataSource(spec))
	}
	for _, spec := range customDataSources {
		out = append(out, fwresource.NewDataSource(spec))
	}
	for _, spec := range generatedDataSources {
		out = append(out, fwresource.NewDataSource(spec))
	}
	for _, spec := range aliasedDataSources() {
		out = append(out, fwresource.NewDataSource(spec))
	}
	return out
}
