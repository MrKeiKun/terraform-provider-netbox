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
				Type:     schema.TypeString,
				Required: true,
			},
			"slug": {
				Type:         schema.TypeString,
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
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			tagsKey: tagsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const deviceRoleAPIPath = "dcim/device-roles/"

type deviceRoleRequest struct {
	Name        string             `json:"name"`
	Slug        string             `json:"slug"`
	Color       string             `json:"color,omitempty"`
	VmRole      bool               `json:"vm_role"`
	Description string             `json:"description,omitempty"`
	Tags        []netboxapi.TagRef `json:"tags"`
}

type deviceRoleResponse struct {
	Name        string             `json:"name"`
	Slug        string             `json:"slug"`
	Color       string             `json:"color"`
	VmRole      bool               `json:"vm_role"`
	Description string             `json:"description"`
	Tags        []netboxapi.TagRef `json:"tags"`
}

func deviceRoleRequestFromResourceData(api *providerState, d *schema.ResourceData) (*deviceRoleRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &deviceRoleRequest{
		Name:        name,
		Slug:        slugFromResourceData(d, name),
		Color:       d.Get("color_hex").(string),
		VmRole:      d.Get("vm_role").(bool),
		Description: d.Get("description").(string),
		Tags:        tags,
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
