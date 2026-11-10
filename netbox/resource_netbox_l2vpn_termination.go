package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxL2VPNTermination() *schema.Resource {
	return &schema.Resource{
		CreateContext: withContext(resourceNetboxL2VPNTerminationCreate),
		ReadContext:   withContext(resourceNetboxL2VPNTerminationRead),
		UpdateContext: withContext(resourceNetboxL2VPNTerminationUpdate),
		DeleteContext: withContext(resourceNetboxL2VPNTerminationDelete),

		Description: `:meta:subcategory:VPN Tunnels:From the [official documentation](https://docs.netbox.dev/en/stable/models/vpn/l2vpntermination/):

> An L2VPN termination attaches a VLAN, an interface or a VM interface to an L2VPN.`,

		Schema: map[string]*schema.Schema{
			"l2vpn_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the l2vpn associated with the l2vpn termination.",
				Required:    true,
			},
			"assigned_object_type": {
				Type:        schema.TypeString,
				Description: "The assigned object type of the l2vpn termination.",
				Required:    true,
			},
			"assigned_object_id": {
				Type:         schema.TypeInt,
				Description:  "The ID of the assigned object associated with the l2vpn termination.",
				Required:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
			tagsKey:         tagsSchema,
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const l2VPNTerminationAPIPath = "vpn/l2vpn-terminations/"

type l2VPNTerminationRequest struct {
	L2vpn              *int64             `json:"l2vpn"`
	AssignedObjectType string             `json:"assigned_object_type"`
	AssignedObjectID   int64              `json:"assigned_object_id"`
	Tags               []netboxapi.TagRef `json:"tags"`
	CustomFields       interface{}        `json:"custom_fields,omitempty"`
}

type l2VPNTerminationResponse struct {
	L2vpn              *netboxapi.Ref         `json:"l2vpn"`
	AssignedObjectType string                 `json:"assigned_object_type"`
	AssignedObjectID   *int64                 `json:"assigned_object_id"`
	Tags               []netboxapi.TagRef     `json:"tags"`
	CustomFields       map[string]interface{} `json:"custom_fields"`
}

func l2VPNTerminationRequestFromResourceData(api *providerState, d *schema.ResourceData) (*l2VPNTerminationRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &l2VPNTerminationRequest{
		L2vpn:              int64ToPtr(int64(d.Get("l2vpn_id").(int))),
		AssignedObjectType: d.Get("assigned_object_type").(string),
		AssignedObjectID:   int64(d.Get("assigned_object_id").(int)),
		Tags:               tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxL2VPNTerminationCreate(d *schema.ResourceData, m interface{}) error {
	data, err := l2VPNTerminationRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, l2VPNTerminationAPIPath, data, resourceNetboxL2VPNTerminationRead)
}

func resourceNetboxL2VPNTerminationRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, l2VPNTerminationAPIPath, func(r *l2VPNTerminationResponse) {
		d.Set("l2vpn_id", netboxapi.RefID(r.L2vpn))
		d.Set("assigned_object_type", r.AssignedObjectType)
		d.Set("assigned_object_id", r.AssignedObjectID)
		api.readTagRefs(d, r.Tags)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
	})
}

func resourceNetboxL2VPNTerminationUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := l2VPNTerminationRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, l2VPNTerminationAPIPath, data, resourceNetboxL2VPNTerminationRead)
}

func resourceNetboxL2VPNTerminationDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, l2VPNTerminationAPIPath)
}
