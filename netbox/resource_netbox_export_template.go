package netbox

import (
	"encoding/json"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxExportTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxExportTemplateCreate,
		Read:   resourceNetboxExportTemplateRead,
		Update: resourceNetboxExportTemplateUpdate,
		Delete: resourceNetboxExportTemplateDelete,

		Description: `:meta:subcategory:Extras:From the [official documentation](https://docs.netbox.dev/en/stable/models/customization/exporttemplate/):

> An export template renders a list of objects into a custom format with Jinja2.`,

		Schema: map[string]*schema.Schema{
			"object_types": {
				Type:        schema.TypeSet,
				Description: "The object types the export template applies to.",
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the export template.",
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the export template.",
				Optional:    true,
			},
			"environment_params": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentJSON,
				Description:      "Any <a href=\"https://jinja.palletsprojects.com/en/stable/api/#jinja2.Environment\">additional parameters</a> to pass when constructing the Jinja environment",
			},
			"template_code": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Jinja template code.",
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
				Default:     false,
				Description: "Download file as attachment",
			},
			"data_source_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the data source associated with the export template.",
				Optional:    true,
			},
			"auto_sync_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable automatic synchronization of data when the data file is updated",
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the export template.",
				Optional:    true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const exportTemplateAPIPath = "extras/export-templates/"

type exportTemplateRequest struct {
	ObjectTypes       []string         `json:"object_types"`
	Name              string           `json:"name"`
	Description       string           `json:"description,omitempty"`
	EnvironmentParams *json.RawMessage `json:"environment_params"`
	TemplateCode      string           `json:"template_code"`
	MimeType          string           `json:"mime_type,omitempty"`
	FileName          string           `json:"file_name,omitempty"`
	FileExtension     string           `json:"file_extension,omitempty"`
	AsAttachment      bool             `json:"as_attachment"`
	DataSource        *int64           `json:"data_source,omitempty"`
	AutoSyncEnabled   bool             `json:"auto_sync_enabled"`
	Owner             *int64           `json:"owner"`
}

type exportTemplateResponse struct {
	ObjectTypes       []string        `json:"object_types"`
	Name              string          `json:"name"`
	Description       string          `json:"description"`
	EnvironmentParams json.RawMessage `json:"environment_params"`
	TemplateCode      string          `json:"template_code"`
	MimeType          string          `json:"mime_type"`
	FileName          string          `json:"file_name"`
	FileExtension     string          `json:"file_extension"`
	AsAttachment      bool            `json:"as_attachment"`
	DataSource        *netboxapi.Ref  `json:"data_source"`
	AutoSyncEnabled   bool            `json:"auto_sync_enabled"`
	Owner             *netboxapi.Ref  `json:"owner"`
}

func exportTemplateRequestFromResourceData(api *providerState, d *schema.ResourceData) (*exportTemplateRequest, error) {
	environmentParamsRaw, err := jsonRawFromResourceData(d, "environment_params")
	if err != nil {
		return nil, err
	}
	data := &exportTemplateRequest{
		ObjectTypes:       getStringSet(d, "object_types"),
		Name:              d.Get("name").(string),
		Description:       getOptionalStr(d, "description", true),
		EnvironmentParams: environmentParamsRaw,
		TemplateCode:      d.Get("template_code").(string),
		MimeType:          getOptionalStr(d, "mime_type", true),
		FileName:          getOptionalStr(d, "file_name", true),
		FileExtension:     getOptionalStr(d, "file_extension", true),
		AsAttachment:      d.Get("as_attachment").(bool),
		DataSource:        getOptionalInt(d, "data_source_id"),
		AutoSyncEnabled:   d.Get("auto_sync_enabled").(bool),
		Owner:             getOptionalInt(d, "owner_id"),
	}
	return data, nil
}

func resourceNetboxExportTemplateCreate(d *schema.ResourceData, m interface{}) error {
	data, err := exportTemplateRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, exportTemplateAPIPath, data, resourceNetboxExportTemplateRead)
}

func resourceNetboxExportTemplateRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, exportTemplateAPIPath, func(r *exportTemplateResponse) {
		d.Set("object_types", r.ObjectTypes)
		d.Set("name", r.Name)
		d.Set("description", r.Description)
		d.Set("environment_params", jsonString(r.EnvironmentParams))
		d.Set("template_code", r.TemplateCode)
		d.Set("mime_type", r.MimeType)
		d.Set("file_name", r.FileName)
		d.Set("file_extension", r.FileExtension)
		d.Set("as_attachment", r.AsAttachment)
		d.Set("data_source_id", netboxapi.RefID(r.DataSource))
		d.Set("auto_sync_enabled", r.AutoSyncEnabled)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
	})
}

func resourceNetboxExportTemplateUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := exportTemplateRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, exportTemplateAPIPath, data, resourceNetboxExportTemplateRead)
}

func resourceNetboxExportTemplateDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, exportTemplateAPIPath)
}
