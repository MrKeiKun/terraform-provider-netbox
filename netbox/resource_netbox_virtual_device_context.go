package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxVirtualDeviceContext() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxVirtualDeviceContextCreate,
		Read:   resourceNetboxVirtualDeviceContextRead,
		Update: resourceNetboxVirtualDeviceContextUpdate,
		Delete: resourceNetboxVirtualDeviceContextDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/virtualdevicecontext/):

> A virtual device context (VDC) is a logical partition of a device, with its own interfaces and addresses.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"device_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"identifier": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 32767),
			},
			"tenant_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"primary_ip4_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"primary_ip6_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"status": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "One of [active, planned, offline]",
				ValidateFunc: validation.StringInSlice([]string{"active", "planned", "offline"}, false),
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

const virtualDeviceContextAPIPath = "dcim/virtual-device-contexts/"

type virtualDeviceContextRequest struct {
	Name         string             `json:"name"`
	Device       *int64             `json:"device"`
	Identifier   *int64             `json:"identifier"`
	Tenant       *int64             `json:"tenant"`
	PrimaryIp4   *int64             `json:"primary_ip4"`
	PrimaryIp6   *int64             `json:"primary_ip6"`
	Status       string             `json:"status,omitempty"`
	Description  string             `json:"description,omitempty"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
}

type virtualDeviceContextResponse struct {
	Name         string                 `json:"name"`
	Device       *netboxapi.Ref         `json:"device"`
	Identifier   *int64                 `json:"identifier"`
	Tenant       *netboxapi.Ref         `json:"tenant"`
	PrimaryIp4   *netboxapi.Ref         `json:"primary_ip4"`
	PrimaryIp6   *netboxapi.Ref         `json:"primary_ip6"`
	Status       *netboxapi.Choice      `json:"status"`
	Description  string                 `json:"description"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	Tags         []netboxapi.TagRef     `json:"tags"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func virtualDeviceContextRequestFromResourceData(api *providerState, d *schema.ResourceData) (*virtualDeviceContextRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &virtualDeviceContextRequest{
		Name:        d.Get("name").(string),
		Device:      int64ToPtr(int64(d.Get("device_id").(int))),
		Identifier:  getOptionalInt(d, "identifier"),
		Tenant:      getOptionalInt(d, "tenant_id"),
		PrimaryIp4:  getOptionalInt(d, "primary_ip4_id"),
		PrimaryIp6:  getOptionalInt(d, "primary_ip6_id"),
		Status:      d.Get("status").(string),
		Description: getOptionalStr(d, "description", true),
		Owner:       getOptionalInt(d, "owner_id"),
		Comments:    getOptionalStr(d, "comments", true),
		Tags:        tags,
	}
	if ct, ok := d.GetOk(customFieldsKey); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxVirtualDeviceContextCreate(d *schema.ResourceData, m interface{}) error {
	data, err := virtualDeviceContextRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, virtualDeviceContextAPIPath, data, resourceNetboxVirtualDeviceContextRead)
}

func resourceNetboxVirtualDeviceContextRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, virtualDeviceContextAPIPath, func(r *virtualDeviceContextResponse) {
		d.Set("name", r.Name)
		d.Set("device_id", netboxapi.RefID(r.Device))
		d.Set("identifier", r.Identifier)
		d.Set("tenant_id", netboxapi.RefID(r.Tenant))
		d.Set("primary_ip4_id", netboxapi.RefID(r.PrimaryIp4))
		d.Set("primary_ip6_id", netboxapi.RefID(r.PrimaryIp6))
		d.Set("status", netboxapi.ChoiceValue(r.Status))
		d.Set("description", r.Description)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("comments", r.Comments)
		api.readTagRefs(d, r.Tags)
		if cf := getCustomFields(r.CustomFields); cf != nil {
			d.Set(customFieldsKey, cf)
		}
	})
}

func resourceNetboxVirtualDeviceContextUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := virtualDeviceContextRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, virtualDeviceContextAPIPath, data, resourceNetboxVirtualDeviceContextRead)
}

func resourceNetboxVirtualDeviceContextDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, virtualDeviceContextAPIPath)
}
