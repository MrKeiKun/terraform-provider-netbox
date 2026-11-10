package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxSiteGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: withContext(resourceNetboxSiteGroupCreate),
		ReadContext:   withContext(resourceNetboxSiteGroupRead),
		UpdateContext: withContext(resourceNetboxSiteGroupUpdate),
		DeleteContext: withContext(resourceNetboxSiteGroupDelete),

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/features/facilities/#site-groups):

> Like regions, site groups can be arranged in a recursive hierarchy for grouping sites. However, whereas regions are intended for geographic organization, site groups may be used for functional grouping. For example, you might classify sites as corporate, branch, or customer sites in addition to where they are physically located.
>
> The use of both regions and site groups affords to independent but complementary dimensions across which sites can be organized.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the site group.",
				Required:    true,
			},
			"slug": {
				Type:         schema.TypeString,
				Description:  "The URL-friendly slug of the site group.",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"parent_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the parent associated with the site group.",
				Optional:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the site group.",
				Optional:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the site group.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the site group.",
				Optional:    true,
			},
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const siteGroupAPIPath = "dcim/site-groups/"

type siteGroupRequest struct {
	Name         string             `json:"name"`
	Slug         string             `json:"slug"`
	Description  string             `json:"description,omitempty"`
	Parent       *int64             `json:"parent,omitempty"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
}

type siteGroupResponse struct {
	Name         string                 `json:"name"`
	Slug         string                 `json:"slug"`
	Description  string                 `json:"description"`
	Parent       *netboxapi.Ref         `json:"parent"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func siteGroupRequestFromResourceData(d *schema.ResourceData) *siteGroupRequest {
	name := d.Get("name").(string)
	data := &siteGroupRequest{
		CustomFields: customFieldsRequest(d),
		Comments:     getOptionalStr(d, "comments", true),
		Owner:        getOptionalInt(d, "owner_id"),
		Name:         name,
		Slug:         slugFromResourceData(d, name),
		Description:  d.Get("description").(string),
		Tags:         []netboxapi.TagRef{},
	}
	if v := int64(d.Get("parent_id").(int)); v != 0 {
		data.Parent = &v
	}
	return data
}

func resourceNetboxSiteGroupCreate(d *schema.ResourceData, m interface{}) error {
	return createObject(d, m, siteGroupAPIPath, siteGroupRequestFromResourceData(d), resourceNetboxSiteGroupRead)
}

func resourceNetboxSiteGroupRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, siteGroupAPIPath, func(r *siteGroupResponse) {
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		d.Set("comments", r.Comments)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("name", r.Name)
		d.Set("slug", r.Slug)
		d.Set("description", r.Description)
		if r.Parent != nil {
			d.Set("parent_id", r.Parent.ID)
		}
	})
}

func resourceNetboxSiteGroupUpdate(d *schema.ResourceData, m interface{}) error {
	return patchObject(d, m, siteGroupAPIPath, siteGroupRequestFromResourceData(d), resourceNetboxSiteGroupRead)
}

func resourceNetboxSiteGroupDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, siteGroupAPIPath)
}
