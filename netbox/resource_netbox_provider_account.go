package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxProviderAccount() *schema.Resource {
	return &schema.Resource{
		CreateContext: withContext(resourceNetboxProviderAccountCreate),
		ReadContext:   withContext(resourceNetboxProviderAccountRead),
		UpdateContext: withContext(resourceNetboxProviderAccountUpdate),
		DeleteContext: withContext(resourceNetboxProviderAccountDelete),

		Description: `:meta:subcategory:Circuits:From the [official documentation](https://docs.netbox.dev/en/stable/models/circuits/provideraccount/):

> A provider account is the account at a circuit provider, which the circuits of that provider are billed to.`,

		Schema: map[string]*schema.Schema{
			"provider_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the provider associated with the provider account.",
				Required:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the provider account.",
				Optional:    true,
			},
			"account": {
				Type:        schema.TypeString,
				Description: "The account of the provider account.",
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the provider account.",
				Optional:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the provider account.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the provider account.",
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

const providerAccountAPIPath = "circuits/provider-accounts/"

type providerAccountRequest struct {
	Provider     *int64             `json:"provider"`
	Name         string             `json:"name,omitempty"`
	Account      string             `json:"account"`
	Description  string             `json:"description,omitempty"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
}

type providerAccountResponse struct {
	Provider     *netboxapi.Ref         `json:"provider"`
	Name         string                 `json:"name"`
	Account      string                 `json:"account"`
	Description  string                 `json:"description"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	Tags         []netboxapi.TagRef     `json:"tags"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func providerAccountRequestFromResourceData(api *providerState, d *schema.ResourceData) (*providerAccountRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &providerAccountRequest{
		Provider:    int64ToPtr(int64(d.Get("provider_id").(int))),
		Name:        getOptionalStr(d, "name", true),
		Account:     d.Get("account").(string),
		Description: getOptionalStr(d, "description", true),
		Owner:       getOptionalInt(d, "owner_id"),
		Comments:    getOptionalStr(d, "comments", true),
		Tags:        tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxProviderAccountCreate(d *schema.ResourceData, m interface{}) error {
	data, err := providerAccountRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, providerAccountAPIPath, data, resourceNetboxProviderAccountRead)
}

func resourceNetboxProviderAccountRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, providerAccountAPIPath, func(r *providerAccountResponse) {
		d.Set("provider_id", netboxapi.RefID(r.Provider))
		d.Set("name", r.Name)
		d.Set("account", r.Account)
		d.Set("description", r.Description)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("comments", r.Comments)
		api.readTagRefs(d, r.Tags)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
	})
}

func resourceNetboxProviderAccountUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := providerAccountRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, providerAccountAPIPath, data, resourceNetboxProviderAccountRead)
}

func resourceNetboxProviderAccountDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, providerAccountAPIPath)
}
