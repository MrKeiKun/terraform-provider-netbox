package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxVirtualMachineType() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxVirtualMachineTypeCreate,
		Read:   resourceNetboxVirtualMachineTypeRead,
		Update: resourceNetboxVirtualMachineTypeUpdate,
		Delete: resourceNetboxVirtualMachineTypeDelete,

		Description: `:meta:subcategory:Virtualization:From the [official documentation](https://docs.netbox.dev/en/stable/models/virtualization/virtualmachinetype/):

> A virtual machine type defines default resources, such as the number of vCPUs and the memory, for virtual machines.

Requires NetBox v4.6.0 or newer.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the virtual machine type.",
				Required:    true,
			},
			"slug": {
				Type:         schema.TypeString,
				Description:  "The URL-friendly slug of the virtual machine type.",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"default_platform_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the default platform associated with the virtual machine type.",
				Optional:    true,
			},
			"default_vcpus": {
				Type:        schema.TypeFloat,
				Description: "The default vcpus of the virtual machine type.",
				Optional:    true,
			},
			"default_memory": {
				Type:         schema.TypeInt,
				Description:  "The default memory of the virtual machine type.",
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 2147483647),
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the virtual machine type.",
				Optional:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the virtual machine type.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the virtual machine type.",
				Optional:    true,
			},
			tagsKey:         tagsSchema,
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const virtualMachineTypeAPIPath = "virtualization/virtual-machine-types/"

type virtualMachineTypeRequest struct {
	Name            string             `json:"name"`
	Slug            string             `json:"slug"`
	DefaultPlatform *int64             `json:"default_platform"`
	DefaultVcpus    *float64           `json:"default_vcpus"`
	DefaultMemory   *int64             `json:"default_memory"`
	Description     string             `json:"description,omitempty"`
	Owner           *int64             `json:"owner"`
	Comments        string             `json:"comments,omitempty"`
	Tags            []netboxapi.TagRef `json:"tags"`
	CustomFields    interface{}        `json:"custom_fields,omitempty"`
}

type virtualMachineTypeResponse struct {
	Name            string                 `json:"name"`
	Slug            string                 `json:"slug"`
	DefaultPlatform *netboxapi.Ref         `json:"default_platform"`
	DefaultVcpus    *float64               `json:"default_vcpus"`
	DefaultMemory   *int64                 `json:"default_memory"`
	Description     string                 `json:"description"`
	Owner           *netboxapi.Ref         `json:"owner"`
	Comments        string                 `json:"comments"`
	Tags            []netboxapi.TagRef     `json:"tags"`
	CustomFields    map[string]interface{} `json:"custom_fields"`
}

func virtualMachineTypeRequestFromResourceData(api *providerState, d *schema.ResourceData) (*virtualMachineTypeRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &virtualMachineTypeRequest{
		Name:            d.Get("name").(string),
		Slug:            slugFromResourceData(d, name),
		DefaultPlatform: getOptionalInt(d, "default_platform_id"),
		DefaultVcpus:    getOptionalFloat(d, "default_vcpus"),
		DefaultMemory:   getOptionalInt(d, "default_memory"),
		Description:     getOptionalStr(d, "description", true),
		Owner:           getOptionalInt(d, "owner_id"),
		Comments:        getOptionalStr(d, "comments", true),
		Tags:            tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxVirtualMachineTypeCreate(d *schema.ResourceData, m interface{}) error {
	data, err := virtualMachineTypeRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, virtualMachineTypeAPIPath, data, resourceNetboxVirtualMachineTypeRead)
}

func resourceNetboxVirtualMachineTypeRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, virtualMachineTypeAPIPath, func(r *virtualMachineTypeResponse) {
		d.Set("name", r.Name)
		d.Set("slug", r.Slug)
		d.Set("default_platform_id", netboxapi.RefID(r.DefaultPlatform))
		d.Set("default_vcpus", r.DefaultVcpus)
		d.Set("default_memory", r.DefaultMemory)
		d.Set("description", r.Description)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("comments", r.Comments)
		api.readTagRefs(d, r.Tags)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
	})
}

func resourceNetboxVirtualMachineTypeUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := virtualMachineTypeRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, virtualMachineTypeAPIPath, data, resourceNetboxVirtualMachineTypeRead)
}

func resourceNetboxVirtualMachineTypeDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, virtualMachineTypeAPIPath)
}
