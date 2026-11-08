package netbox

import (
	"context"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxVirtualChassis() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetboxVirtualChassisCreate,
		ReadContext:   resourceNetboxVirtualChassisRead,
		UpdateContext: resourceNetboxVirtualChassisUpdate,
		DeleteContext: resourceNetboxVirtualChassisDelete,
		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/features/devices-cabling/#virtual-chassis):

		> Sometimes it is necessary to model a set of physical devices as sharing a single management plane. Perhaps the most common example of such a scenario is stackable switches. These can be modeled as virtual chassis in NetBox, with one device acting as the chassis master and the rest as members. All components of member devices will appear on the master.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"domain": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"comments": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"owner_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			tagsKey:         tagsSchema,
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const virtualChassisAPIPath = "dcim/virtual-chassis/"

type virtualChassisRequest struct {
	Name         string             `json:"name"`
	Domain       string             `json:"domain,omitempty"`
	Description  string             `json:"description,omitempty"`
	Comments     string             `json:"comments,omitempty"`
	Owner        *int64             `json:"owner"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
}

type virtualChassisResponse struct {
	Name         string                 `json:"name"`
	Domain       string                 `json:"domain"`
	Description  string                 `json:"description"`
	Comments     string                 `json:"comments"`
	Owner        *netboxapi.Ref         `json:"owner"`
	CustomFields map[string]interface{} `json:"custom_fields"`
	Tags         []netboxapi.TagRef     `json:"tags"`
}

func virtualChassisRequestFromResourceData(api *providerState, d *schema.ResourceData) (*virtualChassisRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &virtualChassisRequest{
		Owner:       getOptionalInt(d, "owner_id"),
		Name:        name,
		Domain:      d.Get("domain").(string),
		Description: getOptionalStr(d, "description", true),
		Comments:    getOptionalStr(d, "comments", true),
		Tags:        tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxVirtualChassisCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	data, err := virtualChassisRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return diag.FromErr(err)
	}
	return createObjectCtx(ctx, d, m, virtualChassisAPIPath, data, resourceNetboxVirtualChassisRead)
}

func resourceNetboxVirtualChassisRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*providerState)
	return readObjectCtx(ctx, d, m, virtualChassisAPIPath, func(r *virtualChassisResponse) {
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("name", r.Name)
		d.Set("domain", r.Domain)
		d.Set("description", r.Description)
		d.Set("comments", r.Comments)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxVirtualChassisUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	data, err := virtualChassisRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return diag.FromErr(err)
	}
	return putObjectCtx(ctx, d, m, virtualChassisAPIPath, data, resourceNetboxVirtualChassisRead)
}

func resourceNetboxVirtualChassisDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return deleteObjectCtx(ctx, d, m, virtualChassisAPIPath)
}

// virtualChassisUpdateMaster sets or clears (master == nil) the master of a
// virtual chassis. Only the master is sent, a missing master is sent as null.
func virtualChassisUpdateMaster(api *providerState, id int64, master *int64) error {
	var value interface{}
	if master != nil {
		value = *master
	}

	body := map[string]interface{}{"master": value}
	_, err := netboxapi.Patch[netboxapi.Ref](context.Background(), api.rest, virtualChassisAPIPath, id, body)
	return err
}
