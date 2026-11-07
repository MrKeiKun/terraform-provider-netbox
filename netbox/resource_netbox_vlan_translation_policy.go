package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxVlanTranslationPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxVlanTranslationPolicyCreate,
		Read:   resourceNetboxVlanTranslationPolicyRead,
		Update: resourceNetboxVlanTranslationPolicyUpdate,
		Delete: resourceNetboxVlanTranslationPolicyDelete,

		Description: `:meta:subcategory:IP Address Management (IPAM):From the [official documentation](https://docs.netbox.dev/en/stable/models/ipam/vlantranslationpolicy/):

> A VLAN translation policy groups the rules that map local VLAN IDs to remote VLAN IDs.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
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
			tagsKey:         tagsSchema,
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const vlanTranslationPolicyAPIPath = "ipam/vlan-translation-policies/"

type vlanTranslationPolicyRequest struct {
	Name         string             `json:"name"`
	Description  string             `json:"description,omitempty"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
}

type vlanTranslationPolicyResponse struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	Tags         []netboxapi.TagRef     `json:"tags"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func vlanTranslationPolicyRequestFromResourceData(api *providerState, d *schema.ResourceData) (*vlanTranslationPolicyRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &vlanTranslationPolicyRequest{
		Name:        d.Get("name").(string),
		Description: getOptionalStr(d, "description", true),
		Owner:       getOptionalInt(d, "owner_id"),
		Comments:    getOptionalStr(d, "comments", true),
		Tags:        tags,
	}
	if ct, ok := d.GetOk(customFieldsKey); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxVlanTranslationPolicyCreate(d *schema.ResourceData, m interface{}) error {
	data, err := vlanTranslationPolicyRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, vlanTranslationPolicyAPIPath, data, resourceNetboxVlanTranslationPolicyRead)
}

func resourceNetboxVlanTranslationPolicyRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, vlanTranslationPolicyAPIPath, func(r *vlanTranslationPolicyResponse) {
		d.Set("name", r.Name)
		d.Set("description", r.Description)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("comments", r.Comments)
		api.readTagRefs(d, r.Tags)
		if cf := getCustomFields(r.CustomFields); cf != nil {
			d.Set(customFieldsKey, cf)
		}
	})
}

func resourceNetboxVlanTranslationPolicyUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := vlanTranslationPolicyRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, vlanTranslationPolicyAPIPath, data, resourceNetboxVlanTranslationPolicyRead)
}

func resourceNetboxVlanTranslationPolicyDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, vlanTranslationPolicyAPIPath)
}
