package netbox

import (
	"encoding/json"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxConfigContext() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxConfigContextCreate,
		Read:   resourceNetboxConfigContextRead,
		Update: resourceNetboxConfigContextUpdate,
		Delete: resourceNetboxConfigContextDelete,

		Description: `:meta:subcategory:Extras:From the [official documentation](https://docs.netbox.dev/en/stable/models/extras/configcontext/):

> Context data is made available to devices and/or virtual machines based on their relationships to other objects in NetBox. For example, context data can be associated only with devices assigned to a particular site, or only to virtual machines in a certain cluster.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"weight": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1000,
			},
			"data": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsJSON,
			},
			"cluster_groups": {
				Type:     schema.TypeSet,
				Optional: true,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"cluster_types": {
				Type:     schema.TypeSet,
				Optional: true,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"clusters": {
				Type:     schema.TypeSet,
				Optional: true,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"device_types": {
				Type:     schema.TypeSet,
				Optional: true,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"locations": {
				Type:     schema.TypeSet,
				Optional: true,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"platforms": {
				Type:     schema.TypeSet,
				Optional: true,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"regions": {
				Type:     schema.TypeSet,
				Optional: true,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"roles": {
				Type:     schema.TypeSet,
				Optional: true,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"site_groups": {
				Type:     schema.TypeSet,
				Optional: true,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"sites": {
				Type:     schema.TypeSet,
				Optional: true,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"tenant_groups": {
				Type:     schema.TypeSet,
				Optional: true,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"tenants": {
				Type:     schema.TypeSet,
				Optional: true,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const configContextAPIPath = "extras/config-contexts/"

type configContextRequest struct {
	Name          string      `json:"name"`
	Data          interface{} `json:"data"`
	Description   string      `json:"description,omitempty"`
	ClusterGroups []int64     `json:"cluster_groups"`
	ClusterTypes  []int64     `json:"cluster_types"`
	Clusters      []int64     `json:"clusters"`
	DeviceTypes   []int64     `json:"device_types"`
	Locations     []int64     `json:"locations"`
	Platforms     []int64     `json:"platforms"`
	Regions       []int64     `json:"regions"`
	Roles         []int64     `json:"roles"`
	SiteGroups    []int64     `json:"site_groups"`
	Sites         []int64     `json:"sites"`
	TenantGroups  []int64     `json:"tenant_groups"`
	Tenants       []int64     `json:"tenants"`
	Tags          []string    `json:"tags"`
	Weight        *int64      `json:"weight,omitempty"`
}

type configContextResponse struct {
	Name          string          `json:"name"`
	Description   string          `json:"description"`
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

func configContextRequestFromResourceData(d *schema.ResourceData) *configContextRequest {
	data := &configContextRequest{
		Name:          d.Get("name").(string),
		Description:   d.Get("description").(string),
		ClusterGroups: toInt64List(d.Get("cluster_groups")),
		ClusterTypes:  toInt64List(d.Get("cluster_types")),
		Clusters:      toInt64List(d.Get("clusters")),
		DeviceTypes:   toInt64List(d.Get("device_types")),
		Locations:     toInt64List(d.Get("locations")),
		Platforms:     toInt64List(d.Get("platforms")),
		Regions:       toInt64List(d.Get("regions")),
		Roles:         toInt64List(d.Get("roles")),
		SiteGroups:    toInt64List(d.Get("site_groups")),
		Sites:         toInt64List(d.Get("sites")),
		TenantGroups:  toInt64List(d.Get("tenant_groups")),
		Tenants:       toInt64List(d.Get("tenants")),
		Tags:          toStringList(d.Get(tagsAllKey)),
		Weight:        int64ToPtr(int64(d.Get("weight").(int))),
	}

	if dataJSON, ok := d.GetOk("data"); ok {
		var jsonObj any
		if err := json.Unmarshal([]byte(dataJSON.(string)), &jsonObj); err == nil {
			data.Data = jsonObj
		}
	}
	return data
}

func resourceNetboxConfigContextCreate(d *schema.ResourceData, m interface{}) error {
	return createObject(d, m, configContextAPIPath, configContextRequestFromResourceData(d), resourceNetboxConfigContextRead)
}

func resourceNetboxConfigContextRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, configContextAPIPath, func(cc *configContextResponse) {
		d.Set("name", cc.Name)
		d.Set("description", cc.Description)
		d.Set("weight", cc.Weight)
		if cc.Data != nil {
			if jsonArr, err := json.Marshal(cc.Data); err == nil {
				d.Set("data", string(jsonArr))
			}
		} else {
			d.Set("data", nil)
		}

		d.Set("cluster_groups", netboxapi.IDs(cc.ClusterGroups))
		d.Set("cluster_types", netboxapi.IDs(cc.ClusterTypes))
		d.Set("clusters", netboxapi.IDs(cc.Clusters))
		d.Set("device_types", netboxapi.IDs(cc.DeviceTypes))
		d.Set("locations", netboxapi.IDs(cc.Locations))
		d.Set("platforms", netboxapi.IDs(cc.Platforms))
		d.Set("regions", netboxapi.IDs(cc.Regions))
		d.Set("roles", netboxapi.IDs(cc.Roles))
		d.Set("site_groups", netboxapi.IDs(cc.SiteGroups))
		d.Set("sites", netboxapi.IDs(cc.Sites))
		d.Set("tenant_groups", netboxapi.IDs(cc.TenantGroups))
		d.Set("tenants", netboxapi.IDs(cc.Tenants))

		// config contexts return the tags as plain names
		api.readTagNames(d, cc.Tags)
	})
}

func resourceNetboxConfigContextUpdate(d *schema.ResourceData, m interface{}) error {
	return patchObject(d, m, configContextAPIPath, configContextRequestFromResourceData(d), resourceNetboxConfigContextRead)
}

func resourceNetboxConfigContextDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, configContextAPIPath)
}
