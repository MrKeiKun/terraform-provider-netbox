package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxRackReservation() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxRackReservationCreate,
		Read:   resourceNetboxRackReservationRead,
		Update: resourceNetboxRackReservationUpdate,
		Delete: resourceNetboxRackReservationDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/rackreservation/):

> Users can reserve specific units within a rack for future use. An arbitrary set of units within a rack can be associated with a single reservation, but reservations cannot span multiple racks. A description is required for each reservation, reservations may optionally be associated with a specific tenant.`,

		Schema: map[string]*schema.Schema{
			"rack_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"units": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Required: true,
				Set:      schema.HashInt,
			},
			"user_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tenant_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"comments": {
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

const rackReservationAPIPath = "dcim/rack-reservations/"

type rackReservationRequest struct {
	Rack        *int64             `json:"rack"`
	Units       []*int64           `json:"units"`
	User        *int64             `json:"user"`
	Description string             `json:"description"`
	Tenant      *int64             `json:"tenant,omitempty"`
	Comments    string             `json:"comments,omitempty"`
	Tags        []netboxapi.TagRef `json:"tags"`
}

type rackReservationResponse struct {
	Rack        *netboxapi.Ref     `json:"rack"`
	Units       []int64            `json:"units"`
	User        *netboxapi.Ref     `json:"user"`
	Description string             `json:"description"`
	Tenant      *netboxapi.Ref     `json:"tenant"`
	Comments    string             `json:"comments"`
	Tags        []netboxapi.TagRef `json:"tags"`
}

func rackReservationRequestFromResourceData(api *providerState, d *schema.ResourceData) *rackReservationRequest {
	// A failed tag lookup leaves the tags empty, same as before.
	tags, _ := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	return &rackReservationRequest{
		Rack:        getOptionalInt(d, "rack_id"),
		Units:       toInt64PtrList(d.Get("units")),
		User:        getOptionalInt(d, "user_id"),
		Description: getOptionalStr(d, "description", false),
		Tenant:      getOptionalInt(d, "tenant_id"),
		Comments:    getOptionalStr(d, "comments", false),
		Tags:        tags,
	}
}

func resourceNetboxRackReservationCreate(d *schema.ResourceData, m interface{}) error {
	return createObject(d, m, rackReservationAPIPath, rackReservationRequestFromResourceData(m.(*providerState), d), resourceNetboxRackReservationRead)
}

func resourceNetboxRackReservationRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, rackReservationAPIPath, func(rackRes *rackReservationResponse) {
		if rackRes.Rack != nil {
			d.Set("rack_id", rackRes.Rack.ID)
		}
		units := []int{}
		for _, unit := range rackRes.Units {
			units = append(units, int(unit))
		}
		d.Set("units", units)
		if rackRes.User != nil {
			d.Set("user_id", rackRes.User.ID)
		}
		d.Set("description", rackRes.Description)
		if rackRes.Tenant != nil {
			d.Set("tenant_id", rackRes.Tenant.ID)
		}
		d.Set("comments", rackRes.Comments)
		api.readTagRefs(d, rackRes.Tags)
	})
}

func resourceNetboxRackReservationUpdate(d *schema.ResourceData, m interface{}) error {
	return patchObject(d, m, rackReservationAPIPath, rackReservationRequestFromResourceData(m.(*providerState), d), resourceNetboxRackReservationRead)
}

func resourceNetboxRackReservationDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, rackReservationAPIPath)
}
