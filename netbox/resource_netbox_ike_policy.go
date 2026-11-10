package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxIKEPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: withContext(resourceNetboxIKEPolicyCreate),
		ReadContext:   withContext(resourceNetboxIKEPolicyRead),
		UpdateContext: withContext(resourceNetboxIKEPolicyUpdate),
		DeleteContext: withContext(resourceNetboxIKEPolicyDelete),

		Description: `:meta:subcategory:VPN Tunnels:From the [official documentation](https://docs.netbox.dev/en/stable/models/vpn/ikepolicy/):

> An IKE policy bundles IKE proposals with the version and mode of the first phase of an IPsec negotiation.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the ike policy.",
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the ike policy.",
				Optional:    true,
			},
			"version": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "One of [1, 2]",
				ValidateFunc: validation.IntInSlice([]int{1, 2}),
			},
			"mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "One of [aggressive, main]",
				ValidateFunc: validation.StringInSlice([]string{"aggressive", "main"}, false),
			},
			"proposal_ids": {
				Type:        schema.TypeSet,
				Description: "The IDs of the proposal associated with the ike policy.",
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"preshared_key": {
				Type:        schema.TypeString,
				Description: "The preshared key of the ike policy.",
				Optional:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the ike policy.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the ike policy.",
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

const iKEPolicyAPIPath = "vpn/ike-policies/"

type iKEPolicyRequest struct {
	Name         string             `json:"name"`
	Description  string             `json:"description,omitempty"`
	Version      int64              `json:"version,omitempty"`
	Mode         *string            `json:"mode,omitempty"`
	Proposals    []int64            `json:"proposals"`
	PresharedKey string             `json:"preshared_key,omitempty"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
}

type iKEPolicyResponse struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Version      *netboxapi.IntChoice   `json:"version"`
	Mode         *netboxapi.Choice      `json:"mode"`
	Proposals    []netboxapi.Ref        `json:"proposals"`
	PresharedKey string                 `json:"preshared_key"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	Tags         []netboxapi.TagRef     `json:"tags"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func iKEPolicyRequestFromResourceData(api *providerState, d *schema.ResourceData) (*iKEPolicyRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &iKEPolicyRequest{
		Name:         d.Get("name").(string),
		Description:  getOptionalStr(d, "description", true),
		Version:      int64(d.Get("version").(int)),
		Mode:         getOptionalStrClear(d, "mode"),
		Proposals:    getIDSet(d, "proposal_ids"),
		PresharedKey: getOptionalStr(d, "preshared_key", true),
		Owner:        getOptionalInt(d, "owner_id"),
		Comments:     getOptionalStr(d, "comments", true),
		Tags:         tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxIKEPolicyCreate(d *schema.ResourceData, m interface{}) error {
	data, err := iKEPolicyRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, iKEPolicyAPIPath, data, resourceNetboxIKEPolicyRead)
}

func resourceNetboxIKEPolicyRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, iKEPolicyAPIPath, func(r *iKEPolicyResponse) {
		d.Set("name", r.Name)
		d.Set("description", r.Description)
		d.Set("version", netboxapi.IntChoiceValue(r.Version))
		d.Set("mode", netboxapi.ChoiceValue(r.Mode))
		d.Set("proposal_ids", netboxapi.IDs(r.Proposals))
		d.Set("preshared_key", r.PresharedKey)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("comments", r.Comments)
		api.readTagRefs(d, r.Tags)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
	})
}

func resourceNetboxIKEPolicyUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := iKEPolicyRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, iKEPolicyAPIPath, data, resourceNetboxIKEPolicyRead)
}

func resourceNetboxIKEPolicyDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, iKEPolicyAPIPath)
}
