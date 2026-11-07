package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxInventoryItem() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxInventoryItemCreate,
		Read:   resourceNetboxInventoryItemRead,
		Update: resourceNetboxInventoryItemUpdate,
		Delete: resourceNetboxInventoryItemDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/inventoryitem/):

> Inventory items represent hardware components installed within a device, such as a power supply or CPU or line card. They are intended to be used primarily for inventory purposes.`,

		Schema: map[string]*schema.Schema{
			"device_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"parent_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"label": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"role_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"manufacturer_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"part_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"serial": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"asset_tag": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"discovered": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"component_type": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"dcim.powerport",
					"dcim.poweroutlet",
					"dcim.frontport",
					"dcim.rearport",
					"dcim.consoleserverport",
					"dcim.consoleport",
					"dcim.interface",
				}, false),
			},
			"component_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				RequiredWith: []string{"component_type"},
			},
			tagsKey:         tagsSchema,
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const inventoryItemAPIPath = "dcim/inventory-items/"

type inventoryItemRequest struct {
	Name          string             `json:"name"`
	Device        *int64             `json:"device"`
	Parent        *int64             `json:"parent"`
	Label         string             `json:"label,omitempty"`
	Role          *int64             `json:"role"`
	Manufacturer  *int64             `json:"manufacturer"`
	PartId        string             `json:"part_id,omitempty"`
	Serial        string             `json:"serial,omitempty"`
	AssetTag      *string            `json:"asset_tag"`
	Discovered    bool               `json:"discovered"`
	Description   string             `json:"description,omitempty"`
	ComponentType *string            `json:"component_type"`
	ComponentID   *int64             `json:"component_id"`
	CustomFields  interface{}        `json:"custom_fields,omitempty"`
	Tags          []netboxapi.TagRef `json:"tags"`
}

type inventoryItemResponse struct {
	Name          string                 `json:"name"`
	Device        *netboxapi.Ref         `json:"device"`
	Parent        *int64                 `json:"parent"`
	Label         string                 `json:"label"`
	Role          *netboxapi.Ref         `json:"role"`
	Manufacturer  *netboxapi.Ref         `json:"manufacturer"`
	PartId        string                 `json:"part_id"`
	Serial        string                 `json:"serial"`
	AssetTag      *string                `json:"asset_tag"`
	Discovered    bool                   `json:"discovered"`
	Description   string                 `json:"description"`
	ComponentType *string                `json:"component_type"`
	ComponentID   *int64                 `json:"component_id"`
	CustomFields  map[string]interface{} `json:"custom_fields"`
	Tags          []netboxapi.TagRef     `json:"tags"`
}

func inventoryItemRequestFromResourceData(api *providerState, d *schema.ResourceData) (*inventoryItemRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &inventoryItemRequest{
		Name:         name,
		Device:       int64ToPtr(int64(d.Get("device_id").(int))),
		Parent:       getOptionalInt(d, "parent_id"),
		Label:        getOptionalStr(d, "label", true),
		Role:         getOptionalInt(d, "role_id"),
		Manufacturer: getOptionalInt(d, "manufacturer_id"),
		PartId:       getOptionalStr(d, "part_id", true),
		Serial:       getOptionalStr(d, "serial", true),
		AssetTag:     nullableString(getOptionalStr(d, "asset_tag", false)),
		Discovered:   d.Get("discovered").(bool),
		Description:  getOptionalStr(d, "description", true),

		Tags: tags,
	}

	if componentType := getOptionalStr(d, "component_type", false); componentType != "" {
		data.ComponentType = &componentType
		data.ComponentID = getOptionalInt(d, "component_id")
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxInventoryItemCreate(d *schema.ResourceData, m interface{}) error {
	data, err := inventoryItemRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, inventoryItemAPIPath, data, resourceNetboxInventoryItemRead)
}

func resourceNetboxInventoryItemRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, inventoryItemAPIPath, func(r *inventoryItemResponse) {
		d.Set("name", r.Name)
		d.Set("device_id", netboxapi.RefID(r.Device))
		d.Set("parent_id", r.Parent)
		d.Set("label", r.Label)
		d.Set("role_id", netboxapi.RefID(r.Role))
		d.Set("manufacturer_id", netboxapi.RefID(r.Manufacturer))
		d.Set("part_id", r.PartId)
		d.Set("serial", r.Serial)
		d.Set("asset_tag", r.AssetTag)
		d.Set("discovered", r.Discovered)
		d.Set("description", r.Description)

		d.Set("component_type", r.ComponentType)
		d.Set("component_id", r.ComponentID)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxInventoryItemUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := inventoryItemRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, inventoryItemAPIPath, data, resourceNetboxInventoryItemRead)
}

func resourceNetboxInventoryItemDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, inventoryItemAPIPath)
}
