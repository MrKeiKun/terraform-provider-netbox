package netbox

import (
	"context"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var resourceNetboxSiteStatusOptions = []string{"planned", "staging", "active", "decommissioning", "retired"}

func resourceNetboxSite() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxSiteCreate,
		Read:   resourceNetboxSiteRead,
		Update: resourceNetboxSiteUpdate,
		Delete: resourceNetboxSiteDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/features/sites-and-racks/#sites):

> How you choose to employ sites when modeling your network may vary depending on the nature of your organization, but generally a site will equate to a building or campus. For example, a chain of banks might create a site to represent each of its branches, a site for its corporate headquarters, and two additional sites for its presence in two colocation facilities.
>
> Each site must be assigned a unique name and may optionally be assigned to a region and/or tenant.`,

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
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "active",
				ValidateFunc: validation.StringInSlice(resourceNetboxSiteStatusOptions, false),
				Description:  buildValidValueDescription(resourceNetboxSiteStatusOptions),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 200),
			},
			"comments": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"facility": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 50),
			},
			"longitude": {
				Type:     schema.TypeFloat,
				Optional: true,
			},
			"latitude": {
				Type:     schema.TypeFloat,
				Optional: true,
			},
			"physical_address": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 200),
			},
			"shipping_address": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 200),
			},
			"region_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"group_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"tenant_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			tagsKey: tagsSchema,
			"timezone": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"asn_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const siteAPIPath = "dcim/sites/"

// siteRequest is the body sent to NetBox when writing a site.
type siteRequest struct {
	Name            string             `json:"name"`
	Slug            string             `json:"slug"`
	Status          string             `json:"status,omitempty"`
	Description     string             `json:"description,omitempty"`
	Comments        string             `json:"comments,omitempty"`
	Facility        string             `json:"facility,omitempty"`
	Latitude        *float64           `json:"latitude"`
	Longitude       *float64           `json:"longitude"`
	PhysicalAddress string             `json:"physical_address,omitempty"`
	ShippingAddress string             `json:"shipping_address,omitempty"`
	Region          *int64             `json:"region,omitempty"`
	Group           *int64             `json:"group,omitempty"`
	Tenant          *int64             `json:"tenant,omitempty"`
	TimeZone        *string            `json:"time_zone,omitempty"`
	Asns            []int64            `json:"asns"`
	Tags            []netboxapi.TagRef `json:"tags"`
	CustomFields    interface{}        `json:"custom_fields,omitempty"`
}

// siteResponse is the subset of the NetBox site object this resource reads.
type siteResponse struct {
	ID              int64                  `json:"id"`
	Name            string                 `json:"name"`
	Slug            string                 `json:"slug"`
	Status          *netboxapi.Choice      `json:"status"`
	Description     string                 `json:"description"`
	Comments        string                 `json:"comments"`
	Facility        string                 `json:"facility"`
	Latitude        *float64               `json:"latitude"`
	Longitude       *float64               `json:"longitude"`
	PhysicalAddress string                 `json:"physical_address"`
	ShippingAddress string                 `json:"shipping_address"`
	Region          *netboxapi.Ref         `json:"region"`
	Group           *netboxapi.Ref         `json:"group"`
	Tenant          *netboxapi.Ref         `json:"tenant"`
	TimeZone        *string                `json:"time_zone"`
	Asns            []netboxapi.Ref        `json:"asns"`
	Tags            []netboxapi.TagRef     `json:"tags"`
	CustomFields    map[string]interface{} `json:"custom_fields"`
}

func siteRequestFromResourceData(api *providerState, d *schema.ResourceData, isUpdate bool) (*siteRequest, error) {
	data := siteRequest{}

	data.Name = d.Get("name").(string)

	// Default slug to generated slug if not given
	data.Slug = getSlug(data.Name)
	if v, ok := d.GetOk("slug"); ok {
		data.Slug = v.(string)
	}

	data.Status = d.Get("status").(string)

	// On update, an empty value that changed is sent as a space so NetBox clears it.
	clearable := func(key string) string {
		if v, ok := d.GetOk(key); ok {
			return v.(string)
		}
		if isUpdate && d.HasChange(key) {
			return " "
		}
		return ""
	}
	data.Description = clearable("description")
	data.Comments = clearable("comments")
	data.PhysicalAddress = clearable("physical_address")
	data.ShippingAddress = clearable("shipping_address")

	if facility, ok := d.GetOk("facility"); ok {
		data.Facility = facility.(string)
	}
	if v, ok := d.GetOk("latitude"); ok {
		data.Latitude = float64ToPtr(v.(float64))
	}
	if v, ok := d.GetOk("longitude"); ok {
		data.Longitude = float64ToPtr(v.(float64))
	}
	if v, ok := d.GetOk("region_id"); ok {
		data.Region = int64ToPtr(int64(v.(int)))
	}
	if v, ok := d.GetOk("group_id"); ok {
		data.Group = int64ToPtr(int64(v.(int)))
	}
	if v, ok := d.GetOk("tenant_id"); ok {
		data.Tenant = int64ToPtr(int64(v.(int)))
	}
	if v, ok := d.GetOk("timezone"); ok {
		data.TimeZone = strToPtr(v.(string))
	}

	data.Asns = []int64{}
	if v, ok := d.GetOk("asn_ids"); ok {
		data.Asns = toInt64List(v)
	}

	var err error
	data.Tags, err = getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}

	if cf, ok := d.GetOk(customFieldsKey); ok {
		data.CustomFields = cf
	}

	return &data, nil
}

func resourceNetboxSiteCreate(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)

	data, err := siteRequestFromResourceData(api, d, false)
	if err != nil {
		return err
	}

	res, err := netboxapi.Create[siteResponse](context.Background(), api.rest, siteAPIPath, data)
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(res.ID, 10))

	return resourceNetboxSiteRead(d, m)
}

func resourceNetboxSiteRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)

	site, err := netboxapi.Get[siteResponse](context.Background(), api.rest, siteAPIPath, id)
	if err != nil {
		if netboxapi.IsNotFound(err) {
			// If the ID is updated to blank, this tells Terraform the resource no longer exists (maybe it was destroyed out of band). Just like the destroy callback, the Read function should gracefully handle this case. https://www.terraform.io/docs/extend/writing-custom-providers.html
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("name", site.Name)
	d.Set("slug", site.Slug)
	d.Set("status", netboxapi.ChoiceValue(site.Status))
	d.Set("description", site.Description)
	d.Set("comments", site.Comments)
	d.Set("facility", site.Facility)
	d.Set("longitude", site.Longitude)
	d.Set("latitude", site.Latitude)
	d.Set("physical_address", site.PhysicalAddress)
	d.Set("shipping_address", site.ShippingAddress)
	d.Set("timezone", site.TimeZone)
	d.Set("asn_ids", netboxapi.IDs(site.Asns))
	d.Set("region_id", netboxapi.RefID(site.Region))
	d.Set("group_id", netboxapi.RefID(site.Group))
	d.Set("tenant_id", netboxapi.RefID(site.Tenant))

	if cf := getCustomFields(site.CustomFields); cf != nil {
		d.Set(customFieldsKey, cf)
	}
	api.readTagRefs(d, site.Tags)

	return nil
}

func resourceNetboxSiteUpdate(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)

	data, err := siteRequestFromResourceData(api, d, true)
	if err != nil {
		return err
	}

	// The generated client used PATCH here, so keep that.
	if _, err := netboxapi.Patch[siteResponse](context.Background(), api.rest, siteAPIPath, id, data); err != nil {
		return err
	}

	return resourceNetboxSiteRead(d, m)
}

func resourceNetboxSiteDelete(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)

	if err := netboxapi.Delete(context.Background(), api.rest, siteAPIPath, id); err != nil {
		if netboxapi.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return err
	}
	return nil
}
