package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var resourceNetboxPrefixStatusOptions = []string{"active", "container", "reserved", "deprecated"}

func resourceNetboxPrefix() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxPrefixCreate,
		Read:   resourceNetboxPrefixRead,
		Update: resourceNetboxPrefixUpdate,
		Delete: resourceNetboxPrefixDelete,

		Description: `:meta:subcategory:IP Address Management (IPAM):From the [official documentation](https://docs.netbox.dev/en/stable/features/ipam/#prefixes):

> A prefix is an IPv4 or IPv6 network and mask expressed in CIDR notation (e.g. 192.0.2.0/24). A prefix entails only the "network portion" of an IP address: All bits in the address not covered by the mask must be zero. (In other words, a prefix cannot be a specific IP address.)
>
> Prefixes are automatically organized by their parent aggregates. Additionally, each prefix can be assigned to a particular site and virtual routing and forwarding instance (VRF). Each VRF represents a separate IP space or routing table. All prefixes not assigned to a VRF are considered to be in the "global" table.`,

		Schema: map[string]*schema.Schema{
			"prefix": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsCIDR,
			},
			"status": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(resourceNetboxPrefixStatusOptions, false),
				Description:  buildValidValueDescription(resourceNetboxPrefixStatusOptions),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"is_pool": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"mark_utilized": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"vrf_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"tenant_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"location_id": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"site_id", "site_group_id", "region_id"},
			},
			"site_id": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"location_id", "site_group_id", "region_id"},
			},
			"site_group_id": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"location_id", "site_id", "region_id"},
			},
			"region_id": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"location_id", "site_id", "site_group_id"},
			},
			"vlan_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"role_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			customFieldsKey: customFieldsSchema,
			tagsKey:         tagsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const prefixAPIPath = "ipam/prefixes/"

type prefixRequest struct {
	Prefix       string             `json:"prefix"`
	Status       string             `json:"status,omitempty"`
	Description  string             `json:"description,omitempty"`
	IsPool       bool               `json:"is_pool"`
	MarkUtilized bool               `json:"mark_utilized"`
	Vrf          *int64             `json:"vrf"`
	Tenant       *int64             `json:"tenant"`
	Vlan         *int64             `json:"vlan"`
	Role         *int64             `json:"role"`
	ScopeType    *string            `json:"scope_type"`
	ScopeID      *int64             `json:"scope_id"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
}

type prefixResponse struct {
	Prefix       string                 `json:"prefix"`
	Status       *netboxapi.Choice      `json:"status"`
	Description  string                 `json:"description"`
	IsPool       bool                   `json:"is_pool"`
	MarkUtilized bool                   `json:"mark_utilized"`
	Vrf          *netboxapi.Ref         `json:"vrf"`
	Tenant       *netboxapi.Ref         `json:"tenant"`
	Vlan         *netboxapi.Ref         `json:"vlan"`
	Role         *netboxapi.Ref         `json:"role"`
	ScopeType    *string                `json:"scope_type"`
	ScopeID      *int64                 `json:"scope_id"`
	CustomFields map[string]interface{} `json:"custom_fields"`
	Tags         []netboxapi.TagRef     `json:"tags"`
}

func prefixRequestFromResourceData(api *providerState, d *schema.ResourceData) (*prefixRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &prefixRequest{
		Prefix:       d.Get("prefix").(string),
		Status:       d.Get("status").(string),
		Description:  getOptionalStr(d, "description", true),
		IsPool:       d.Get("is_pool").(bool),
		MarkUtilized: d.Get("mark_utilized").(bool),

		Tags: tags,
	}
	if v, ok := d.GetOk("vrf_id"); ok {
		data.Vrf = int64ToPtr(int64(v.(int)))
	}
	if v, ok := d.GetOk("tenant_id"); ok {
		data.Tenant = int64ToPtr(int64(v.(int)))
	}
	if v, ok := d.GetOk("vlan_id"); ok {
		data.Vlan = int64ToPtr(int64(v.(int)))
	}
	if v, ok := d.GetOk("role_id"); ok {
		data.Role = int64ToPtr(int64(v.(int)))
	}

	siteID := getOptionalInt(d, "site_id")
	siteGroupID := getOptionalInt(d, "site_group_id")
	locationID := getOptionalInt(d, "location_id")
	regionID := getOptionalInt(d, "region_id")

	switch {
	case siteID != nil:
		data.ScopeType = strToPtr("dcim.site")
		data.ScopeID = siteID
	case siteGroupID != nil:
		data.ScopeType = strToPtr("dcim.sitegroup")
		data.ScopeID = siteGroupID
	case locationID != nil:
		data.ScopeType = strToPtr("dcim.location")
		data.ScopeID = locationID
	case regionID != nil:
		data.ScopeType = strToPtr("dcim.region")
		data.ScopeID = regionID
	default:
		data.ScopeType = nil
		data.ScopeID = nil
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxPrefixCreate(d *schema.ResourceData, m interface{}) error {
	data, err := prefixRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, prefixAPIPath, data, resourceNetboxPrefixRead)
}

func resourceNetboxPrefixRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, prefixAPIPath, func(r *prefixResponse) {
		d.Set("prefix", r.Prefix)
		d.Set("status", netboxapi.ChoiceValue(r.Status))
		d.Set("description", r.Description)
		d.Set("is_pool", r.IsPool)
		d.Set("mark_utilized", r.MarkUtilized)
		d.Set("vrf_id", netboxapi.RefID(r.Vrf))
		d.Set("tenant_id", netboxapi.RefID(r.Tenant))
		d.Set("vlan_id", netboxapi.RefID(r.Vlan))
		d.Set("role_id", netboxapi.RefID(r.Role))

		d.Set("site_id", nil)
		d.Set("site_group_id", nil)
		d.Set("location_id", nil)
		d.Set("region_id", nil)
		if r.ScopeType != nil && r.ScopeID != nil {
			switch *r.ScopeType {
			case "dcim.site":
				d.Set("site_id", *r.ScopeID)
			case "dcim.sitegroup":
				d.Set("site_group_id", *r.ScopeID)
			case "dcim.location":
				d.Set("location_id", *r.ScopeID)
			case "dcim.region":
				d.Set("region_id", *r.ScopeID)
			}
		}
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxPrefixUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := prefixRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return putObject(d, m, prefixAPIPath, data, resourceNetboxPrefixRead)
}

func resourceNetboxPrefixDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, prefixAPIPath)
}
