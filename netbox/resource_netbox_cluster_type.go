package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxClusterType() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxClusterTypeCreate,
		Read:   resourceNetboxClusterTypeRead,
		Update: resourceNetboxClusterTypeUpdate,
		Delete: resourceNetboxClusterTypeDelete,

		Description: `:meta:subcategory:Virtualization:From the [official documentation](https://docs.netbox.dev/en/stable/features/virtualization/#cluster-types):

> A cluster type represents a technology or mechanism by which a cluster is formed. For example, you might create a cluster type named "VMware vSphere" for a locally hosted cluster or "DigitalOcean NYC3" for one hosted by a cloud provider.`,

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
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const clusterTypeAPIPath = "virtualization/cluster-types/"

type clusterTypeRequest struct {
	Name string             `json:"name"`
	Slug string             `json:"slug"`
	Tags []netboxapi.TagRef `json:"tags"`
}

type clusterTypeResponse struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func clusterTypeRequestFromResourceData(d *schema.ResourceData) *clusterTypeRequest {
	name := d.Get("name").(string)
	return &clusterTypeRequest{
		Name: name,
		Slug: slugFromResourceData(d, name),
		Tags: []netboxapi.TagRef{},
	}
}

func resourceNetboxClusterTypeCreate(d *schema.ResourceData, m interface{}) error {
	return createObject(d, m, clusterTypeAPIPath, clusterTypeRequestFromResourceData(d), resourceNetboxClusterTypeRead)
}

func resourceNetboxClusterTypeRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, clusterTypeAPIPath, func(res *clusterTypeResponse) {
		d.Set("name", res.Name)
		d.Set("slug", res.Slug)
	})
}

func resourceNetboxClusterTypeUpdate(d *schema.ResourceData, m interface{}) error {
	return patchObject(d, m, clusterTypeAPIPath, clusterTypeRequestFromResourceData(d), resourceNetboxClusterTypeRead)
}

func resourceNetboxClusterTypeDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, clusterTypeAPIPath)
}
