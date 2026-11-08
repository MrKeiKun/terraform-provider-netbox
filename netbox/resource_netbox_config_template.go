package netbox

import (
	"context"
	"encoding/json"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxConfigTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetboxConfigTemplateCreate,
		ReadContext:   resourceNetboxConfigTemplateRead,
		UpdateContext: resourceNetboxConfigTemplateUpdate,
		DeleteContext: resourceNetboxConfigTemplateDelete,

		Description: `:meta:subcategory:Extras:From the [official documentation](https://docs.netbox.dev/en/stable/models/extras/configtemplate/):

> Configuration templates can be used to render device configurations from context data. Templates are written in the Jinja2 language and can be associated with devices roles, platforms, and/or individual devices.

> Context data is made available to devices and/or virtual machines based on their relationships to other objects in NetBox. For example, context data can be associated only with devices assigned to a particular site, or only to virtual machines in a certain cluster.`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Description:  "The name of the config template.",
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the config template.",
				Optional:    true,
			},
			"template_code": {
				Type:        schema.TypeString,
				Description: "The template code of the config template.",
				Required:    true,
			},
			"environment_params": {
				Type:         schema.TypeString,
				Description:  "The environment params of the config template.",
				Optional:     true,
				Default:      "{}",
				ValidateFunc: validation.StringIsJSON,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the config template.",
				Optional:    true,
			},
			"mime_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Defaults to <code>text/plain; charset=utf-8</code>",
			},
			"file_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filename to give to the rendered export file",
			},
			"file_extension": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Extension to append to the rendered filename",
			},
			"as_attachment": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Download file as attachment",
			},
			"debug": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Enable verbose error output when rendering this template. Not recommended for production use.",
			},
			"auto_sync_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Enable automatic synchronization of data when the data file is updated",
			},
			tagsKey: tagsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const configTemplateAPIPath = "extras/config-templates/"

type configTemplateRequest struct {
	Name              string             `json:"name"`
	Description       string             `json:"description,omitempty"`
	TemplateCode      string             `json:"template_code"`
	EnvironmentParams interface{}        `json:"environment_params,omitempty"`
	Owner             *int64             `json:"owner"`
	MimeType          string             `json:"mime_type,omitempty"`
	FileName          string             `json:"file_name,omitempty"`
	FileExtension     string             `json:"file_extension,omitempty"`
	AsAttachment      *bool              `json:"as_attachment,omitempty"`
	Debug             *bool              `json:"debug,omitempty"`
	AutoSyncEnabled   *bool              `json:"auto_sync_enabled,omitempty"`
	Tags              []netboxapi.TagRef `json:"tags"`
}

type configTemplateResponse struct {
	Name              string         `json:"name"`
	Description       string         `json:"description"`
	TemplateCode      string         `json:"template_code"`
	EnvironmentParams interface{}    `json:"environment_params"`
	Owner             *netboxapi.Ref `json:"owner"`
	MimeType          string         `json:"mime_type"`
	FileName          string         `json:"file_name"`
	FileExtension     string         `json:"file_extension"`
	AsAttachment      bool           `json:"as_attachment"`
	Debug             bool           `json:"debug"`
	AutoSyncEnabled   bool           `json:"auto_sync_enabled"`
}

func configTemplateRequestFromResourceData(api *providerState, d *schema.ResourceData) (*configTemplateRequest, error) {
	// A failed tag lookup leaves the tags empty, same as before.
	tags, _ := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))

	data := &configTemplateRequest{
		AutoSyncEnabled: getOptionalBool(d, "auto_sync_enabled"),
		Debug:           getOptionalBool(d, "debug"),
		AsAttachment:    getOptionalBool(d, "as_attachment"),
		FileExtension:   getOptionalStr(d, "file_extension", true),
		FileName:        getOptionalStr(d, "file_name", true),
		MimeType:        getOptionalStr(d, "mime_type", true),
		Owner:           getOptionalInt(d, "owner_id"),
		Name:            d.Get("name").(string),
		Description:     d.Get("description").(string),
		TemplateCode:    d.Get("template_code").(string),
		Tags:            tags,
	}

	// Unmarshal environment_params and add it to data if valid
	if environmentParamsJSON, ok := d.GetOk("environment_params"); ok {
		var environmentParams any
		if err := json.Unmarshal([]byte(environmentParamsJSON.(string)), &environmentParams); err != nil {
			return nil, err
		}
		data.EnvironmentParams = environmentParams
	}
	return data, nil
}

func resourceNetboxConfigTemplateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	data, err := configTemplateRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return diag.FromErr(err)
	}
	return createObjectCtx(ctx, d, m, configTemplateAPIPath, data, resourceNetboxConfigTemplateRead)
}

func resourceNetboxConfigTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var marshalErr error
	diags := readObjectCtx(ctx, d, m, configTemplateAPIPath, func(tmpl *configTemplateResponse) {
		d.Set("auto_sync_enabled", tmpl.AutoSyncEnabled)
		d.Set("debug", tmpl.Debug)
		d.Set("as_attachment", tmpl.AsAttachment)
		d.Set("file_extension", tmpl.FileExtension)
		d.Set("file_name", tmpl.FileName)
		d.Set("mime_type", tmpl.MimeType)
		d.Set("owner_id", netboxapi.RefID(tmpl.Owner))
		d.Set("name", tmpl.Name)
		d.Set("description", tmpl.Description)
		d.Set("template_code", tmpl.TemplateCode)
		if tmpl.EnvironmentParams != nil {
			environmentParamsJSON, err := json.Marshal(tmpl.EnvironmentParams)
			if err != nil {
				marshalErr = err
				return
			}
			d.Set("environment_params", string(environmentParamsJSON))
		} else {
			d.Set("environment_params", "{}")
		}
	})
	if marshalErr != nil {
		diags = append(diags, diag.FromErr(marshalErr)...)
	}
	return diags
}

func resourceNetboxConfigTemplateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	data, err := configTemplateRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return diag.FromErr(err)
	}
	return putObjectCtx(ctx, d, m, configTemplateAPIPath, data, resourceNetboxConfigTemplateRead)
}

func resourceNetboxConfigTemplateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return deleteObjectCtx(ctx, d, m, configTemplateAPIPath)
}
