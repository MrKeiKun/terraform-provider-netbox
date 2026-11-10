package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxTenantGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: withContext(resourceNetboxTenantGroupCreate),
		ReadContext:   withContext(resourceNetboxTenantGroupRead),
		UpdateContext: withContext(resourceNetboxTenantGroupUpdate),
		DeleteContext: withContext(resourceNetboxTenantGroupDelete),

		Description: `:meta:subcategory:Tenancy:From the [official documentation](https://docs.netbox.dev/en/stable/features/tenancy/#tenant-groups):

> Tenants can be organized by custom groups. For instance, you might create one group called "Customers" and one called "Departments." The assignment of a tenant to a group is optional.
>
> Tenant groups may be nested recursively to achieve a multi-level hierarchy. For example, you might have a group called "Customers" containing subgroups of individual tenants grouped by product or account team.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the tenant group.",
				Required:    true,
			},
			"slug": {
				Type:         schema.TypeString,
				Description:  "The URL-friendly slug of the tenant group.",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"parent_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the parent associated with the tenant group.",
				Optional:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the tenant group.",
				Optional:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the tenant group.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the tenant group.",
				Optional:    true,
			},
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const tenantGroupAPIPath = "tenancy/tenant-groups/"

type tenantGroupRequest struct {
	Name         string             `json:"name"`
	Slug         string             `json:"slug"`
	Description  string             `json:"description,omitempty"`
	Parent       *int64             `json:"parent,omitempty"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
}

type tenantGroupResponse struct {
	Name         string                 `json:"name"`
	Slug         string                 `json:"slug"`
	Description  string                 `json:"description"`
	Parent       *netboxapi.Ref         `json:"parent"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func tenantGroupRequestFromResourceData(d *schema.ResourceData) *tenantGroupRequest {
	name := d.Get("name").(string)
	data := &tenantGroupRequest{
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

func resourceNetboxTenantGroupCreate(d *schema.ResourceData, m interface{}) error {
	return createObject(d, m, tenantGroupAPIPath, tenantGroupRequestFromResourceData(d), resourceNetboxTenantGroupRead)
}

func resourceNetboxTenantGroupRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, tenantGroupAPIPath, func(r *tenantGroupResponse) {
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

func resourceNetboxTenantGroupUpdate(d *schema.ResourceData, m interface{}) error {
	return patchObject(d, m, tenantGroupAPIPath, tenantGroupRequestFromResourceData(d), resourceNetboxTenantGroupRead)
}

func resourceNetboxTenantGroupDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, tenantGroupAPIPath)
}
