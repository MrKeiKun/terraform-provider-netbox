package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxDeviceRole() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxDeviceRoleCreate,
		Read:   resourceNetboxDeviceRoleRead,
		Update: resourceNetboxDeviceRoleUpdate,
		Delete: resourceNetboxDeviceRoleDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/features/devices/#device-roles):

> Devices can be organized by functional roles, which are fully customizable by the user. For example, you might create roles for core switches, distribution switches, and access switches within your network.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the device role.",
				Required:    true,
			},
			"slug": {
				Type:         schema.TypeString,
				Description:  "The URL-friendly slug of the device role.",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"vm_role": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Virtual machines may be assigned to this role.",
			},
			"color_hex": {
				Type:        schema.TypeString,
				Description: "The color hex of the device role.",
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the device role.",
				Optional:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the device role.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the device role.",
				Optional:    true,
			},
			"config_template_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the config template associated with the device role.",
				Optional:    true,
			},
			customFieldsKey: customFieldsSchema,
			tagsKey:         tagsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const deviceRoleAPIPath = "dcim/device-roles/"

type deviceRoleRequest struct {
	Name           string             `json:"name"`
	Slug           string             `json:"slug"`
	Color          string             `json:"color,omitempty"`
	VmRole         bool               `json:"vm_role"`
	Description    string             `json:"description,omitempty"`
	Owner          *int64             `json:"owner"`
	Comments       string             `json:"comments,omitempty"`
	ConfigTemplate *int64             `json:"config_template"`
	CustomFields   interface{}        `json:"custom_fields,omitempty"`
	Tags           []netboxapi.TagRef `json:"tags"`
}

type deviceRoleResponse struct {
	Name           string                 `json:"name"`
	Slug           string                 `json:"slug"`
	Color          string                 `json:"color"`
	VmRole         bool                   `json:"vm_role"`
	Description    string                 `json:"description"`
	Owner          *netboxapi.Ref         `json:"owner"`
	Comments       string                 `json:"comments"`
	ConfigTemplate *netboxapi.Ref         `json:"config_template"`
	CustomFields   map[string]interface{} `json:"custom_fields"`
	Tags           []netboxapi.TagRef     `json:"tags"`
}

func deviceRoleRequestFromResourceData(api *providerState, d *schema.ResourceData) (*deviceRoleRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &deviceRoleRequest{
		CustomFields:   customFieldsRequest(d),
		ConfigTemplate: getOptionalInt(d, "config_template_id"),
		Comments:       getOptionalStr(d, "comments", true),
		Owner:          getOptionalInt(d, "owner_id"),
		Name:           name,
		Slug:           slugFromResourceData(d, name),
		Color:          d.Get("color_hex").(string),
		VmRole:         d.Get("vm_role").(bool),
		Description:    d.Get("description").(string),
		Tags:           tags,
	}
	return data, nil
}

func resourceNetboxDeviceRoleCreate(d *schema.ResourceData, m interface{}) error {
	data, err := deviceRoleRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, deviceRoleAPIPath, data, resourceNetboxDeviceRoleRead)
}

func resourceNetboxDeviceRoleRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, deviceRoleAPIPath, func(r *deviceRoleResponse) {
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		d.Set("config_template_id", netboxapi.RefID(r.ConfigTemplate))
		d.Set("comments", r.Comments)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("name", r.Name)
		d.Set("slug", r.Slug)
		d.Set("color_hex", r.Color)
		d.Set("vm_role", r.VmRole)
		d.Set("description", r.Description)
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxDeviceRoleUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := deviceRoleRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, deviceRoleAPIPath, data, resourceNetboxDeviceRoleRead)
}

func resourceNetboxDeviceRoleDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, deviceRoleAPIPath)
}
