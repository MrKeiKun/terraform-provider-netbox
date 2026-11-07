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
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"template_code": {
				Type:     schema.TypeString,
				Required: true,
			},
			"environment_params": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "{}",
				ValidateFunc: validation.StringIsJSON,
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
	Tags              []netboxapi.TagRef `json:"tags"`
}

type configTemplateResponse struct {
	Name              string      `json:"name"`
	Description       string      `json:"description"`
	TemplateCode      string      `json:"template_code"`
	EnvironmentParams interface{} `json:"environment_params"`
}

func configTemplateRequestFromResourceData(api *providerState, d *schema.ResourceData) (*configTemplateRequest, error) {
	// A failed tag lookup leaves the tags empty, same as before.
	tags, _ := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))

	data := &configTemplateRequest{
		Name:         d.Get("name").(string),
		Description:  d.Get("description").(string),
		TemplateCode: d.Get("template_code").(string),
		Tags:         tags,
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
