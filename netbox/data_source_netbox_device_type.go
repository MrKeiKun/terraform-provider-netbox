package netbox

import (
	"net/url"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxDeviceType() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxDeviceTypeRead,
		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):Retrieves a single device type from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"is_full_depth": {
				Type:        schema.TypeBool,
				Description: "Whether the device type is full depth.",
				Computed:    true,
			},
			"manufacturer": {
				Type:        schema.TypeString,
				Description: "The manufacturer of the device type.",
				Optional:    true,
			},
			"manufacturer_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the manufacturer associated with the device type.",
				Computed:    true,
			},
			"model": {
				Type:        schema.TypeString,
				Description: "The model of the device type.",
				Optional:    true,
			},
			"part_number": {
				Type:        schema.TypeString,
				Description: "The part number of the device type.",
				Optional:    true,
			},
			"slug": {
				Type:        schema.TypeString,
				Description: "The URL-friendly slug of the device type.",
				Optional:    true,
			},
			"subdevice_role": {
				Type:        schema.TypeString,
				Description: "The subdevice role of the device type.",
				Optional:    true,
			},
			"u_height": {
				Type:        schema.TypeFloat,
				Description: "The height of the device type in rack units.",
				Computed:    true,
			},
		},
	}
}

type deviceTypeDataSourceResponse struct {
	ID            int64             `json:"id"`
	IsFullDepth   bool              `json:"is_full_depth"`
	Manufacturer  *netboxapi.Ref    `json:"manufacturer"`
	Model         string            `json:"model"`
	PartNumber    string            `json:"part_number"`
	SubdeviceRole *netboxapi.Choice `json:"subdevice_role"`
	Slug          string            `json:"slug"`
	UHeight       *float64          `json:"u_height"`
}

func dataSourceNetboxDeviceTypeRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "manufacturer", "manufacturer")
	addStrQuery(query, d, "model", "model")
	addStrQuery(query, d, "part_number", "part_number")
	addStrQuery(query, d, "slug", "slug")

	r, err := findOne[deviceTypeDataSourceResponse](m, "dcim/device-types/", query, "device type")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(r.ID, 10))
	d.Set("is_full_depth", r.IsFullDepth)
	if r.Manufacturer != nil {
		d.Set("manufacturer_id", r.Manufacturer.ID)
	}
	d.Set("model", r.Model)
	d.Set("part_number", r.PartNumber)
	d.Set("subdevice_role", netboxapi.ChoiceValue(r.SubdeviceRole))
	d.Set("slug", r.Slug)
	d.Set("u_height", r.UHeight)
	return nil
}
