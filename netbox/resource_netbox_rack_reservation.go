package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxRackReservation() *schema.Resource {
	return &schema.Resource{
		CreateContext: withContext(resourceNetboxRackReservationCreate),
		ReadContext:   withContext(resourceNetboxRackReservationRead),
		UpdateContext: withContext(resourceNetboxRackReservationUpdate),
		DeleteContext: withContext(resourceNetboxRackReservationDelete),

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/rackreservation/):

> Users can reserve specific units within a rack for future use. An arbitrary set of units within a rack can be associated with a single reservation, but reservations cannot span multiple racks. A description is required for each reservation, reservations may optionally be associated with a specific tenant.`,

		Schema: map[string]*schema.Schema{
			"rack_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the rack associated with the rack reservation.",
				Required:    true,
			},
			"units": {
				Type:        schema.TypeSet,
				Description: "The units of the rack reservation.",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Required: true,
				Set:      schema.HashInt,
			},
			"user_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the user associated with the rack reservation.",
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the rack reservation.",
				Required:    true,
			},
			"tenant_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the tenant associated with the rack reservation.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the rack reservation.",
				Optional:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the rack reservation.",
				Optional:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "One of [pending, active, stale]",
			},
			customFieldsKey: customFieldsSchema,
			tagsKey:         tagsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const rackReservationAPIPath = "dcim/rack-reservations/"

type rackReservationRequest struct {
	Rack         *int64             `json:"rack"`
	Units        []*int64           `json:"units"`
	User         *int64             `json:"user"`
	Description  string             `json:"description"`
	Tenant       *int64             `json:"tenant,omitempty"`
	Comments     string             `json:"comments,omitempty"`
	Owner        *int64             `json:"owner"`
	Status       string             `json:"status,omitempty"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
}

type rackReservationResponse struct {
	Rack         *netboxapi.Ref         `json:"rack"`
	Units        []int64                `json:"units"`
	User         *netboxapi.Ref         `json:"user"`
	Description  string                 `json:"description"`
	Tenant       *netboxapi.Ref         `json:"tenant"`
	Comments     string                 `json:"comments"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Status       *netboxapi.Choice      `json:"status"`
	CustomFields map[string]interface{} `json:"custom_fields"`
	Tags         []netboxapi.TagRef     `json:"tags"`
}

func rackReservationRequestFromResourceData(api *providerState, d *schema.ResourceData) *rackReservationRequest {
	// A failed tag lookup leaves the tags empty, same as before.
	tags, _ := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	return &rackReservationRequest{
		CustomFields: customFieldsRequest(d),
		Status:       d.Get("status").(string),
		Owner:        getOptionalInt(d, "owner_id"),
		Rack:         getOptionalInt(d, "rack_id"),
		Units:        toInt64PtrList(d.Get("units")),
		User:         getOptionalInt(d, "user_id"),
		Description:  getOptionalStr(d, "description", false),
		Tenant:       getOptionalInt(d, "tenant_id"),
		Comments:     getOptionalStr(d, "comments", false),
		Tags:         tags,
	}
}

func resourceNetboxRackReservationCreate(d *schema.ResourceData, m interface{}) error {
	return createObject(d, m, rackReservationAPIPath, rackReservationRequestFromResourceData(m.(*providerState), d), resourceNetboxRackReservationRead)
}

func resourceNetboxRackReservationRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, rackReservationAPIPath, func(rackRes *rackReservationResponse) {
		d.Set(customFieldsKey, getCustomFields(rackRes.CustomFields))
		d.Set("status", netboxapi.ChoiceValue(rackRes.Status))
		d.Set("owner_id", netboxapi.RefID(rackRes.Owner))
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
