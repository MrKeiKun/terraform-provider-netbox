package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var resourceNetboxVlanStatusOptions = []string{"active", "reserved", "deprecated"}

func resourceNetboxVlan() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxVlanCreate,
		Read:   resourceNetboxVlanRead,
		Update: resourceNetboxVlanUpdate,
		Delete: resourceNetboxVlanDelete,

		Description: `:meta:subcategory:IP Address Management (IPAM):From the [official documentation](https://docs.netbox.dev/en/stable/features/vlans/#vlans):

> A VLAN represents an isolated layer two domain, identified by a name and a numeric ID (1-4094) as defined in IEEE 802.1Q. VLANs are arranged into VLAN groups to define scope and to enforce uniqueness.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"vid": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "active",
				ValidateFunc: validation.StringInSlice(resourceNetboxVlanStatusOptions, false),
				Description:  buildValidValueDescription(resourceNetboxVlanStatusOptions),
			},
			"group_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"tenant_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},

			"role_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},

			"site_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"owner_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"comments": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"qinq_role": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "One of [svlan, cvlan]",
				ValidateFunc: validation.StringInSlice([]string{"svlan", "cvlan"}, false),
			},
			"qinq_svlan": {
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

const vlanAPIPath = "ipam/vlans/"

type vlanRequest struct {
	Name         string             `json:"name"`
	Vid          *int64             `json:"vid"`
	Status       string             `json:"status,omitempty"`
	Description  string             `json:"description,omitempty"`
	Group        *int64             `json:"group,omitempty"`
	Site         *int64             `json:"site,omitempty"`
	Tenant       *int64             `json:"tenant,omitempty"`
	Role         *int64             `json:"role,omitempty"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	QinqRole     *string            `json:"qinq_role,omitempty"`
	QinqSvlan    *int64             `json:"qinq_svlan"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
}

type vlanResponse struct {
	Name         string                 `json:"name"`
	Vid          *int64                 `json:"vid"`
	Status       *netboxapi.Choice      `json:"status"`
	Description  string                 `json:"description"`
	Group        *netboxapi.Ref         `json:"group"`
	Site         *netboxapi.Ref         `json:"site"`
	Tenant       *netboxapi.Ref         `json:"tenant"`
	Role         *netboxapi.Ref         `json:"role"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	QinqRole     *netboxapi.Choice      `json:"qinq_role"`
	QinqSvlan    *int64                 `json:"qinq_svlan"`
	CustomFields map[string]interface{} `json:"custom_fields"`
	Tags         []netboxapi.TagRef     `json:"tags"`
}

func vlanRequestFromResourceData(api *providerState, d *schema.ResourceData) (*vlanRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &vlanRequest{
		QinqSvlan:   getOptionalInt(d, "qinq_svlan"),
		QinqRole:    getOptionalStrClear(d, "qinq_role"),
		Comments:    getOptionalStr(d, "comments", true),
		Owner:       getOptionalInt(d, "owner_id"),
		Name:        name,
		Vid:         int64ToPtr(int64(d.Get("vid").(int))),
		Status:      d.Get("status").(string),
		Description: d.Get("description").(string),
		Tags:        tags,
	}
	if v, ok := d.GetOk("group_id"); ok {
		data.Group = int64ToPtr(int64(v.(int)))
	}
	if v, ok := d.GetOk("site_id"); ok {
		data.Site = int64ToPtr(int64(v.(int)))
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

func resourceNetboxVlanCreate(d *schema.ResourceData, m interface{}) error {
	data, err := vlanRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, vlanAPIPath, data, resourceNetboxVlanRead)
}

func resourceNetboxVlanRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, vlanAPIPath, func(r *vlanResponse) {
		d.Set("qinq_svlan", r.QinqSvlan)
		d.Set("qinq_role", netboxapi.ChoiceValue(r.QinqRole))
		d.Set("comments", r.Comments)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("name", r.Name)
		d.Set("vid", r.Vid)
		d.Set("status", netboxapi.ChoiceValue(r.Status))
		d.Set("description", r.Description)
		if r.Group != nil {
			d.Set("group_id", r.Group.ID)
		}
		if r.Site != nil {
			d.Set("site_id", r.Site.ID)
		}
		if r.Tenant != nil {
			d.Set("tenant_id", r.Tenant.ID)
		}
		if r.Role != nil {
			d.Set("role_id", r.Role.ID)
		}
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxVlanUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := vlanRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return putObject(d, m, vlanAPIPath, data, resourceNetboxVlanRead)
}

func resourceNetboxVlanDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, vlanAPIPath)
}
