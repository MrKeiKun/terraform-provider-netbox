package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxIKEProposal() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxIKEProposalCreate,
		Read:   resourceNetboxIKEProposalRead,
		Update: resourceNetboxIKEProposalUpdate,
		Delete: resourceNetboxIKEProposalDelete,

		Description: `:meta:subcategory:VPN Tunnels:From the [official documentation](https://docs.netbox.dev/en/stable/models/vpn/ikeproposal/):

> An IKE proposal defines the algorithms that are used in the first phase of an IPsec negotiation.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the ike proposal.",
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the ike proposal.",
				Optional:    true,
			},
			"authentication_method": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "One of [preshared-keys, certificates, rsa-signatures, dsa-signatures]",
				ValidateFunc: validation.StringInSlice([]string{"preshared-keys", "certificates", "rsa-signatures", "dsa-signatures"}, false),
			},
			"encryption_algorithm": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "One of [aes-128-cbc, aes-128-gcm, aes-192-cbc, aes-192-gcm, aes-256-cbc, aes-256-gcm, 3des-cbc, des-cbc]",
				ValidateFunc: validation.StringInSlice([]string{"aes-128-cbc", "aes-128-gcm", "aes-192-cbc", "aes-192-gcm", "aes-256-cbc", "aes-256-gcm", "3des-cbc", "des-cbc"}, false),
			},
			"authentication_algorithm": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "One of [hmac-sha1, hmac-sha256, hmac-sha384, hmac-sha512, hmac-md5]",
				ValidateFunc: validation.StringInSlice([]string{"hmac-sha1", "hmac-sha256", "hmac-sha384", "hmac-sha512", "hmac-md5"}, false),
			},
			"group": {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "One of [1, 2, 5, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34]",
				ValidateFunc: validation.IntInSlice([]int{1, 2, 5, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34}),
			},
			"sa_lifetime": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 2147483647),
				Description:  "Security association lifetime (in seconds)",
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the ike proposal.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the ike proposal.",
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

const iKEProposalAPIPath = "vpn/ike-proposals/"

type iKEProposalRequest struct {
	Name                    string             `json:"name"`
	Description             string             `json:"description,omitempty"`
	AuthenticationMethod    string             `json:"authentication_method,omitempty"`
	EncryptionAlgorithm     string             `json:"encryption_algorithm,omitempty"`
	AuthenticationAlgorithm *string            `json:"authentication_algorithm,omitempty"`
	Group                   int64              `json:"group"`
	SaLifetime              *int64             `json:"sa_lifetime"`
	Owner                   *int64             `json:"owner"`
	Comments                string             `json:"comments,omitempty"`
	Tags                    []netboxapi.TagRef `json:"tags"`
	CustomFields            interface{}        `json:"custom_fields,omitempty"`
}

type iKEProposalResponse struct {
	Name                    string                 `json:"name"`
	Description             string                 `json:"description"`
	AuthenticationMethod    *netboxapi.Choice      `json:"authentication_method"`
	EncryptionAlgorithm     *netboxapi.Choice      `json:"encryption_algorithm"`
	AuthenticationAlgorithm *netboxapi.Choice      `json:"authentication_algorithm"`
	Group                   *netboxapi.IntChoice   `json:"group"`
	SaLifetime              *int64                 `json:"sa_lifetime"`
	Owner                   *netboxapi.Ref         `json:"owner"`
	Comments                string                 `json:"comments"`
	Tags                    []netboxapi.TagRef     `json:"tags"`
	CustomFields            map[string]interface{} `json:"custom_fields"`
}

func iKEProposalRequestFromResourceData(api *providerState, d *schema.ResourceData) (*iKEProposalRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &iKEProposalRequest{
		Name:                    d.Get("name").(string),
		Description:             getOptionalStr(d, "description", true),
		AuthenticationMethod:    d.Get("authentication_method").(string),
		EncryptionAlgorithm:     d.Get("encryption_algorithm").(string),
		AuthenticationAlgorithm: getOptionalStrClear(d, "authentication_algorithm"),
		Group:                   int64(d.Get("group").(int)),
		SaLifetime:              getOptionalInt(d, "sa_lifetime"),
		Owner:                   getOptionalInt(d, "owner_id"),
		Comments:                getOptionalStr(d, "comments", true),
		Tags:                    tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxIKEProposalCreate(d *schema.ResourceData, m interface{}) error {
	data, err := iKEProposalRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, iKEProposalAPIPath, data, resourceNetboxIKEProposalRead)
}

func resourceNetboxIKEProposalRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, iKEProposalAPIPath, func(r *iKEProposalResponse) {
		d.Set("name", r.Name)
		d.Set("description", r.Description)
		d.Set("authentication_method", netboxapi.ChoiceValue(r.AuthenticationMethod))
		d.Set("encryption_algorithm", netboxapi.ChoiceValue(r.EncryptionAlgorithm))
		d.Set("authentication_algorithm", netboxapi.ChoiceValue(r.AuthenticationAlgorithm))
		d.Set("group", netboxapi.IntChoiceValue(r.Group))
		d.Set("sa_lifetime", r.SaLifetime)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("comments", r.Comments)
		api.readTagRefs(d, r.Tags)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
	})
}

func resourceNetboxIKEProposalUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := iKEProposalRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, iKEProposalAPIPath, data, resourceNetboxIKEProposalRead)
}

func resourceNetboxIKEProposalDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, iKEProposalAPIPath)
}
