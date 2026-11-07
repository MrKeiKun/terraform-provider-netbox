package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var resourceNetboxVlanGroupScopeTypeOptions = []string{"dcim.location", "dcim.site", "dcim.sitegroup", "dcim.region", "dcim.rack", "virtualization.cluster", "virtualization.clustergroup"}

func resourceNetboxVlanGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxVlanGroupCreate,
		Read:   resourceNetboxVlanGroupRead,
		Update: resourceNetboxVlanGroupUpdate,
		Delete: resourceNetboxVlanGroupDelete,

		Description: `:meta:subcategory:IP Address Management (IPAM):

> A VLAN Group represents a collection of VLANs. Generally, these are limited by one of a number of scopes such as "Site" or "Virtualization Cluster".`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"slug": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"scope_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(resourceNetboxVlanGroupScopeTypeOptions, false),
				Description:  buildValidValueDescription(resourceNetboxVlanGroupScopeTypeOptions),
			},
			"scope_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				RequiredWith: []string{"scope_type"},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"vid_ranges": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeList,
					Elem: &schema.Schema{
						Type: schema.TypeInt,
					},
				},
				Required: true,
			},
			"owner_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"comments": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tenant_id": {
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

const vlanGroupAPIPath = "ipam/vlan-groups/"

type vlanGroupRequest struct {
	Name         string             `json:"name"`
	Slug         string             `json:"slug"`
	Description  string             `json:"description,omitempty"`
	VidRanges    [][]int64          `json:"vid_ranges"`
	ScopeType    *string            `json:"scope_type,omitempty"`
	ScopeID      *int64             `json:"scope_id,omitempty"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	Tenant       *int64             `json:"tenant"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
}

type vlanGroupResponse struct {
	Name         string                 `json:"name"`
	Slug         string                 `json:"slug"`
	Description  string                 `json:"description"`
	VidRanges    [][]int64              `json:"vid_ranges"`
	ScopeType    *string                `json:"scope_type"`
	ScopeID      *int64                 `json:"scope_id"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	Tenant       *netboxapi.Ref         `json:"tenant"`
	CustomFields map[string]interface{} `json:"custom_fields"`
	Tags         []netboxapi.TagRef     `json:"tags"`
}

func vlanGroupRequestFromResourceData(api *providerState, d *schema.ResourceData) (*vlanGroupRequest, error) {
	vidRanges := make([][]int64, 0)
	for _, v := range d.Get("vid_ranges").([]interface{}) {
		inner := v.([]interface{})
		vidRanges = append(vidRanges, []int64{int64(inner[0].(int)), int64(inner[1].(int))})
	}

	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}

	data := &vlanGroupRequest{
		CustomFields: customFieldsRequest(d),
		Tenant:       getOptionalInt(d, "tenant_id"),
		Comments:     getOptionalStr(d, "comments", true),
		Owner:        getOptionalInt(d, "owner_id"),
		Name:         d.Get("name").(string),
		Slug:         d.Get("slug").(string),
		Description:  d.Get("description").(string),
		VidRanges:    vidRanges,
		Tags:         tags,
	}
	if scopeType, ok := d.GetOk("scope_type"); ok {
		data.ScopeType = strToPtr(scopeType.(string))
	}
	if scopeID, ok := d.GetOk("scope_id"); ok {
		data.ScopeID = int64ToPtr(int64(scopeID.(int)))
	}
	return data, nil
}

func resourceNetboxVlanGroupCreate(d *schema.ResourceData, m interface{}) error {
	data, err := vlanGroupRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, vlanGroupAPIPath, data, resourceNetboxVlanGroupRead)
}

func resourceNetboxVlanGroupRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, vlanGroupAPIPath, func(vlanGroup *vlanGroupResponse) {
		d.Set(customFieldsKey, getCustomFields(vlanGroup.CustomFields))
		d.Set("tenant_id", netboxapi.RefID(vlanGroup.Tenant))
		d.Set("comments", vlanGroup.Comments)
		d.Set("owner_id", netboxapi.RefID(vlanGroup.Owner))
		d.Set("name", vlanGroup.Name)
		d.Set("slug", vlanGroup.Slug)
		d.Set("description", vlanGroup.Description)
		d.Set("vid_ranges", vlanGroup.VidRanges)
		api.readTagRefs(d, vlanGroup.Tags)
		if vlanGroup.ScopeType != nil {
			d.Set("scope_type", vlanGroup.ScopeType)
		}
		if vlanGroup.ScopeID != nil {
			d.Set("scope_id", vlanGroup.ScopeID)
		}
	})
}

func resourceNetboxVlanGroupUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := vlanGroupRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return putObject(d, m, vlanGroupAPIPath, data, resourceNetboxVlanGroupRead)
}

func resourceNetboxVlanGroupDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, vlanGroupAPIPath)
}
