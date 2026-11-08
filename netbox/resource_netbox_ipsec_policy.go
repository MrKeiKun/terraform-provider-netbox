package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxIPSecPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxIPSecPolicyCreate,
		Read:   resourceNetboxIPSecPolicyRead,
		Update: resourceNetboxIPSecPolicyUpdate,
		Delete: resourceNetboxIPSecPolicyDelete,

		Description: `:meta:subcategory:VPN Tunnels:From the [official documentation](https://docs.netbox.dev/en/stable/models/vpn/ipsecpolicy/):

> An IPsec policy bundles IPsec proposals with the Perfect Forward Secrecy group.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the ipsec policy.",
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the ipsec policy.",
				Optional:    true,
			},
			"proposal_ids": {
				Type:        schema.TypeSet,
				Description: "The IDs of the proposal associated with the ipsec policy.",
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"pfs_group": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "One of [1, 2, 5, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34]",
				ValidateFunc: validation.IntInSlice([]int{1, 2, 5, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34}),
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the ipsec policy.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the ipsec policy.",
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

const iPSecPolicyAPIPath = "vpn/ipsec-policies/"

type iPSecPolicyRequest struct {
	Name         string             `json:"name"`
	Description  string             `json:"description,omitempty"`
	Proposals    []int64            `json:"proposals"`
	PfsGroup     *int64             `json:"pfs_group,omitempty"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
}

type iPSecPolicyResponse struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Proposals    []netboxapi.Ref        `json:"proposals"`
	PfsGroup     *netboxapi.IntChoice   `json:"pfs_group"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	Tags         []netboxapi.TagRef     `json:"tags"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func iPSecPolicyRequestFromResourceData(api *providerState, d *schema.ResourceData) (*iPSecPolicyRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &iPSecPolicyRequest{
		Name:        d.Get("name").(string),
		Description: getOptionalStr(d, "description", true),
		Proposals:   getIDSet(d, "proposal_ids"),
		PfsGroup:    getOptionalInt(d, "pfs_group"),
		Owner:       getOptionalInt(d, "owner_id"),
		Comments:    getOptionalStr(d, "comments", true),
		Tags:        tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxIPSecPolicyCreate(d *schema.ResourceData, m interface{}) error {
	data, err := iPSecPolicyRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, iPSecPolicyAPIPath, data, resourceNetboxIPSecPolicyRead)
}

func resourceNetboxIPSecPolicyRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, iPSecPolicyAPIPath, func(r *iPSecPolicyResponse) {
		d.Set("name", r.Name)
		d.Set("description", r.Description)
		d.Set("proposal_ids", netboxapi.IDs(r.Proposals))
		d.Set("pfs_group", netboxapi.IntChoiceValue(r.PfsGroup))
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("comments", r.Comments)
		api.readTagRefs(d, r.Tags)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
	})
}

func resourceNetboxIPSecPolicyUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := iPSecPolicyRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, iPSecPolicyAPIPath, data, resourceNetboxIPSecPolicyRead)
}

func resourceNetboxIPSecPolicyDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, iPSecPolicyAPIPath)
}
