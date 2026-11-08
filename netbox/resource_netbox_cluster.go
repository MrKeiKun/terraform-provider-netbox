package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxClusterCreate,
		Read:   resourceNetboxClusterRead,
		Update: resourceNetboxClusterUpdate,
		Delete: resourceNetboxClusterDelete,

		Description: `:meta:subcategory:Virtualization:From the [official documentation](https://netboxlabs.com/docs/netbox/models/virtualization/cluster/):
> A cluster is a logical grouping of physical resources within which virtual machines run. Physical devices may be associated with clusters as hosts. This allows users to track on which host(s) a particular virtual machine may reside.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"cluster_type_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"cluster_group_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"comments": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"location_id": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"site_id", "site_group_id", "region_id"},
			},
			"site_id": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"location_id", "site_group_id", "region_id"},
			},
			"site_group_id": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"location_id", "site_id", "region_id"},
			},
			"region_id": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"location_id", "site_id", "site_group_id"},
			},
			"tenant_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"owner_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "One of [planned, staging, active, decommissioning, offline]",
				ValidateFunc: validation.StringInSlice([]string{"planned", "staging", "active", "decommissioning", "offline"}, false),
			},
			customFieldsKey: customFieldsSchema,
			tagsKey:         tagsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const clusterAPIPath = "virtualization/clusters/"

type clusterRequest struct {
	Name         string             `json:"name"`
	Type         *int64             `json:"type"`
	Group        *int64             `json:"group,omitempty"`
	Comments     string             `json:"comments,omitempty"`
	Description  string             `json:"description,omitempty"`
	Tenant       *int64             `json:"tenant,omitempty"`
	ScopeType    *string            `json:"scope_type"`
	ScopeID      *int64             `json:"scope_id"`
	Owner        *int64             `json:"owner"`
	Status       string             `json:"status,omitempty"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
}

type clusterResponse struct {
	Name         string                 `json:"name"`
	Type         *netboxapi.Ref         `json:"type"`
	Group        *netboxapi.Ref         `json:"group"`
	Comments     string                 `json:"comments"`
	Description  string                 `json:"description"`
	Tenant       *netboxapi.Ref         `json:"tenant"`
	ScopeType    *string                `json:"scope_type"`
	ScopeID      *int64                 `json:"scope_id"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Status       *netboxapi.Choice      `json:"status"`
	CustomFields map[string]interface{} `json:"custom_fields"`
	Tags         []netboxapi.TagRef     `json:"tags"`
}

func clusterRequestFromResourceData(api *providerState, d *schema.ResourceData) (*clusterRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &clusterRequest{
		CustomFields: customFieldsRequest(d),
		Status:       d.Get("status").(string),
		Owner:        getOptionalInt(d, "owner_id"),
		Name:         name,
		Type:         int64ToPtr(int64(d.Get("cluster_type_id").(int))),
		Comments:     getOptionalStr(d, "comments", true),
		Description:  getOptionalStr(d, "description", true),

		Tags: tags,
	}
	if v, ok := d.GetOk("cluster_group_id"); ok {
		data.Group = int64ToPtr(int64(v.(int)))
	}
	if v, ok := d.GetOk("tenant_id"); ok {
		data.Tenant = int64ToPtr(int64(v.(int)))
	}

	siteID := getOptionalInt(d, "site_id")
	siteGroupID := getOptionalInt(d, "site_group_id")
	locationID := getOptionalInt(d, "location_id")
	regionID := getOptionalInt(d, "region_id")

	switch {
	case siteID != nil:
		data.ScopeType = strToPtr("dcim.site")
		data.ScopeID = siteID
	case siteGroupID != nil:
		data.ScopeType = strToPtr("dcim.sitegroup")
		data.ScopeID = siteGroupID
	case locationID != nil:
		data.ScopeType = strToPtr("dcim.location")
		data.ScopeID = locationID
	case regionID != nil:
		data.ScopeType = strToPtr("dcim.region")
		data.ScopeID = regionID
	default:
		data.ScopeType = nil
		data.ScopeID = nil
	}
	return data, nil
}

func resourceNetboxClusterCreate(d *schema.ResourceData, m interface{}) error {
	data, err := clusterRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, clusterAPIPath, data, resourceNetboxClusterRead)
}

func resourceNetboxClusterRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, clusterAPIPath, func(r *clusterResponse) {
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		d.Set("status", netboxapi.ChoiceValue(r.Status))
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("name", r.Name)
		d.Set("cluster_type_id", netboxapi.RefID(r.Type))
		d.Set("cluster_group_id", netboxapi.RefID(r.Group))
		d.Set("comments", r.Comments)
		d.Set("description", r.Description)
		d.Set("tenant_id", netboxapi.RefID(r.Tenant))

		d.Set("site_id", nil)
		d.Set("site_group_id", nil)
		d.Set("location_id", nil)
		d.Set("region_id", nil)
		if r.ScopeType != nil && r.ScopeID != nil {
			switch *r.ScopeType {
			case "dcim.site":
				d.Set("site_id", *r.ScopeID)
			case "dcim.sitegroup":
				d.Set("site_group_id", *r.ScopeID)
			case "dcim.location":
				d.Set("location_id", *r.ScopeID)
			case "dcim.region":
				d.Set("region_id", *r.ScopeID)
			}
		}
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxClusterUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := clusterRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, clusterAPIPath, data, resourceNetboxClusterRead)
}

func resourceNetboxClusterDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, clusterAPIPath)
}
