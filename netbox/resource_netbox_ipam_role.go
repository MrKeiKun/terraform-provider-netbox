package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxIpamRole() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxIpamRoleCreate,
		Read:   resourceNetboxIpamRoleRead,
		Update: resourceNetboxIpamRoleUpdate,
		Delete: resourceNetboxIpamRoleDelete,

		Description: `:meta:subcategory:IP Address Management (IPAM):From the [official documentation](https://docs.netbox.dev/en/stable/features/ipam/#prefixvlan-roles):

> A role indicates the function of a prefix or VLAN. For example, you might define Data, Voice, and Security roles. Generally, a prefix will be assigned the same functional role as the VLAN to which it is assigned (if any).`,

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
			"weight": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 32767),
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
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const ipamRoleAPIPath = "ipam/roles/"

type ipamRoleRequest struct {
	Name         string             `json:"name"`
	Slug         string             `json:"slug"`
	Weight       *int64             `json:"weight,omitempty"`
	Description  string             `json:"description,omitempty"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
}

type ipamRoleResponse struct {
	Name         *string                `json:"name"`
	Slug         *string                `json:"slug"`
	Weight       *int64                 `json:"weight"`
	Description  string                 `json:"description"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func ipamRoleRequestFromResourceData(d *schema.ResourceData) *ipamRoleRequest {
	name := d.Get("name").(string)
	weight := int64(d.Get("weight").(int))
	return &ipamRoleRequest{
		CustomFields: customFieldsRequest(d),
		Comments:     getOptionalStr(d, "comments", true),
		Owner:        getOptionalInt(d, "owner_id"),
		Name:         name,
		Slug:         slugFromResourceData(d, name),
		Weight:       &weight,
		Description:  d.Get("description").(string),
		Tags:         []netboxapi.TagRef{},
	}
}

func resourceNetboxIpamRoleCreate(d *schema.ResourceData, m interface{}) error {
	return createObject(d, m, ipamRoleAPIPath, ipamRoleRequestFromResourceData(d), resourceNetboxIpamRoleUpdate)
}

func resourceNetboxIpamRoleRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, ipamRoleAPIPath, func(res *ipamRoleResponse) {
		d.Set(customFieldsKey, getCustomFields(res.CustomFields))
		d.Set("comments", res.Comments)
		d.Set("owner_id", netboxapi.RefID(res.Owner))
		if res.Name != nil {
			d.Set("name", res.Name)
		}
		if res.Slug != nil {
			d.Set("slug", res.Slug)
		}
		if res.Weight != nil {
			d.Set("weight", res.Weight)
		}
		if res.Description != "" {
			d.Set("description", res.Description)
		}
	})
}

func resourceNetboxIpamRoleUpdate(d *schema.ResourceData, m interface{}) error {
	return putObject(d, m, ipamRoleAPIPath, ipamRoleRequestFromResourceData(d), resourceNetboxIpamRoleRead)
}

func resourceNetboxIpamRoleDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, ipamRoleAPIPath)
}
