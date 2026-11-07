package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxIPSecProposal() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxIPSecProposalCreate,
		Read:   resourceNetboxIPSecProposalRead,
		Update: resourceNetboxIPSecProposalUpdate,
		Delete: resourceNetboxIPSecProposalDelete,

		Description: `:meta:subcategory:VPN Tunnels:From the [official documentation](https://docs.netbox.dev/en/stable/models/vpn/ipsecproposal/):

> An IPsec proposal defines the algorithms that are used in the second phase of an IPsec negotiation.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"encryption_algorithm": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "One of [aes-128-cbc, aes-128-gcm, aes-192-cbc, aes-192-gcm, aes-256-cbc, aes-256-gcm, 3des-cbc, des-cbc]",
				ValidateFunc: validation.StringInSlice([]string{"aes-128-cbc", "aes-128-gcm", "aes-192-cbc", "aes-192-gcm", "aes-256-cbc", "aes-256-gcm", "3des-cbc", "des-cbc"}, false),
			},
			"authentication_algorithm": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "One of [hmac-sha1, hmac-sha256, hmac-sha384, hmac-sha512, hmac-md5]",
				ValidateFunc: validation.StringInSlice([]string{"hmac-sha1", "hmac-sha256", "hmac-sha384", "hmac-sha512", "hmac-md5"}, false),
			},
			"sa_lifetime_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 2147483647),
				Description:  "Security association lifetime (seconds)",
			},
			"sa_lifetime_data": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 2147483647),
				Description:  "Security association lifetime (in kilobytes)",
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

const iPSecProposalAPIPath = "vpn/ipsec-proposals/"

type iPSecProposalRequest struct {
	Name                    string             `json:"name"`
	Description             string             `json:"description,omitempty"`
	EncryptionAlgorithm     *string            `json:"encryption_algorithm,omitempty"`
	AuthenticationAlgorithm *string            `json:"authentication_algorithm,omitempty"`
	SaLifetimeSeconds       *int64             `json:"sa_lifetime_seconds"`
	SaLifetimeData          *int64             `json:"sa_lifetime_data"`
	Owner                   *int64             `json:"owner"`
	Comments                string             `json:"comments,omitempty"`
	Tags                    []netboxapi.TagRef `json:"tags"`
	CustomFields            interface{}        `json:"custom_fields,omitempty"`
}

type iPSecProposalResponse struct {
	Name                    string                 `json:"name"`
	Description             string                 `json:"description"`
	EncryptionAlgorithm     *netboxapi.Choice      `json:"encryption_algorithm"`
	AuthenticationAlgorithm *netboxapi.Choice      `json:"authentication_algorithm"`
	SaLifetimeSeconds       *int64                 `json:"sa_lifetime_seconds"`
	SaLifetimeData          *int64                 `json:"sa_lifetime_data"`
	Owner                   *netboxapi.Ref         `json:"owner"`
	Comments                string                 `json:"comments"`
	Tags                    []netboxapi.TagRef     `json:"tags"`
	CustomFields            map[string]interface{} `json:"custom_fields"`
}

func iPSecProposalRequestFromResourceData(api *providerState, d *schema.ResourceData) (*iPSecProposalRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &iPSecProposalRequest{
		Name:                    d.Get("name").(string),
		Description:             getOptionalStr(d, "description", true),
		EncryptionAlgorithm:     getOptionalStrClear(d, "encryption_algorithm"),
		AuthenticationAlgorithm: getOptionalStrClear(d, "authentication_algorithm"),
		SaLifetimeSeconds:       getOptionalInt(d, "sa_lifetime_seconds"),
		SaLifetimeData:          getOptionalInt(d, "sa_lifetime_data"),
		Owner:                   getOptionalInt(d, "owner_id"),
		Comments:                getOptionalStr(d, "comments", true),
		Tags:                    tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxIPSecProposalCreate(d *schema.ResourceData, m interface{}) error {
	data, err := iPSecProposalRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, iPSecProposalAPIPath, data, resourceNetboxIPSecProposalRead)
}

func resourceNetboxIPSecProposalRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, iPSecProposalAPIPath, func(r *iPSecProposalResponse) {
		d.Set("name", r.Name)
		d.Set("description", r.Description)
		d.Set("encryption_algorithm", netboxapi.ChoiceValue(r.EncryptionAlgorithm))
		d.Set("authentication_algorithm", netboxapi.ChoiceValue(r.AuthenticationAlgorithm))
		d.Set("sa_lifetime_seconds", r.SaLifetimeSeconds)
		d.Set("sa_lifetime_data", r.SaLifetimeData)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("comments", r.Comments)
		api.readTagRefs(d, r.Tags)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
	})
}

func resourceNetboxIPSecProposalUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := iPSecProposalRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, iPSecProposalAPIPath, data, resourceNetboxIPSecProposalRead)
}

func resourceNetboxIPSecProposalDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, iPSecProposalAPIPath)
}
