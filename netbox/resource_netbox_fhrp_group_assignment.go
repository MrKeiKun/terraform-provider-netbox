package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxFhrpGroupAssignment() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxFhrpGroupAssignmentCreate,
		Read:   resourceNetboxFhrpGroupAssignmentRead,
		Update: resourceNetboxFhrpGroupAssignmentUpdate,
		Delete: resourceNetboxFhrpGroupAssignmentDelete,

		Description: `:meta:subcategory:IP Address Management (IPAM):From the [official documentation](https://netboxlabs.com/docs/netbox/models/ipam/fhrpgroup/):`,

		Schema: map[string]*schema.Schema{
			"group_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Group ID",
			},
			"interface_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Interface ID",
			},
			"priority": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Priority",
			},
			"interface_type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Interface type",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const fhrpGroupAssignmentAPIPath = "ipam/fhrp-group-assignments/"

type fhrpGroupAssignmentRequest struct {
	Group         *int64 `json:"group"`
	InterfaceId   *int64 `json:"interface_id"`
	InterfaceType string `json:"interface_type"`
	Priority      *int64 `json:"priority"`
}

type fhrpGroupAssignmentResponse struct {
	Group         *netboxapi.Ref `json:"group"`
	InterfaceId   *int64         `json:"interface_id"`
	InterfaceType string         `json:"interface_type"`
	Priority      *int64         `json:"priority"`
}

func fhrpGroupAssignmentRequestFromResourceData(d *schema.ResourceData) *fhrpGroupAssignmentRequest {
	data := &fhrpGroupAssignmentRequest{
		Group:         int64ToPtr(int64(d.Get("group_id").(int))),
		InterfaceId:   int64ToPtr(int64(d.Get("interface_id").(int))),
		InterfaceType: d.Get("interface_type").(string),
		Priority:      int64ToPtr(int64(d.Get("priority").(int))),
	}
	return data
}

func resourceNetboxFhrpGroupAssignmentCreate(d *schema.ResourceData, m interface{}) error {
	return createObject(d, m, fhrpGroupAssignmentAPIPath, fhrpGroupAssignmentRequestFromResourceData(d), resourceNetboxFhrpGroupAssignmentRead)
}

func resourceNetboxFhrpGroupAssignmentRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, fhrpGroupAssignmentAPIPath, func(r *fhrpGroupAssignmentResponse) {
		d.Set("group_id", netboxapi.RefID(r.Group))
		d.Set("interface_id", r.InterfaceId)
		d.Set("interface_type", r.InterfaceType)
		d.Set("priority", r.Priority)
	})
}

func resourceNetboxFhrpGroupAssignmentUpdate(d *schema.ResourceData, m interface{}) error {
	return putObject(d, m, fhrpGroupAssignmentAPIPath, fhrpGroupAssignmentRequestFromResourceData(d), resourceNetboxFhrpGroupAssignmentRead)
}

func resourceNetboxFhrpGroupAssignmentDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, fhrpGroupAssignmentAPIPath)
}
