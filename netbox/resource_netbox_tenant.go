package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxTenant() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxTenantCreate,
		Read:   resourceNetboxTenantRead,
		Update: resourceNetboxTenantUpdate,
		Delete: resourceNetboxTenantDelete,

		Description: `:meta:subcategory:Tenancy:From the [official documentation](https://docs.netbox.dev/en/stable/features/tenancy/#tenants):

> A tenant represents a discrete grouping of resources used for administrative purposes. Typically, tenants are used to represent individual customers or internal departments within an organization.
>
> Tenant assignment is used to signify the ownership of an object in NetBox. As such, each object may only be owned by a single tenant. For example, if you have a firewall dedicated to a particular customer, you would assign it to the tenant which represents that customer. However, if the firewall serves multiple customers, it doesn't belong to any particular customer, so tenant assignment would not be appropriate.`,

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
			tagsKey: tagsSchema,
			"group_id": {
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

const tenantAPIPath = "tenancy/tenants/"

type tenantRequest struct {
	Name        string             `json:"name"`
	Slug        string             `json:"slug"`
	Description string             `json:"description,omitempty"`
	Group       *int64             `json:"group,omitempty"`
	Tags        []netboxapi.TagRef `json:"tags"`
}

type tenantResponse struct {
	Name        string         `json:"name"`
	Slug        string         `json:"slug"`
	Description string         `json:"description"`
	Group       *netboxapi.Ref `json:"group"`
}

func tenantRequestFromResourceData(api *providerState, d *schema.ResourceData) (*tenantRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &tenantRequest{
		Name:        name,
		Slug:        slugFromResourceData(d, name),
		Description: d.Get("description").(string),
		Tags:        tags,
	}
	if v := int64(d.Get("group_id").(int)); v != 0 {
		data.Group = &v
	}
	return data, nil
}

func resourceNetboxTenantCreate(d *schema.ResourceData, m interface{}) error {
	data, err := tenantRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, tenantAPIPath, data, resourceNetboxTenantRead)
}

func resourceNetboxTenantRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, tenantAPIPath, func(r *tenantResponse) {
		d.Set("name", r.Name)
		d.Set("slug", r.Slug)
		d.Set("description", r.Description)
		if r.Group != nil {
			d.Set("group_id", r.Group.ID)
		}
	})
}

func resourceNetboxTenantUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := tenantRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, tenantAPIPath, data, resourceNetboxTenantRead)
}

func resourceNetboxTenantDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, tenantAPIPath)
}
