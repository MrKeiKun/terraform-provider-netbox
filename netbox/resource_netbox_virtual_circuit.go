package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxVirtualCircuit() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxVirtualCircuitCreate,
		Read:   resourceNetboxVirtualCircuitRead,
		Update: resourceNetboxVirtualCircuitUpdate,
		Delete: resourceNetboxVirtualCircuitDelete,

		Description: `:meta:subcategory:Circuits:From the [official documentation](https://docs.netbox.dev/en/stable/models/circuits/virtualcircuit/):

> A virtual circuit is a logical connection between interfaces that rides on a provider network, for example an L2 service.`,

		Schema: map[string]*schema.Schema{
			"cid": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique circuit ID",
			},
			"provider_network_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"provider_account_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"type_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "One of [planned, provisioning, active, offline, deprovisioning, decommissioned]",
				ValidateFunc: validation.StringInSlice([]string{"planned", "provisioning", "active", "offline", "deprovisioning", "decommissioned"}, false),
			},
			"tenant_id": {
				Type:     schema.TypeInt,
				Optional: true,
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

const virtualCircuitAPIPath = "circuits/virtual-circuits/"

type virtualCircuitRequest struct {
	Cid             string             `json:"cid"`
	ProviderNetwork *int64             `json:"provider_network"`
	ProviderAccount *int64             `json:"provider_account"`
	Type            *int64             `json:"type"`
	Status          string             `json:"status,omitempty"`
	Tenant          *int64             `json:"tenant"`
	Description     string             `json:"description,omitempty"`
	Owner           *int64             `json:"owner"`
	Comments        string             `json:"comments,omitempty"`
	Tags            []netboxapi.TagRef `json:"tags"`
	CustomFields    interface{}        `json:"custom_fields,omitempty"`
}

type virtualCircuitResponse struct {
	Cid             string                 `json:"cid"`
	ProviderNetwork *netboxapi.Ref         `json:"provider_network"`
	ProviderAccount *netboxapi.Ref         `json:"provider_account"`
	Type            *netboxapi.Ref         `json:"type"`
	Status          *netboxapi.Choice      `json:"status"`
	Tenant          *netboxapi.Ref         `json:"tenant"`
	Description     string                 `json:"description"`
	Owner           *netboxapi.Ref         `json:"owner"`
	Comments        string                 `json:"comments"`
	Tags            []netboxapi.TagRef     `json:"tags"`
	CustomFields    map[string]interface{} `json:"custom_fields"`
}

func virtualCircuitRequestFromResourceData(api *providerState, d *schema.ResourceData) (*virtualCircuitRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &virtualCircuitRequest{
		Cid:             d.Get("cid").(string),
		ProviderNetwork: int64ToPtr(int64(d.Get("provider_network_id").(int))),
		ProviderAccount: getOptionalInt(d, "provider_account_id"),
		Type:            int64ToPtr(int64(d.Get("type_id").(int))),
		Status:          d.Get("status").(string),
		Tenant:          getOptionalInt(d, "tenant_id"),
		Description:     getOptionalStr(d, "description", true),
		Owner:           getOptionalInt(d, "owner_id"),
		Comments:        getOptionalStr(d, "comments", true),
		Tags:            tags,
	}
	if ct, ok := d.GetOk(customFieldsKey); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxVirtualCircuitCreate(d *schema.ResourceData, m interface{}) error {
	data, err := virtualCircuitRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, virtualCircuitAPIPath, data, resourceNetboxVirtualCircuitRead)
}

func resourceNetboxVirtualCircuitRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, virtualCircuitAPIPath, func(r *virtualCircuitResponse) {
		d.Set("cid", r.Cid)
		d.Set("provider_network_id", netboxapi.RefID(r.ProviderNetwork))
		d.Set("provider_account_id", netboxapi.RefID(r.ProviderAccount))
		d.Set("type_id", netboxapi.RefID(r.Type))
		d.Set("status", netboxapi.ChoiceValue(r.Status))
		d.Set("tenant_id", netboxapi.RefID(r.Tenant))
		d.Set("description", r.Description)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("comments", r.Comments)
		api.readTagRefs(d, r.Tags)
		if cf := getCustomFields(r.CustomFields); cf != nil {
			d.Set(customFieldsKey, cf)
		}
	})
}

func resourceNetboxVirtualCircuitUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := virtualCircuitRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, virtualCircuitAPIPath, data, resourceNetboxVirtualCircuitRead)
}

func resourceNetboxVirtualCircuitDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, virtualCircuitAPIPath)
}
