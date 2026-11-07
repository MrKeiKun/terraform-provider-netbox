package netbox

import (
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxConfigContext() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxConfigContextRead,
		Description: `:meta:subcategory:Extras:`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"weight": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"data": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_groups": {
				Type:     schema.TypeList,
				Computed: true,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"cluster_types": {
				Type:     schema.TypeList,
				Computed: true,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"clusters": {
				Type:     schema.TypeList,
				Computed: true,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"device_types": {
				Type:     schema.TypeList,
				Computed: true,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"locations": {
				Type:     schema.TypeList,
				Computed: true,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"platforms": {
				Type:     schema.TypeList,
				Computed: true,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"regions": {
				Type:     schema.TypeList,
				Computed: true,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"roles": {
				Type:     schema.TypeList,
				Computed: true,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"site_groups": {
				Type:     schema.TypeList,
				Computed: true,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"sites": {
				Type:     schema.TypeList,
				Computed: true,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"tenant_groups": {
				Type:     schema.TypeList,
				Computed: true,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"tenants": {
				Type:     schema.TypeList,
				Computed: true,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"tags": {
				Type:     schema.TypeList,
				Computed: true,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

type configContextDataSourceResponse struct {
	ID            int64           `json:"id"`
	Name          string          `json:"name"`
	Weight        *int64          `json:"weight"`
	Data          interface{}     `json:"data"`
	ClusterGroups []netboxapi.Ref `json:"cluster_groups"`
	ClusterTypes  []netboxapi.Ref `json:"cluster_types"`
	Clusters      []netboxapi.Ref `json:"clusters"`
	DeviceTypes   []netboxapi.Ref `json:"device_types"`
	Locations     []netboxapi.Ref `json:"locations"`
	Platforms     []netboxapi.Ref `json:"platforms"`
	Regions       []netboxapi.Ref `json:"regions"`
	Roles         []netboxapi.Ref `json:"roles"`
	SiteGroups    []netboxapi.Ref `json:"site_groups"`
	Sites         []netboxapi.Ref `json:"sites"`
	TenantGroups  []netboxapi.Ref `json:"tenant_groups"`
	Tenants       []netboxapi.Ref `json:"tenants"`
	Tags          []string        `json:"tags"`
}

func dataSourceNetboxConfigContextRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	query.Set("name", d.Get("name").(string))

	result, err := findOneWith[configContextDataSourceResponse](m, "extras/config-contexts/", query,
		"more than one result. Specify a more narrow filter", "no result")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(result.ID, 10))
	d.Set("name", result.Name)
	d.Set("weight", result.Weight)
	if result.Data != nil {
		if jsonArr, err := json.Marshal(result.Data); err == nil {
			d.Set("data", string(jsonArr))
		}
	} else {
		d.Set("data", nil)
	}

	// the lists are never nil, an empty list is stored as an empty list
	ids := func(refs []netboxapi.Ref) []int64 { return netboxapi.IDs(refs) }
	d.Set("cluster_groups", ids(result.ClusterGroups))
	d.Set("cluster_types", ids(result.ClusterTypes))
	d.Set("clusters", ids(result.Clusters))
	d.Set("device_types", ids(result.DeviceTypes))
	d.Set("locations", ids(result.Locations))
	d.Set("platforms", ids(result.Platforms))
	d.Set("regions", ids(result.Regions))
	d.Set("roles", ids(result.Roles))
	d.Set("site_groups", ids(result.SiteGroups))
	d.Set("sites", ids(result.Sites))
	d.Set("tenant_groups", ids(result.TenantGroups))
	d.Set("tenants", ids(result.Tenants))
	d.Set("tags", result.Tags)
	return nil
}
