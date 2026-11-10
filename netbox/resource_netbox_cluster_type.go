package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxClusterType() *schema.Resource {
	return &schema.Resource{
		CreateContext: withContext(resourceNetboxClusterTypeCreate),
		ReadContext:   withContext(resourceNetboxClusterTypeRead),
		UpdateContext: withContext(resourceNetboxClusterTypeUpdate),
		DeleteContext: withContext(resourceNetboxClusterTypeDelete),

		Description: `:meta:subcategory:Virtualization:From the [official documentation](https://docs.netbox.dev/en/stable/features/virtualization/#cluster-types):

> A cluster type represents a technology or mechanism by which a cluster is formed. For example, you might create a cluster type named "VMware vSphere" for a locally hosted cluster or "DigitalOcean NYC3" for one hosted by a cloud provider.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the cluster type.",
				Required:    true,
			},
			"slug": {
				Type:         schema.TypeString,
				Description:  "The URL-friendly slug of the cluster type.",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the cluster type.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the cluster type.",
				Optional:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the cluster type.",
				Optional:    true,
			},
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const clusterTypeAPIPath = "virtualization/cluster-types/"

type clusterTypeRequest struct {
	Name         string             `json:"name"`
	Slug         string             `json:"slug"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	Description  string             `json:"description,omitempty"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
}

type clusterTypeResponse struct {
	Name         string                 `json:"name"`
	Slug         string                 `json:"slug"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	Description  string                 `json:"description"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func clusterTypeRequestFromResourceData(d *schema.ResourceData) *clusterTypeRequest {
	name := d.Get("name").(string)
	return &clusterTypeRequest{
		CustomFields: customFieldsRequest(d),
		Description:  getOptionalStr(d, "description", true),
		Comments:     getOptionalStr(d, "comments", true),
		Owner:        getOptionalInt(d, "owner_id"),
		Name:         name,
		Slug:         slugFromResourceData(d, name),
		Tags:         []netboxapi.TagRef{},
	}
}

func resourceNetboxClusterTypeCreate(d *schema.ResourceData, m interface{}) error {
	return createObject(d, m, clusterTypeAPIPath, clusterTypeRequestFromResourceData(d), resourceNetboxClusterTypeRead)
}

func resourceNetboxClusterTypeRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, clusterTypeAPIPath, func(res *clusterTypeResponse) {
		d.Set(customFieldsKey, getCustomFields(res.CustomFields))
		d.Set("description", res.Description)
		d.Set("comments", res.Comments)
		d.Set("owner_id", netboxapi.RefID(res.Owner))
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
