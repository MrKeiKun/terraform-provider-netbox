package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxRouteTarget() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxRouteTargetCreate,
		Read:   resourceNetboxRouteTargetRead,
		Update: resourceNetboxRouteTargetUpdate,
		Delete: resourceNetboxRouteTargetDelete,

		Description: `:meta:subcategory:IP Address Management (IPAM):From the [official documentation](https://docs.netbox.dev/en/stable/models/ipam/routetarget/):

> A route target is a particular type of extended BGP community used to control the redistribution of routes among VRF tables in a network. Route targets can be assigned to individual VRFs in NetBox as import or export targets (or both) to model this exchange in an L3VPN. Each route target must be given a unique name, which should be in a format prescribed by RFC 4364, similar to a VR route distinguisher.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 21),
			},
			"tenant_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 200),
			},
			tagsKey: tagsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const routeTargetAPIPath = "ipam/route-targets/"

type routeTargetRequest struct {
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Tenant      *int64             `json:"tenant,omitempty"`
	Tags        []netboxapi.TagRef `json:"tags"`
}

type routeTargetResponse struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Tenant      *netboxapi.Ref     `json:"tenant"`
	Tags        []netboxapi.TagRef `json:"tags"`
}

func routeTargetRequestFromResourceData(d *schema.ResourceData) *routeTargetRequest {
	name := d.Get("name").(string)
	data := &routeTargetRequest{
		Name:        name,
		Description: d.Get("description").(string),
		Tags:        []netboxapi.TagRef{},
	}
	if v := int64(d.Get("tenant_id").(int)); v != 0 {
		data.Tenant = &v
	}
	return data
}

func resourceNetboxRouteTargetCreate(d *schema.ResourceData, m interface{}) error {
	return createObject(d, m, routeTargetAPIPath, routeTargetRequestFromResourceData(d), resourceNetboxRouteTargetUpdate)
}

func resourceNetboxRouteTargetRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, routeTargetAPIPath, func(r *routeTargetResponse) {
		d.Set("name", r.Name)
		if r.Description != "" {
			d.Set("description", r.Description)
		}
		if r.Tenant != nil {
			d.Set("tenant_id", r.Tenant.ID)
		}
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxRouteTargetUpdate(d *schema.ResourceData, m interface{}) error {
	return putObject(d, m, routeTargetAPIPath, routeTargetRequestFromResourceData(d), resourceNetboxRouteTargetRead)
}

func resourceNetboxRouteTargetDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, routeTargetAPIPath)
}
