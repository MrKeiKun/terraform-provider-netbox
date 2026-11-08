package netbox

import (
	"net/url"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxCluster() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxClusterRead,
		Description: `:meta:subcategory:Virtualization:Retrieves a single cluster from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the cluster associated with the cluster.",
				Computed:    true,
			},
			"id": {
				Type:         schema.TypeString,
				Description:  "The numeric ID of the cluster.",
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name", "site_id", "id"},
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the cluster.",
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the cluster.",
				Computed:    true,
			},
			"site_id": {
				Type:         schema.TypeInt,
				Description:  "The ID of the site associated with the cluster.",
				Computed:     true,
				Optional:     true,
				AtLeastOneOf: []string{"name", "site_id", "id"},
			},
			"site_group_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the site group associated with the cluster.",
				Computed:    true,
			},
			"location_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the location associated with the cluster.",
				Computed:    true,
			},
			"region_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the region associated with the cluster.",
				Computed:    true,
			},
			"name": {
				Type:         schema.TypeString,
				Description:  "The name of the cluster.",
				Optional:     true,
				AtLeastOneOf: []string{"name", "site_id", "id"},
			},
			"cluster_type_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the cluster type associated with the cluster.",
				Computed:    true,
			},
			"cluster_group_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the cluster group associated with the cluster.",
				Computed:    true,
				Optional:    true,
			},
			"scope_type": {
				Type:        schema.TypeString,
				Description: "The scope type of the cluster.",
				Computed:    true,
			},
			"scope_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the scope associated with the cluster.",
				Computed:    true,
			},
			"custom_fields": {
				Type:        schema.TypeMap,
				Description: "Custom field values of the cluster, as a map of field name to value.",
				Computed:    true,
			},
			tagsKey: tagsSchemaRead,
		},
	}
}

type clusterDataSourceResponse struct {
	ID           int64                  `json:"id"`
	Name         string                 `json:"name"`
	Type         *netboxapi.Ref         `json:"type"`
	Group        *netboxapi.Ref         `json:"group"`
	Comments     string                 `json:"comments"`
	Description  string                 `json:"description"`
	ScopeType    *string                `json:"scope_type"`
	ScopeID      *int64                 `json:"scope_id"`
	CustomFields map[string]interface{} `json:"custom_fields"`
	Tags         []netboxapi.TagRef     `json:"tags"`
}

func dataSourceNetboxClusterRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addStrQuery(query, d, "name", "name")
	addIntQuery(query, d, "site_id", "site_id")
	addStrQuery(query, d, "id", "id")
	addIntQuery(query, d, "cluster_group_id", "group_id")

	result, err := findOneWith[clusterDataSourceResponse](m, "virtualization/clusters/", query,
		"more than one result, specify a more narrow filter", "no result")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(result.ID, 10))
	d.Set("cluster_id", result.ID)
	d.Set("name", result.Name)
	if result.Type != nil {
		d.Set("cluster_type_id", result.Type.ID)
	}
	d.Set("cluster_group_id", netboxapi.RefID(result.Group))
	d.Set("comments", result.Comments)
	d.Set("description", result.Description)
	if result.ScopeType != nil && result.ScopeID != nil {
		d.Set("scope_type", *result.ScopeType)
		d.Set("scope_id", *result.ScopeID)
		switch *result.ScopeType {
		case "dcim.site":
			d.Set("site_id", *result.ScopeID)
		case "dcim.sitegroup":
			d.Set("site_group_id", *result.ScopeID)
		case "dcim.location":
			d.Set("location_id", *result.ScopeID)
		case "dcim.region":
			d.Set("region_id", *result.ScopeID)
		}
	}
	if result.CustomFields != nil {
		d.Set("custom_fields", flattenCustomFields(result.CustomFields))
	}
	d.Set(tagsKey, tagNames(result.Tags))
	return nil
}
