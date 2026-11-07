package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
				Type:          schema.TypeSet,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"protocol", "ports"},
				Elem:          &schema.Schema{Type: schema.TypeString},
				Description:   "The ports as `protocol/port` strings, for example `tcp/80`. Use this for a template that spans more than one protocol. Needs NetBox 4.6 or newer. Use `protocol` and `ports` on older versions.",
			},
			"protocol": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "One of [tcp, udp, sctp]",
				ValidateFunc: validation.StringInSlice([]string{"tcp", "udp", "sctp"}, false),
			},
			"ports": {
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
				Description: "The ports, together with `protocol`. Use `port_mappings` for a template that spans more than one protocol.",
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
	PortMappings []string           `json:"port_mappings,omitempty"`
	Protocol     *string            `json:"protocol,omitempty"`
	Ports        []int64            `json:"ports,omitempty"`
	Description  string             `json:"description,omitempty"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
}

type serviceTemplateResponse struct {
	Name         string                 `json:"name"`
	PortMappings []string               `json:"port_mappings"`
	Protocol     *netboxapi.Choice      `json:"protocol"`
	Ports        []int64                `json:"ports"`
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
		Name:        d.Get("name").(string),
		Description: getOptionalStr(d, "description", true),
		Owner:       getOptionalInt(d, "owner_id"),
		Comments:    getOptionalStr(d, "comments", true),
		Tags:        tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	// Either the port mappings or the protocol with the ports is sent, never both.
	if raw := d.GetRawConfig(); !raw.IsNull() && !raw.GetAttr("port_mappings").IsNull() {
		data.PortMappings = getStringSet(d, "port_mappings")
	} else {
		data.Protocol = getOptionalStrClear(d, "protocol")
		data.Ports = getIDSet(d, "ports")
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
		d.Set("protocol", netboxapi.ChoiceValue(r.Protocol))
		d.Set("ports", r.Ports)
		d.Set("description", r.Description)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("comments", r.Comments)
		api.readTagRefs(d, r.Tags)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
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
