package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxIPSecProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxIPSecProfileCreate,
		Read:   resourceNetboxIPSecProfileRead,
		Update: resourceNetboxIPSecProfileUpdate,
		Delete: resourceNetboxIPSecProfileDelete,

		Description: `:meta:subcategory:VPN Tunnels:From the [official documentation](https://docs.netbox.dev/en/stable/models/vpn/ipsecprofile/):

> An IPsec profile combines an IKE policy and an IPsec policy for a tunnel.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"mode": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "One of [esp, ah]",
				ValidateFunc: validation.StringInSlice([]string{"esp", "ah"}, false),
			},
			"ike_policy_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"ipsec_policy_id": {
				Type:     schema.TypeInt,
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
			tagsKey:         tagsSchema,
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const iPSecProfileAPIPath = "vpn/ipsec-profiles/"

type iPSecProfileRequest struct {
	Name         string             `json:"name"`
	Description  string             `json:"description,omitempty"`
	Mode         string             `json:"mode,omitempty"`
	IkePolicy    *int64             `json:"ike_policy"`
	IpsecPolicy  *int64             `json:"ipsec_policy"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
}

type iPSecProfileResponse struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Mode         *netboxapi.Choice      `json:"mode"`
	IkePolicy    *netboxapi.Ref         `json:"ike_policy"`
	IpsecPolicy  *netboxapi.Ref         `json:"ipsec_policy"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	Tags         []netboxapi.TagRef     `json:"tags"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func iPSecProfileRequestFromResourceData(api *providerState, d *schema.ResourceData) (*iPSecProfileRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &iPSecProfileRequest{
		Name:        d.Get("name").(string),
		Description: getOptionalStr(d, "description", true),
		Mode:        d.Get("mode").(string),
		IkePolicy:   int64ToPtr(int64(d.Get("ike_policy_id").(int))),
		IpsecPolicy: int64ToPtr(int64(d.Get("ipsec_policy_id").(int))),
		Owner:       getOptionalInt(d, "owner_id"),
		Comments:    getOptionalStr(d, "comments", true),
		Tags:        tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxIPSecProfileCreate(d *schema.ResourceData, m interface{}) error {
	data, err := iPSecProfileRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, iPSecProfileAPIPath, data, resourceNetboxIPSecProfileRead)
}

func resourceNetboxIPSecProfileRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, iPSecProfileAPIPath, func(r *iPSecProfileResponse) {
		d.Set("name", r.Name)
		d.Set("description", r.Description)
		d.Set("mode", netboxapi.ChoiceValue(r.Mode))
		d.Set("ike_policy_id", netboxapi.RefID(r.IkePolicy))
		d.Set("ipsec_policy_id", netboxapi.RefID(r.IpsecPolicy))
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("comments", r.Comments)
		api.readTagRefs(d, r.Tags)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
	})
}

func resourceNetboxIPSecProfileUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := iPSecProfileRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, iPSecProfileAPIPath, data, resourceNetboxIPSecProfileRead)
}

func resourceNetboxIPSecProfileDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, iPSecProfileAPIPath)
}
