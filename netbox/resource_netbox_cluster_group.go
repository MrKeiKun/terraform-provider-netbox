package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxClusterGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxClusterGroupCreate,
		Read:   resourceNetboxClusterGroupRead,
		Update: resourceNetboxClusterGroupUpdate,
		Delete: resourceNetboxClusterGroupDelete,

		Description: `:meta:subcategory:Virtualization:From the [official documentation](https://docs.netbox.dev/en/stable/features/virtualization/#cluster-groups):

> Cluster groups may be created for the purpose of organizing clusters. The arrangement of clusters into groups is optional.`,

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
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const clusterGroupAPIPath = "virtualization/cluster-groups/"

type clusterGroupRequest struct {
	Name        string             `json:"name"`
	Slug        string             `json:"slug"`
	Description string             `json:"description,omitempty"`
	Tags        []netboxapi.TagRef `json:"tags"`
}

type clusterGroupResponse struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

func clusterGroupRequestFromResourceData(d *schema.ResourceData) *clusterGroupRequest {
	name := d.Get("name").(string)
	return &clusterGroupRequest{
		Name: name,
		Slug: slugFromResourceData(d, name),
		// description omits empty values, so a cleared one is sent as a single space
		Description: getOptionalStr(d, "description", true),
		Tags:        []netboxapi.TagRef{},
	}
}

func resourceNetboxClusterGroupCreate(d *schema.ResourceData, m interface{}) error {
	return createObject(d, m, clusterGroupAPIPath, clusterGroupRequestFromResourceData(d), resourceNetboxClusterGroupRead)
}

func resourceNetboxClusterGroupRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, clusterGroupAPIPath, func(res *clusterGroupResponse) {
		d.Set("name", res.Name)
		d.Set("slug", res.Slug)
		d.Set("description", res.Description)
	})
}

func resourceNetboxClusterGroupUpdate(d *schema.ResourceData, m interface{}) error {
	return patchObject(d, m, clusterGroupAPIPath, clusterGroupRequestFromResourceData(d), resourceNetboxClusterGroupRead)
}

func resourceNetboxClusterGroupDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, clusterGroupAPIPath)
}
