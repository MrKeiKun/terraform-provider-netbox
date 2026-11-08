package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxVlanTranslationRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxVlanTranslationRuleCreate,
		Read:   resourceNetboxVlanTranslationRuleRead,
		Update: resourceNetboxVlanTranslationRuleUpdate,
		Delete: resourceNetboxVlanTranslationRuleDelete,

		Description: `:meta:subcategory:IP Address Management (IPAM):From the [official documentation](https://docs.netbox.dev/en/stable/models/ipam/vlantranslationrule/):

> A VLAN translation rule maps a local VLAN ID to a remote VLAN ID.`,

		Schema: map[string]*schema.Schema{
			"policy_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"local_vid": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 4094),
				Description:  "Numeric VLAN ID (1-4094)",
			},
			"remote_vid": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 4094),
				Description:  "Numeric VLAN ID (1-4094)",
			},
			"description": {
				Type:     schema.TypeString,
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

const vlanTranslationRuleAPIPath = "ipam/vlan-translation-rules/"

type vlanTranslationRuleRequest struct {
	Policy       *int64             `json:"policy"`
	LocalVid     int64              `json:"local_vid"`
	RemoteVid    int64              `json:"remote_vid"`
	Description  string             `json:"description,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
}

type vlanTranslationRuleResponse struct {
	Policy       *int64                 `json:"policy"`
	LocalVid     *int64                 `json:"local_vid"`
	RemoteVid    *int64                 `json:"remote_vid"`
	Description  string                 `json:"description"`
	Tags         []netboxapi.TagRef     `json:"tags"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func vlanTranslationRuleRequestFromResourceData(api *providerState, d *schema.ResourceData) (*vlanTranslationRuleRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &vlanTranslationRuleRequest{
		Policy:      int64ToPtr(int64(d.Get("policy_id").(int))),
		LocalVid:    int64(d.Get("local_vid").(int)),
		RemoteVid:   int64(d.Get("remote_vid").(int)),
		Description: getOptionalStr(d, "description", true),
		Tags:        tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxVlanTranslationRuleCreate(d *schema.ResourceData, m interface{}) error {
	data, err := vlanTranslationRuleRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, vlanTranslationRuleAPIPath, data, resourceNetboxVlanTranslationRuleRead)
}

func resourceNetboxVlanTranslationRuleRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, vlanTranslationRuleAPIPath, func(r *vlanTranslationRuleResponse) {
		d.Set("policy_id", r.Policy)
		d.Set("local_vid", r.LocalVid)
		d.Set("remote_vid", r.RemoteVid)
		d.Set("description", r.Description)
		api.readTagRefs(d, r.Tags)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
	})
}

func resourceNetboxVlanTranslationRuleUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := vlanTranslationRuleRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, vlanTranslationRuleAPIPath, data, resourceNetboxVlanTranslationRuleRead)
}

func resourceNetboxVlanTranslationRuleDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, vlanTranslationRuleAPIPath)
}
