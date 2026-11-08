package netbox

import (
	"context"
	"errors"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxRackType() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxRackTypeRead,
		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):Look up a rack type by ` + "`model`" + ` or ` + "`slug`" + `.`,
		Schema: map[string]*schema.Schema{
			"model": {
				Type:         schema.TypeString,
				Description:  "The model of the rack type.",
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"model", "slug"},
			},
			"slug": {
				Type:         schema.TypeString,
				Description:  "The URL-friendly slug of the rack type.",
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"model", "slug"},
			},
			"manufacturer_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the manufacturer associated with the rack type.",
				Computed:    true,
			},
			"form_factor": {
				Type:        schema.TypeString,
				Description: "The form factor of the rack type.",
				Computed:    true,
			},
			"width": {
				Type:        schema.TypeInt,
				Description: "The width of the rack type.",
				Computed:    true,
			},
			"u_height": {
				Type:        schema.TypeInt,
				Description: "The height of the rack type in rack units.",
				Computed:    true,
			},
			"starting_unit": {
				Type:        schema.TypeInt,
				Description: "The starting unit of the rack type.",
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the rack type.",
				Computed:    true,
			},
			"outer_width": {
				Type:        schema.TypeInt,
				Description: "The outer width of the rack type.",
				Computed:    true,
			},
			"outer_depth": {
				Type:        schema.TypeInt,
				Description: "The outer depth of the rack type.",
				Computed:    true,
			},
			"outer_unit": {
				Type:        schema.TypeString,
				Description: "The outer unit of the rack type.",
				Computed:    true,
			},
			"weight": {
				Type:        schema.TypeFloat,
				Description: "The weight of the rack type.",
				Computed:    true,
			},
			"max_weight": {
				Type:        schema.TypeInt,
				Description: "The max weight of the rack type.",
				Computed:    true,
			},
			"weight_unit": {
				Type:        schema.TypeString,
				Description: "The unit of the weight.",
				Computed:    true,
			},
			"mounting_depth_mm": {
				Type:        schema.TypeInt,
				Description: "The mounting depth mm of the rack type.",
				Computed:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the rack type.",
				Computed:    true,
			},
		},
	}
}

type rackTypeDataSourceResponse struct {
	ID            int64                `json:"id"`
	Model         *string              `json:"model"`
	Slug          *string              `json:"slug"`
	Manufacturer  *netboxapi.Ref       `json:"manufacturer"`
	FormFactor    *netboxapi.Choice    `json:"form_factor"`
	Width         *netboxapi.IntChoice `json:"width"`
	UHeight       int64                `json:"u_height"`
	StartingUnit  int64                `json:"starting_unit"`
	Description   string               `json:"description"`
	OuterWidth    *int64               `json:"outer_width"`
	OuterDepth    *int64               `json:"outer_depth"`
	OuterUnit     *netboxapi.Choice    `json:"outer_unit"`
	Weight        *float64             `json:"weight"`
	MaxWeight     *int64               `json:"max_weight"`
	WeightUnit    *netboxapi.Choice    `json:"weight_unit"`
	MountingDepth *int64               `json:"mounting_depth"`
	Comments      string               `json:"comments"`
}

func dataSourceNetboxRackTypeRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)

	model := d.Get("model").(string)
	slug := d.Get("slug").(string)
	if model == "" && slug == "" {
		return errors.New("either 'model' or 'slug' must be specified")
	}

	// The rack-types list endpoint has no server-side model/slug filter, so we
	// list the rack types and match exactly client-side.
	rackTypes, err := netboxapi.List[rackTypeDataSourceResponse](context.Background(), api.rest, "dcim/rack-types/", nil)
	if err != nil {
		return err
	}

	var matches []rackTypeDataSourceResponse
	for _, rackType := range rackTypes {
		if model != "" && (rackType.Model == nil || *rackType.Model != model) {
			continue
		}
		if slug != "" && (rackType.Slug == nil || *rackType.Slug != slug) {
			continue
		}
		matches = append(matches, rackType)
	}
	if len(matches) == 0 {
		return errors.New("no rack type found matching filter")
	}
	if len(matches) > 1 {
		return errors.New("more than one rack type returned, specify a more narrow filter")
	}

	result := matches[0]
	d.SetId(strconv.FormatInt(result.ID, 10))
	d.Set("model", result.Model)
	d.Set("slug", result.Slug)
	if result.Manufacturer != nil {
		d.Set("manufacturer_id", result.Manufacturer.ID)
	}
	if result.FormFactor != nil {
		d.Set("form_factor", result.FormFactor.Value)
	}
	if result.Width != nil {
		d.Set("width", result.Width.Value)
	}
	d.Set("u_height", result.UHeight)
	d.Set("starting_unit", result.StartingUnit)
	d.Set("description", result.Description)
	d.Set("outer_width", result.OuterWidth)
	d.Set("outer_depth", result.OuterDepth)
	if result.OuterUnit != nil {
		d.Set("outer_unit", result.OuterUnit.Value)
	}
	d.Set("weight", result.Weight)
	d.Set("max_weight", result.MaxWeight)
	if result.WeightUnit != nil {
		d.Set("weight_unit", result.WeightUnit.Value)
	}
	d.Set("mounting_depth_mm", result.MountingDepth)
	d.Set("comments", result.Comments)
	return nil
}
