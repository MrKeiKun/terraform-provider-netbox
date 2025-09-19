package netbox

import (
	"errors"
	"fmt"

	"github.com/fbreckle/go-netbox/netbox/client/ipam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceNetboxAggregates() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxAggregatesRead,
		Description: `:meta:subcategory:IP Address Management (IPAM):`,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"limit": {
				Type:             schema.TypeInt,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(1)),
				Default:          0,
			},
			"aggregates": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"prefix": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"rir_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"tenant_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceNetboxAggregatesRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)

	params := ipam.NewIpamAggregatesListParams()

	if limitValue, ok := d.GetOk("limit"); ok {
		params.Limit = int64ToPtr(int64(limitValue.(int)))
	}

	if filter, ok := d.GetOk("filter"); ok {
		var filterParams = filter.(*schema.Set)
		var tags []string
		for _, f := range filterParams.List() {
			k := f.(map[string]interface{})["name"]
			v := f.(map[string]interface{})["value"]
			vString := v.(string)
			switch k {
			case "id":
				params.ID = &vString
			case "prefix":
				params.Prefix = &vString
			case "description":
				params.Description = &vString
			case "rir":
				params.Rir = &vString
			case "rir__n":
				params.Rirn = &vString
			case "rir_id":
				params.RirID = &vString
			case "rir_id__n":
				params.RirIDn = &vString
			case "tenant":
				params.Tenant = &vString
			case "tenant__n":
				params.Tenantn = &vString
			case "tenant_id":
				params.TenantID = &vString
			case "tenant_id__n":
				params.TenantIDn = &vString
			case "tenant_group":
				params.TenantGroup = &vString
			case "tenant_group__n":
				params.TenantGroupn = &vString
			case "tenant_group_id":
				params.TenantGroupID = &vString
			case "tenant_group_id__n":
				params.TenantGroupIDn = &vString
			case "tag":
				tags = append(tags, vString)
				params.Tag = tags
			default:
				return fmt.Errorf("'%s' is not a supported filter parameter", k)
			}
		}
	}

	res, err := api.Ipam.IpamAggregatesList(params, nil)
	if err != nil {
		return err
	}

	if *res.GetPayload().Count == int64(0) {
		return errors.New("no result")
	}

	filteredAggregates := res.GetPayload().Results

	var s []map[string]interface{}
	for _, v := range filteredAggregates {
		var mapping = make(map[string]interface{})

		mapping["id"] = v.ID
		mapping["prefix"] = v.Prefix
		mapping["description"] = v.Description
		if v.Rir != nil {
			mapping["rir_id"] = v.Rir.ID
		}
		if v.Tenant != nil {
			mapping["tenant_id"] = v.Tenant.ID
		}

		s = append(s, mapping)
	}

	d.SetId(id.UniqueId())
	return d.Set("aggregates", s)
}
