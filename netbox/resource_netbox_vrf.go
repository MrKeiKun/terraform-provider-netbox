package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxVrf() *schema.Resource {
	return &schema.Resource{
		CreateContext: withContext(resourceNetboxVrfCreate),
		ReadContext:   withContext(resourceNetboxVrfRead),
		UpdateContext: withContext(resourceNetboxVrfUpdate),
		DeleteContext: withContext(resourceNetboxVrfDelete),

		Description: `:meta:subcategory:IP Address Management (IPAM):From the [official documentation](https://docs.netbox.dev/en/stable/features/ipam/#virtual-routing-and-forwarding-vrf):

> A VRF object in NetBox represents a virtual routing and forwarding (VRF) domain. Each VRF is essentially a separate routing table. VRFs are commonly used to isolate customers or organizations from one another within a network, or to route overlapping address space (e.g. multiple instances of the 10.0.0.0/8 space). Each VRF may be assigned to a specific tenant to aid in organizing the available IP space by customer or internal user.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the vrf.",
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the vrf.",
				Optional:    true,
			},
			"tenant_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the tenant associated with the vrf.",
				Optional:    true,
			},
			"enforce_unique": {
				Type:        schema.TypeBool,
				Description: "Whether enforce unique is enabled for the vrf.",
				Optional:    true,
				Default:     true,
			},
			"rd": {
				Type:         schema.TypeString,
				Description:  "The route distinguisher of the vrf.",
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 21),
			},

			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the vrf.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the vrf.",
				Optional:    true,
			},
			customFieldsKey: customFieldsSchema,
			tagsKey:         tagsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const vrfAPIPath = "ipam/vrfs/"

type vrfRequest struct {
	Name          string             `json:"name"`
	Rd            *string            `json:"rd"`
	Tenant        *int64             `json:"tenant,omitempty"`
	Description   string             `json:"description,omitempty"`
	EnforceUnique bool               `json:"enforce_unique"`
	ExportTargets []int64            `json:"export_targets"`
	ImportTargets []int64            `json:"import_targets"`
	Owner         *int64             `json:"owner"`
	Comments      string             `json:"comments,omitempty"`
	CustomFields  interface{}        `json:"custom_fields,omitempty"`
	Tags          []netboxapi.TagRef `json:"tags"`
}

type vrfResponse struct {
	Name          string                 `json:"name"`
	Rd            *string                `json:"rd"`
	Tenant        *netboxapi.Ref         `json:"tenant"`
	Description   string                 `json:"description"`
	EnforceUnique bool                   `json:"enforce_unique"`
	Owner         *netboxapi.Ref         `json:"owner"`
	Comments      string                 `json:"comments"`
	CustomFields  map[string]interface{} `json:"custom_fields"`
	Tags          []netboxapi.TagRef     `json:"tags"`
}

func vrfRequestFromResourceData(api *providerState, d *schema.ResourceData) (*vrfRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}

	data := &vrfRequest{
		CustomFields:  customFieldsRequest(d),
		Comments:      getOptionalStr(d, "comments", true),
		Owner:         getOptionalInt(d, "owner_id"),
		Name:          d.Get("name").(string),
		Description:   getOptionalStr(d, "description", true),
		EnforceUnique: d.Get("enforce_unique").(bool),
		ExportTargets: []int64{},
		ImportTargets: []int64{},
		Tags:          tags,
	}
	if rd, ok := d.GetOk("rd"); ok {
		data.Rd = strToPtr(rd.(string))
	}
	if tenantID, ok := d.GetOk("tenant_id"); ok {
		data.Tenant = int64ToPtr(int64(tenantID.(int)))
	}
	return data, nil
}

func resourceNetboxVrfCreate(d *schema.ResourceData, m interface{}) error {
	data, err := vrfRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, vrfAPIPath, data, resourceNetboxVrfRead)
}

func resourceNetboxVrfRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, vrfAPIPath, func(vrf *vrfResponse) {
		d.Set(customFieldsKey, getCustomFields(vrf.CustomFields))
		d.Set("comments", vrf.Comments)
		d.Set("owner_id", netboxapi.RefID(vrf.Owner))
		d.Set("name", vrf.Name)
		d.Set("description", vrf.Description)
		d.Set("enforce_unique", vrf.EnforceUnique)
		if vrf.Rd != nil {
			d.Set("rd", *vrf.Rd)
		} else {
			d.Set("rd", nil)
		}
		d.Set("tenant_id", netboxapi.RefID(vrf.Tenant))
		api.readTagRefs(d, vrf.Tags)
	})
}

func resourceNetboxVrfUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := vrfRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, vrfAPIPath, data, resourceNetboxVrfRead)
}

func resourceNetboxVrfDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, vrfAPIPath)
}
