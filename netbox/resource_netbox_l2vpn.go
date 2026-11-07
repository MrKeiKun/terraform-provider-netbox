package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxL2VPN() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxL2VPNCreate,
		Read:   resourceNetboxL2VPNRead,
		Update: resourceNetboxL2VPNUpdate,
		Delete: resourceNetboxL2VPNDelete,

		Description: `:meta:subcategory:VPN Tunnels:From the [official documentation](https://docs.netbox.dev/en/stable/models/vpn/l2vpn/):

> An L2VPN is a layer 2 virtual network, such as a VXLAN or an MPLS EVPN, with its terminations.`,

		Schema: map[string]*schema.Schema{
			"identifier": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(-9223372036854775808, 9223372036854775807),
			},
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
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "One of [vpws, vpls, vxlan, vxlan-evpn, mpls-evpn, pbb-evpn, evpn-vpws, epl, evpl, ep-lan, evp-lan, ep-tree, evp-tree, spb]",
				ValidateFunc: validation.StringInSlice([]string{"vpws", "vpls", "vxlan", "vxlan-evpn", "mpls-evpn", "pbb-evpn", "evpn-vpws", "epl", "evpl", "ep-lan", "evp-lan", "ep-tree", "evp-tree", "spb"}, false),
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "One of [active, planned, decommissioning]",
				ValidateFunc: validation.StringInSlice([]string{"active", "planned", "decommissioning"}, false),
			},
			"import_target_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
			"export_target_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
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
			tagsKey:         tagsSchema,
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const l2VPNAPIPath = "vpn/l2vpns/"

type l2VPNRequest struct {
	Identifier    *int64             `json:"identifier"`
	Name          string             `json:"name"`
	Slug          string             `json:"slug"`
	Type          string             `json:"type,omitempty"`
	Status        string             `json:"status,omitempty"`
	ImportTargets []int64            `json:"import_targets"`
	ExportTargets []int64            `json:"export_targets"`
	Description   string             `json:"description,omitempty"`
	Owner         *int64             `json:"owner"`
	Comments      string             `json:"comments,omitempty"`
	Tenant        *int64             `json:"tenant"`
	Tags          []netboxapi.TagRef `json:"tags"`
	CustomFields  interface{}        `json:"custom_fields,omitempty"`
}

type l2VPNResponse struct {
	Identifier    *int64                 `json:"identifier"`
	Name          string                 `json:"name"`
	Slug          string                 `json:"slug"`
	Type          *netboxapi.Choice      `json:"type"`
	Status        *netboxapi.Choice      `json:"status"`
	ImportTargets []netboxapi.Ref        `json:"import_targets"`
	ExportTargets []netboxapi.Ref        `json:"export_targets"`
	Description   string                 `json:"description"`
	Owner         *netboxapi.Ref         `json:"owner"`
	Comments      string                 `json:"comments"`
	Tenant        *netboxapi.Ref         `json:"tenant"`
	Tags          []netboxapi.TagRef     `json:"tags"`
	CustomFields  map[string]interface{} `json:"custom_fields"`
}

func l2VPNRequestFromResourceData(api *providerState, d *schema.ResourceData) (*l2VPNRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &l2VPNRequest{
		Identifier:    getOptionalInt(d, "identifier"),
		Name:          d.Get("name").(string),
		Slug:          slugFromResourceData(d, name),
		Type:          d.Get("type").(string),
		Status:        d.Get("status").(string),
		ImportTargets: getIDSet(d, "import_target_ids"),
		ExportTargets: getIDSet(d, "export_target_ids"),
		Description:   getOptionalStr(d, "description", true),
		Owner:         getOptionalInt(d, "owner_id"),
		Comments:      getOptionalStr(d, "comments", true),
		Tenant:        getOptionalInt(d, "tenant_id"),
		Tags:          tags,
	}
	if ct, ok := d.GetOk(customFieldsKey); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxL2VPNCreate(d *schema.ResourceData, m interface{}) error {
	data, err := l2VPNRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, l2VPNAPIPath, data, resourceNetboxL2VPNRead)
}

func resourceNetboxL2VPNRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, l2VPNAPIPath, func(r *l2VPNResponse) {
		d.Set("identifier", r.Identifier)
		d.Set("name", r.Name)
		d.Set("slug", r.Slug)
		d.Set("type", netboxapi.ChoiceValue(r.Type))
		d.Set("status", netboxapi.ChoiceValue(r.Status))
		d.Set("import_target_ids", netboxapi.IDs(r.ImportTargets))
		d.Set("export_target_ids", netboxapi.IDs(r.ExportTargets))
		d.Set("description", r.Description)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("comments", r.Comments)
		d.Set("tenant_id", netboxapi.RefID(r.Tenant))
		api.readTagRefs(d, r.Tags)
		if cf := getCustomFields(r.CustomFields); cf != nil {
			d.Set(customFieldsKey, cf)
		}
	})
}

func resourceNetboxL2VPNUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := l2VPNRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, l2VPNAPIPath, data, resourceNetboxL2VPNRead)
}

func resourceNetboxL2VPNDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, l2VPNAPIPath)
}
