package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxCircuitProviderNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxCircuitProviderNetworkCreate,
		Read:   resourceNetboxCircuitProviderNetworkRead,
		Update: resourceNetboxCircuitProviderNetworkUpdate,
		Delete: resourceNetboxCircuitProviderNetworkDelete,

		Description: `:meta:subcategory:Circuits:From the [offical documentation](https://netboxlabs.com/docs/netbox/models/circuits/providernetwork/):

> This model can be used to represent the boundary of a provider network, the details of which are unknown or unimportant to the NetBox user. For example, it might represent a provider's regional MPLS network to which multiple circuits provide connectivity.`,

		Schema: map[string]*schema.Schema{
			"comments": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"provider_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"service_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"owner_id": {
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

const circuitProviderNetworkAPIPath = "circuits/provider-networks/"

type circuitProviderNetworkRequest struct {
	Name         string             `json:"name"`
	Comments     string             `json:"comments,omitempty"`
	Description  string             `json:"description,omitempty"`
	Provider     *int64             `json:"provider"`
	ServiceId    string             `json:"service_id,omitempty"`
	Owner        *int64             `json:"owner"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
}

type circuitProviderNetworkResponse struct {
	Name         string                 `json:"name"`
	Comments     string                 `json:"comments"`
	Description  string                 `json:"description"`
	Provider     *netboxapi.Ref         `json:"provider"`
	ServiceId    string                 `json:"service_id"`
	Owner        *netboxapi.Ref         `json:"owner"`
	CustomFields map[string]interface{} `json:"custom_fields"`
	Tags         []netboxapi.TagRef     `json:"tags"`
}

func circuitProviderNetworkRequestFromResourceData(api *providerState, d *schema.ResourceData) (*circuitProviderNetworkRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &circuitProviderNetworkRequest{
		Owner:       getOptionalInt(d, "owner_id"),
		Name:        name,
		Comments:    d.Get("comments").(string),
		Description: d.Get("description").(string),
		Provider:    int64ToPtr(int64(d.Get("provider_id").(int))),
		ServiceId:   d.Get("service_id").(string),
		Tags:        tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxCircuitProviderNetworkCreate(d *schema.ResourceData, m interface{}) error {
	data, err := circuitProviderNetworkRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, circuitProviderNetworkAPIPath, data, resourceNetboxCircuitProviderNetworkRead)
}

func resourceNetboxCircuitProviderNetworkRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, circuitProviderNetworkAPIPath, func(r *circuitProviderNetworkResponse) {
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("name", r.Name)
		d.Set("comments", r.Comments)
		d.Set("description", r.Description)
		d.Set("provider_id", netboxapi.RefID(r.Provider))
		d.Set("service_id", r.ServiceId)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxCircuitProviderNetworkUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := circuitProviderNetworkRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, circuitProviderNetworkAPIPath, data, resourceNetboxCircuitProviderNetworkRead)
}

func resourceNetboxCircuitProviderNetworkDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, circuitProviderNetworkAPIPath)
}
