package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxCoolingSource() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxCoolingSourceCreate,
		Read:   resourceNetboxCoolingSourceRead,
		Update: resourceNetboxCoolingSourceUpdate,
		Delete: resourceNetboxCoolingSourceDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/coolingsource/):

> A cooling source is a facility device such as a chiller or a cooling tower that supplies cooling to cooling feeds.

Requires NetBox v4.7.0 or newer.`,

		Schema: map[string]*schema.Schema{
			"site_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"location_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "One of [chiller, cooling-tower, dry-cooler, crac, crah]",
				ValidateFunc: validation.StringInSlice([]string{"chiller", "cooling-tower", "dry-cooler", "crac", "crah"}, false),
			},
			"status": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "One of [offline, active, planned, failed]",
			},
			"fluid_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "One of [water, water-glycol, dielectric, refrigerant]",
				ValidateFunc: validation.StringInSlice([]string{"water", "water-glycol", "dielectric", "refrigerant"}, false),
			},
			"cooling_capacity": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Description: "Total rated cooling capacity (kW)",
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

const coolingSourceAPIPath = "dcim/cooling-sources/"

type coolingSourceRequest struct {
	Site            *int64             `json:"site"`
	Location        *int64             `json:"location"`
	Name            string             `json:"name"`
	Type            string             `json:"type,omitempty"`
	Status          string             `json:"status,omitempty"`
	FluidType       *string            `json:"fluid_type,omitempty"`
	CoolingCapacity *float64           `json:"cooling_capacity"`
	Description     string             `json:"description,omitempty"`
	Owner           *int64             `json:"owner"`
	Comments        string             `json:"comments,omitempty"`
	Tags            []netboxapi.TagRef `json:"tags"`
	CustomFields    interface{}        `json:"custom_fields,omitempty"`
}

type coolingSourceResponse struct {
	Site            *netboxapi.Ref         `json:"site"`
	Location        *netboxapi.Ref         `json:"location"`
	Name            string                 `json:"name"`
	Type            *netboxapi.Choice      `json:"type"`
	Status          *netboxapi.Choice      `json:"status"`
	FluidType       *netboxapi.Choice      `json:"fluid_type"`
	CoolingCapacity *float64               `json:"cooling_capacity"`
	Description     string                 `json:"description"`
	Owner           *netboxapi.Ref         `json:"owner"`
	Comments        string                 `json:"comments"`
	Tags            []netboxapi.TagRef     `json:"tags"`
	CustomFields    map[string]interface{} `json:"custom_fields"`
}

func coolingSourceRequestFromResourceData(api *providerState, d *schema.ResourceData) (*coolingSourceRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &coolingSourceRequest{
		Site:            int64ToPtr(int64(d.Get("site_id").(int))),
		Location:        getOptionalInt(d, "location_id"),
		Name:            d.Get("name").(string),
		Type:            d.Get("type").(string),
		Status:          d.Get("status").(string),
		FluidType:       getOptionalStrClear(d, "fluid_type"),
		CoolingCapacity: getOptionalFloat(d, "cooling_capacity"),
		Description:     getOptionalStr(d, "description", true),
		Owner:           getOptionalInt(d, "owner_id"),
		Comments:        getOptionalStr(d, "comments", true),
		Tags:            tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxCoolingSourceCreate(d *schema.ResourceData, m interface{}) error {
	data, err := coolingSourceRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, coolingSourceAPIPath, data, resourceNetboxCoolingSourceRead)
}

func resourceNetboxCoolingSourceRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, coolingSourceAPIPath, func(r *coolingSourceResponse) {
		d.Set("site_id", netboxapi.RefID(r.Site))
		d.Set("location_id", netboxapi.RefID(r.Location))
		d.Set("name", r.Name)
		d.Set("type", netboxapi.ChoiceValue(r.Type))
		d.Set("status", netboxapi.ChoiceValue(r.Status))
		d.Set("fluid_type", netboxapi.ChoiceValue(r.FluidType))
		d.Set("cooling_capacity", r.CoolingCapacity)
		d.Set("description", r.Description)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("comments", r.Comments)
		api.readTagRefs(d, r.Tags)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
	})
}

func resourceNetboxCoolingSourceUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := coolingSourceRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, coolingSourceAPIPath, data, resourceNetboxCoolingSourceRead)
}

func resourceNetboxCoolingSourceDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, coolingSourceAPIPath)
}
