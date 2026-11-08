package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxModule() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxModuleCreate,
		Read:   resourceNetboxModuleRead,
		Update: resourceNetboxModuleUpdate,
		Delete: resourceNetboxModuleDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/module/):

> A module is a field-replaceable hardware component installed within a device which houses its own child components. The most common example is a chassis-based router or switch.

Similar to devices, modules are instantiated from module types, and any components associated with the module type are automatically instantiated on the new model. Each module must be installed within a module bay on a device, and each module bay may have only one module installed in it.`,

		Schema: map[string]*schema.Schema{
			"device_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"module_bay_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"module_type_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"status": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "One of [offline, active, planned, staged, failed, decommissioning]",
			},
			"serial": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"asset_tag": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"comments": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"owner_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"replicate_components": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Automatically populate components associated with this module type (default: true)",
			},
			"adopt_components": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Adopt already existing components",
			},
			tagsKey:         tagsSchema,
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const moduleAPIPath = "dcim/modules/"

type moduleRequest struct {
	Device              *int64             `json:"device"`
	ModuleBay           *int64             `json:"module_bay"`
	ModuleType          *int64             `json:"module_type"`
	Status              string             `json:"status,omitempty"`
	Serial              string             `json:"serial,omitempty"`
	AssetTag            *string            `json:"asset_tag"`
	Description         string             `json:"description,omitempty"`
	Comments            string             `json:"comments,omitempty"`
	Owner               *int64             `json:"owner"`
	ReplicateComponents *bool              `json:"replicate_components,omitempty"`
	AdoptComponents     *bool              `json:"adopt_components,omitempty"`
	CustomFields        interface{}        `json:"custom_fields,omitempty"`
	Tags                []netboxapi.TagRef `json:"tags"`
}

type moduleResponse struct {
	Device              *netboxapi.Ref         `json:"device"`
	ModuleBay           *netboxapi.Ref         `json:"module_bay"`
	ModuleType          *netboxapi.Ref         `json:"module_type"`
	Status              *netboxapi.Choice      `json:"status"`
	Serial              string                 `json:"serial"`
	AssetTag            *string                `json:"asset_tag"`
	Description         string                 `json:"description"`
	Comments            string                 `json:"comments"`
	Owner               *netboxapi.Ref         `json:"owner"`
	ReplicateComponents bool                   `json:"replicate_components"`
	AdoptComponents     bool                   `json:"adopt_components"`
	CustomFields        map[string]interface{} `json:"custom_fields"`
	Tags                []netboxapi.TagRef     `json:"tags"`
}

func moduleRequestFromResourceData(api *providerState, d *schema.ResourceData) (*moduleRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &moduleRequest{
		AdoptComponents:     getOptionalBool(d, "adopt_components"),
		ReplicateComponents: getOptionalBool(d, "replicate_components"),
		Owner:               getOptionalInt(d, "owner_id"),
		Device:              int64ToPtr(int64(d.Get("device_id").(int))),
		ModuleBay:           int64ToPtr(int64(d.Get("module_bay_id").(int))),
		ModuleType:          int64ToPtr(int64(d.Get("module_type_id").(int))),
		Status:              d.Get("status").(string),
		Serial:              getOptionalStr(d, "serial", true),
		AssetTag:            nullableString(getOptionalStr(d, "asset_tag", false)),
		Description:         getOptionalStr(d, "description", true),
		Comments:            getOptionalStr(d, "comments", true),
		Tags:                tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxModuleCreate(d *schema.ResourceData, m interface{}) error {
	data, err := moduleRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, moduleAPIPath, data, resourceNetboxModuleRead)
}

func resourceNetboxModuleRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, moduleAPIPath, func(r *moduleResponse) {
		d.Set("adopt_components", r.AdoptComponents)
		d.Set("replicate_components", r.ReplicateComponents)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("device_id", netboxapi.RefID(r.Device))
		d.Set("module_bay_id", netboxapi.RefID(r.ModuleBay))
		d.Set("module_type_id", netboxapi.RefID(r.ModuleType))
		d.Set("status", netboxapi.ChoiceValue(r.Status))
		d.Set("serial", r.Serial)
		d.Set("asset_tag", r.AssetTag)
		d.Set("description", r.Description)
		d.Set("comments", r.Comments)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxModuleUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := moduleRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, moduleAPIPath, data, resourceNetboxModuleRead)
}

func resourceNetboxModuleDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, moduleAPIPath)
}
