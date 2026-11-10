package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var resourceNetboxIPRangeStatusOptions = []string{"active", "reserved", "deprecated"}

func resourceNetboxIPRange() *schema.Resource {
	return &schema.Resource{
		CreateContext: withContext(resourceNetboxIPRangeCreate),
		ReadContext:   withContext(resourceNetboxIPRangeRead),
		UpdateContext: withContext(resourceNetboxIPRangeUpdate),
		DeleteContext: withContext(resourceNetboxIPRangeDelete),

		Description: `:meta:subcategory:IP Address Management (IPAM):From the [official documentation](https://docs.netbox.dev/en/stable/features/ipam/#ip-ranges):

> This model represents an arbitrary range of individual IPv4 or IPv6 addresses, inclusive of its starting and ending addresses. For instance, the range 192.0.2.10 to 192.0.2.20 has eleven members. (The total member count is available as the size property on an IPRange instance.) Like prefixes and IP addresses, each IP range may optionally be assigned to a VRF and/or tenant.`,

		Schema: map[string]*schema.Schema{
			"start_address": {
				Type:        schema.TypeString,
				Description: "The start address of the ip range.",
				Required:    true,
			},
			"end_address": {
				Type:        schema.TypeString,
				Description: "The end address of the ip range.",
				Required:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "active",
				Description: buildValidValueDescription(resourceNetboxIPRangeStatusOptions),
			},
			"tenant_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the tenant associated with the ip range.",
				Optional:    true,
			},
			"role_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the role associated with the ip range.",
				Optional:    true,
			},
			"vrf_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the vrf associated with the ip range.",
				Optional:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the ip range.",
				Optional:    true,
			},
			"mark_populated": {
				Type:        schema.TypeBool,
				Description: "Whether mark populated is enabled for the ip range.",
				Optional:    true,
				Computed:    true,
			},
			"mark_utilized": {
				Type:        schema.TypeBool,
				Description: "Whether the ip range is treated as fully utilized.",
				Optional:    true,
				Computed:    true,
			},
			"size": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The total member count of the IP range",
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the ip range.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the ip range.",
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

const ipRangeAPIPath = "ipam/ip-ranges/"

type ipRangeRequest struct {
	StartAddress  string             `json:"start_address"`
	EndAddress    string             `json:"end_address"`
	Status        string             `json:"status,omitempty"`
	Description   string             `json:"description,omitempty"`
	MarkPopulated bool               `json:"mark_populated"`
	MarkUtilized  bool               `json:"mark_utilized"`
	Vrf           *int64             `json:"vrf,omitempty"`
	Tenant        *int64             `json:"tenant,omitempty"`
	Role          *int64             `json:"role,omitempty"`
	Owner         *int64             `json:"owner"`
	Comments      string             `json:"comments,omitempty"`
	CustomFields  interface{}        `json:"custom_fields,omitempty"`
	Tags          []netboxapi.TagRef `json:"tags"`
}

type ipRangeResponse struct {
	StartAddress  string                 `json:"start_address"`
	EndAddress    string                 `json:"end_address"`
	Status        *netboxapi.Choice      `json:"status"`
	Description   string                 `json:"description"`
	MarkPopulated bool                   `json:"mark_populated"`
	MarkUtilized  bool                   `json:"mark_utilized"`
	Vrf           *netboxapi.Ref         `json:"vrf"`
	Tenant        *netboxapi.Ref         `json:"tenant"`
	Role          *netboxapi.Ref         `json:"role"`
	Size          int64                  `json:"size"`
	Owner         *netboxapi.Ref         `json:"owner"`
	Comments      string                 `json:"comments"`
	CustomFields  map[string]interface{} `json:"custom_fields"`
	Tags          []netboxapi.TagRef     `json:"tags"`
}

func ipRangeRequestFromResourceData(api *providerState, d *schema.ResourceData) (*ipRangeRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &ipRangeRequest{
		Comments:      getOptionalStr(d, "comments", true),
		Owner:         getOptionalInt(d, "owner_id"),
		StartAddress:  d.Get("start_address").(string),
		EndAddress:    d.Get("end_address").(string),
		Status:        d.Get("status").(string),
		Description:   d.Get("description").(string),
		MarkPopulated: d.Get("mark_populated").(bool),
		MarkUtilized:  d.Get("mark_utilized").(bool),
		Tags:          tags,
	}
	if v, ok := d.GetOk("vrf_id"); ok {
		data.Vrf = int64ToPtr(int64(v.(int)))
	}
	if v, ok := d.GetOk("tenant_id"); ok {
		data.Tenant = int64ToPtr(int64(v.(int)))
	}
	if v, ok := d.GetOk("role_id"); ok {
		data.Role = int64ToPtr(int64(v.(int)))
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxIPRangeCreate(d *schema.ResourceData, m interface{}) error {
	data, err := ipRangeRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, ipRangeAPIPath, data, resourceNetboxIPRangeUpdate)
}

func resourceNetboxIPRangeRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, ipRangeAPIPath, func(r *ipRangeResponse) {
		d.Set("comments", r.Comments)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("start_address", r.StartAddress)
		d.Set("end_address", r.EndAddress)
		d.Set("status", netboxapi.ChoiceValue(r.Status))
		if r.Description != "" {
			d.Set("description", r.Description)
		}
		d.Set("mark_populated", r.MarkPopulated)
		d.Set("mark_utilized", r.MarkUtilized)
		if r.Vrf != nil {
			d.Set("vrf_id", r.Vrf.ID)
		}
		if r.Tenant != nil {
			d.Set("tenant_id", r.Tenant.ID)
		}
		if r.Role != nil {
			d.Set("role_id", r.Role.ID)
		}
		if r.Size != 0 {
			d.Set("size", r.Size)
		}
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxIPRangeUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := ipRangeRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return putObject(d, m, ipRangeAPIPath, data, resourceNetboxIPRangeRead)
}

func resourceNetboxIPRangeDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, ipRangeAPIPath)
}
