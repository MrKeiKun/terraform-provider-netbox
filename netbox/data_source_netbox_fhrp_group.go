package netbox

import (
	"net/url"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxFhrpGroup() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxFhrpGroupRead,
		Description: `:meta:subcategory:IP Address Management (IPAM):`,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"protocol": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"group_id": {
				Type:     schema.TypeInt,
				Optional: false,
				Required: true,
			},
			"auth_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auth_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"comments": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchemaRead,
		},
	}
}

type fhrpGroupDataSourceResponse struct {
	ID          int64              `json:"id"`
	Protocol    string             `json:"protocol"`
	GroupId     int64              `json:"group_id"`
	AuthType    string             `json:"auth_type"`
	AuthKey     string             `json:"auth_key"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Comments    string             `json:"comments"`
	Tags        []netboxapi.TagRef `json:"tags"`
}

func dataSourceNetboxFhrpGroupRead(d *schema.ResourceData, m interface{}) error {
	query := url.Values{}
	addIntQuery(query, d, "group_id", "group_id")
	addStrQuery(query, d, "protocol", "protocol")
	r, err := findOne[fhrpGroupDataSourceResponse](m, "ipam/fhrp-groups/", query, "group")
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(r.ID, 10))
	d.Set("id", r.ID)
	d.Set("protocol", r.Protocol)
	d.Set("group_id", r.GroupId)
	d.Set("auth_type", r.AuthType)
	d.Set("auth_key", r.AuthKey)
	d.Set("name", r.Name)
	d.Set("description", r.Description)
	d.Set("comments", r.Comments)
	d.Set("tags", tagNames(r.Tags))
	return nil
}
