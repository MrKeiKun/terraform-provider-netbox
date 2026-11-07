package netbox

import (
	"context"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxVirtualDisks() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetboxVirtualDisksCreate,
		ReadContext:   resourceNetboxVirtualDisksRead,
		UpdateContext: resourceNetboxVirtualDisksUpdate,
		DeleteContext: resourceNetboxVirtualDisksDelete,
		Description: `:meta:subcategory:Virtualization:From the [official documentation](https://docs.netbox.dev/en/stable/models/virtualization/virtualdisk/):

		> A virtual disk is used to model discrete virtual hard disks assigned to virtual machines.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"size_mb": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"virtual_machine_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			tagsKey:         tagsSchema,
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceNetboxVirtualDiskResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceNetboxVirtualDiskStateUpgradeV0,
				Version: 0,
			},
		},
	}
}

const virtualDiskAPIPath = "virtualization/virtual-disks/"

type virtualDiskRequest struct {
	Name           string             `json:"name"`
	Size           *int64             `json:"size"`
	VirtualMachine *int64             `json:"virtual_machine"`
	Description    string             `json:"description,omitempty"`
	CustomFields   interface{}        `json:"custom_fields,omitempty"`
	Tags           []netboxapi.TagRef `json:"tags"`
}

type virtualDiskResponse struct {
	Name           string                 `json:"name"`
	Size           *int64                 `json:"size"`
	VirtualMachine *netboxapi.Ref         `json:"virtual_machine"`
	Description    string                 `json:"description"`
	CustomFields   map[string]interface{} `json:"custom_fields"`
	Tags           []netboxapi.TagRef     `json:"tags"`
}

func virtualDiskRequestFromResourceData(api *providerState, d *schema.ResourceData) (*virtualDiskRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &virtualDiskRequest{
		Name:           name,
		Size:           int64ToPtr(int64(d.Get("size_mb").(int))),
		VirtualMachine: int64ToPtr(int64(d.Get("virtual_machine_id").(int))),
		Description:    getOptionalStr(d, "description", true),
		Tags:           tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxVirtualDisksCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	data, err := virtualDiskRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return diag.FromErr(err)
	}
	return createObjectCtx(ctx, d, m, virtualDiskAPIPath, data, resourceNetboxVirtualDisksRead)
}

func resourceNetboxVirtualDisksRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*providerState)
	return readObjectCtx(ctx, d, m, virtualDiskAPIPath, func(r *virtualDiskResponse) {
		d.Set("name", r.Name)
		d.Set("size_mb", r.Size)
		if r.VirtualMachine != nil {
			d.Set("virtual_machine_id", r.VirtualMachine.ID)
		}
		d.Set("description", r.Description)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxVirtualDisksUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	data, err := virtualDiskRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return diag.FromErr(err)
	}
	return putObjectCtx(ctx, d, m, virtualDiskAPIPath, data, resourceNetboxVirtualDisksRead)
}

func resourceNetboxVirtualDisksDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return deleteObjectCtx(ctx, d, m, virtualDiskAPIPath)
}
