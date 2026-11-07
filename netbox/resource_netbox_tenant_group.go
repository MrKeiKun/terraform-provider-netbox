package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxTenantGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxTenantGroupCreate,
		Read:   resourceNetboxTenantGroupRead,
		Update: resourceNetboxTenantGroupUpdate,
		Delete: resourceNetboxTenantGroupDelete,

		Description: `:meta:subcategory:Tenancy:From the [official documentation](https://docs.netbox.dev/en/stable/features/tenancy/#tenant-groups):

> Tenants can be organized by custom groups. For instance, you might create one group called "Customers" and one called "Departments." The assignment of a tenant to a group is optional.
>
> Tenant groups may be nested recursively to achieve a multi-level hierarchy. For example, you might have a group called "Customers" containing subgroups of individual tenants grouped by product or account team.`,

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
			"parent_id": {
				Type:     schema.TypeInt,
				Optional: true,
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

const tenantGroupAPIPath = "tenancy/tenant-groups/"

type tenantGroupRequest struct {
	Name        string             `json:"name"`
	Slug        string             `json:"slug"`
	Description string             `json:"description,omitempty"`
	Parent      *int64             `json:"parent,omitempty"`
	Tags        []netboxapi.TagRef `json:"tags"`
}

type tenantGroupResponse struct {
	Name        string         `json:"name"`
	Slug        string         `json:"slug"`
	Description string         `json:"description"`
	Parent      *netboxapi.Ref `json:"parent"`
}

func tenantGroupRequestFromResourceData(d *schema.ResourceData) *tenantGroupRequest {
	name := d.Get("name").(string)
	data := &tenantGroupRequest{
		Name:        name,
		Slug:        slugFromResourceData(d, name),
		Description: d.Get("description").(string),
		Tags:        []netboxapi.TagRef{},
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
