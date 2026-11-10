package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxInventoryItemTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxInventoryItemTemplateCreate,
		Read:   resourceNetboxInventoryItemTemplateRead,
		Update: resourceNetboxInventoryItemTemplateUpdate,
		Delete: resourceNetboxInventoryItemTemplateDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/inventoryitemtemplate/):

> An inventory item template describes a part that is created on every device of a device type.`,

		Schema: map[string]*schema.Schema{
			"device_type_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the device type associated with the inventory item template.",
				Required:    true,
			},
			"parent_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the parent associated with the inventory item template.",
				Optional:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "{module} is accepted as a substitution for the module bay position when attached to a module type.",
			},
			"label": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Physical label",
			},
			"role_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the role associated with the inventory item template.",
				Optional:    true,
			},
			"manufacturer_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the manufacturer associated with the inventory item template.",
				Optional:    true,
			},
			"part_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Manufacturer-assigned part identifier",
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the inventory item template.",
				Optional:    true,
			},
			"component_type": {
				Type:        schema.TypeString,
				Description: "The component type of the inventory item template.",
				Optional:    true,
			},
			"component_id": {
				Type:         schema.TypeInt,
				Description:  "The ID of the component associated with the inventory item template.",
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const inventoryItemTemplateAPIPath = "dcim/inventory-item-templates/"

type inventoryItemTemplateRequest struct {
	DeviceType    *int64 `json:"device_type"`
	Parent        *int64 `json:"parent"`
	Name          string `json:"name"`
	Label         string `json:"label,omitempty"`
	Role          *int64 `json:"role"`
	Manufacturer  *int64 `json:"manufacturer"`
	PartID        string `json:"part_id,omitempty"`
	Description   string `json:"description,omitempty"`
	ComponentType string `json:"component_type,omitempty"`
	ComponentID   *int64 `json:"component_id"`
}

type inventoryItemTemplateResponse struct {
	DeviceType    *netboxapi.Ref `json:"device_type"`
	Parent        *netboxapi.Ref `json:"parent"`
	Name          string         `json:"name"`
	Label         string         `json:"label"`
	Role          *netboxapi.Ref `json:"role"`
	Manufacturer  *netboxapi.Ref `json:"manufacturer"`
	PartID        string         `json:"part_id"`
	Description   string         `json:"description"`
	ComponentType string         `json:"component_type"`
	ComponentID   *int64         `json:"component_id"`
}

func inventoryItemTemplateRequestFromResourceData(api *providerState, d *schema.ResourceData) (*inventoryItemTemplateRequest, error) {
	data := &inventoryItemTemplateRequest{
		DeviceType:    int64ToPtr(int64(d.Get("device_type_id").(int))),
		Parent:        getOptionalInt(d, "parent_id"),
		Name:          d.Get("name").(string),
		Label:         getOptionalStr(d, "label", true),
		Role:          getOptionalInt(d, "role_id"),
		Manufacturer:  getOptionalInt(d, "manufacturer_id"),
		PartID:        getOptionalStr(d, "part_id", true),
		Description:   getOptionalStr(d, "description", true),
		ComponentType: getOptionalStr(d, "component_type", true),
		ComponentID:   getOptionalInt(d, "component_id"),
	}
	return data, nil
}

func resourceNetboxInventoryItemTemplateCreate(d *schema.ResourceData, m interface{}) error {
	data, err := inventoryItemTemplateRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, inventoryItemTemplateAPIPath, data, resourceNetboxInventoryItemTemplateRead)
}

func resourceNetboxInventoryItemTemplateRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, inventoryItemTemplateAPIPath, func(r *inventoryItemTemplateResponse) {
		d.Set("device_type_id", netboxapi.RefID(r.DeviceType))
		d.Set("parent_id", netboxapi.RefID(r.Parent))
		d.Set("name", r.Name)
		d.Set("label", r.Label)
		d.Set("role_id", netboxapi.RefID(r.Role))
		d.Set("manufacturer_id", netboxapi.RefID(r.Manufacturer))
		d.Set("part_id", r.PartID)
		d.Set("description", r.Description)
		d.Set("component_type", r.ComponentType)
		d.Set("component_id", r.ComponentID)
	})
}

func resourceNetboxInventoryItemTemplateUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := inventoryItemTemplateRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, inventoryItemTemplateAPIPath, data, resourceNetboxInventoryItemTemplateRead)
}

func resourceNetboxInventoryItemTemplateDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, inventoryItemTemplateAPIPath)
}
