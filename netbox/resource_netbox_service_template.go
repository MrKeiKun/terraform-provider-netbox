package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxServiceTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxServiceTemplateCreate,
		Read:   resourceNetboxServiceTemplateRead,
		Update: resourceNetboxServiceTemplateUpdate,
		Delete: resourceNetboxServiceTemplateDelete,

		Description: `:meta:subcategory:IP Address Management (IPAM):From the [official documentation](https://docs.netbox.dev/en/stable/models/ipam/servicetemplate/):

> A service template is a reusable definition of a service, with its protocol and ports.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"port_mappings": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"owner_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"comments": {
				Type:     schema.TypeString,
				Optional: true,
			},
			tagsKey:         tagsSchema,
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const serviceTemplateAPIPath = "ipam/service-templates/"

type serviceTemplateRequest struct {
	Name         string             `json:"name"`
	PortMappings []string           `json:"port_mappings"`
	Description  string             `json:"description,omitempty"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
}

type serviceTemplateResponse struct {
	Name         string                 `json:"name"`
	PortMappings []string               `json:"port_mappings"`
	Description  string                 `json:"description"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	Tags         []netboxapi.TagRef     `json:"tags"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func serviceTemplateRequestFromResourceData(api *providerState, d *schema.ResourceData) (*serviceTemplateRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &serviceTemplateRequest{
		Name:         d.Get("name").(string),
		PortMappings: getStringSet(d, "port_mappings"),
		Description:  getOptionalStr(d, "description", true),
		Owner:        getOptionalInt(d, "owner_id"),
		Comments:     getOptionalStr(d, "comments", true),
		Tags:         tags,
	}
	if ct, ok := d.GetOk(customFieldsKey); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxServiceTemplateCreate(d *schema.ResourceData, m interface{}) error {
	data, err := serviceTemplateRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, serviceTemplateAPIPath, data, resourceNetboxServiceTemplateRead)
}

func resourceNetboxServiceTemplateRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, serviceTemplateAPIPath, func(r *serviceTemplateResponse) {
		d.Set("name", r.Name)
		d.Set("port_mappings", r.PortMappings)
		d.Set("description", r.Description)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("comments", r.Comments)
		api.readTagRefs(d, r.Tags)
		if cf := getCustomFields(r.CustomFields); cf != nil {
			d.Set(customFieldsKey, cf)
		}
	})
}

func resourceNetboxServiceTemplateUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := serviceTemplateRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, serviceTemplateAPIPath, data, resourceNetboxServiceTemplateRead)
}

func resourceNetboxServiceTemplateDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, serviceTemplateAPIPath)
}
