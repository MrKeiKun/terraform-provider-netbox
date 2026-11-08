package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var resourceNetboxCircuitStatusOptions = []string{"planned", "provisioning", "active", "offline", "deprovisioning", "decommissioning"}

func resourceNetboxCircuit() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxCircuitCreate,
		Read:   resourceNetboxCircuitRead,
		Update: resourceNetboxCircuitUpdate,
		Delete: resourceNetboxCircuitDelete,

		Description: `:meta:subcategory:Circuits:From the [official documentation](https://docs.netbox.dev/en/stable/features/circuits/#circuits_1):

> A communications circuit represents a single physical link connecting exactly two endpoints, commonly referred to as its A and Z terminations. A circuit in NetBox may have zero, one, or two terminations defined. It is common to have only one termination defined when you don't necessarily care about the details of the provider side of the circuit, e.g. for Internet access circuits. Both terminations would likely be modeled for circuits which connect one customer site to another.
>
> Each circuit is associated with a provider and a user-defined type. For example, you might have Internet access circuits delivered to each site by one provider, and private MPLS circuits delivered by another. Each circuit must be assigned a circuit ID, each of which must be unique per provider.`,

		Schema: map[string]*schema.Schema{
			"provider_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"cid": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"tenant_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"status": {
				Type:        schema.TypeString,
				Required:    true,
				Description: buildValidValueDescription(resourceNetboxCircuitStatusOptions),
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
			"provider_account_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"install_date": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"termination_date": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"commit_rate": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 2147483647),
				Description:  "Committed rate",
			},
			"distance": {
				Type:     schema.TypeFloat,
				Optional: true,
			},
			"distance_unit": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "One of [km, m, mi, ft]",
				ValidateFunc: validation.StringInSlice([]string{"km", "m", "mi", "ft"}, false),
			},
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const circuitAPIPath = "circuits/circuits/"

type circuitRequest struct {
	Cid             string             `json:"cid"`
	Status          string             `json:"status,omitempty"`
	Description     string             `json:"description,omitempty"`
	Provider        *int64             `json:"provider"`
	Type            *int64             `json:"type"`
	Tenant          *int64             `json:"tenant,omitempty"`
	Owner           *int64             `json:"owner"`
	Comments        string             `json:"comments,omitempty"`
	ProviderAccount *int64             `json:"provider_account"`
	InstallDate     string             `json:"install_date,omitempty"`
	TerminationDate string             `json:"termination_date,omitempty"`
	CommitRate      *int64             `json:"commit_rate"`
	Distance        *float64           `json:"distance"`
	DistanceUnit    *string            `json:"distance_unit,omitempty"`
	CustomFields    interface{}        `json:"custom_fields,omitempty"`
	Tags            []netboxapi.TagRef `json:"tags"`
}

type circuitResponse struct {
	Cid             string                 `json:"cid"`
	Status          *netboxapi.Choice      `json:"status"`
	Description     string                 `json:"description"`
	Provider        *netboxapi.Ref         `json:"provider"`
	Type            *netboxapi.Ref         `json:"type"`
	Tenant          *netboxapi.Ref         `json:"tenant"`
	Owner           *netboxapi.Ref         `json:"owner"`
	Comments        string                 `json:"comments"`
	ProviderAccount *netboxapi.Ref         `json:"provider_account"`
	InstallDate     string                 `json:"install_date"`
	TerminationDate string                 `json:"termination_date"`
	CommitRate      *int64                 `json:"commit_rate"`
	Distance        *float64               `json:"distance"`
	DistanceUnit    *netboxapi.Choice      `json:"distance_unit"`
	CustomFields    map[string]interface{} `json:"custom_fields"`
}

func circuitRequestFromResourceData(d *schema.ResourceData) *circuitRequest {
	data := &circuitRequest{
		CustomFields:    customFieldsRequest(d),
		DistanceUnit:    getOptionalStrClear(d, "distance_unit"),
		Distance:        getOptionalFloat(d, "distance"),
		CommitRate:      getOptionalInt(d, "commit_rate"),
		TerminationDate: getOptionalStr(d, "termination_date", true),
		InstallDate:     getOptionalStr(d, "install_date", true),
		ProviderAccount: getOptionalInt(d, "provider_account_id"),
		Comments:        getOptionalStr(d, "comments", true),
		Owner:           getOptionalInt(d, "owner_id"),
		Cid:             d.Get("cid").(string),
		Status:          d.Get("status").(string),
		Description:     d.Get("description").(string),
		Tags:            []netboxapi.TagRef{},
	}
	if v, ok := d.GetOk("provider_id"); ok {
		data.Provider = int64ToPtr(int64(v.(int)))
	}
	if v, ok := d.GetOk("type_id"); ok {
		data.Type = int64ToPtr(int64(v.(int)))
	}
	if v, ok := d.GetOk("tenant_id"); ok {
		data.Tenant = int64ToPtr(int64(v.(int)))
	}
	return data
}

func resourceNetboxCircuitCreate(d *schema.ResourceData, m interface{}) error {
	return createObject(d, m, circuitAPIPath, circuitRequestFromResourceData(d), resourceNetboxCircuitRead)
}

func resourceNetboxCircuitRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, circuitAPIPath, func(r *circuitResponse) {
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		d.Set("distance_unit", netboxapi.ChoiceValue(r.DistanceUnit))
		d.Set("distance", r.Distance)
		d.Set("commit_rate", r.CommitRate)
		d.Set("termination_date", r.TerminationDate)
		d.Set("install_date", r.InstallDate)
		d.Set("provider_account_id", netboxapi.RefID(r.ProviderAccount))
		d.Set("comments", r.Comments)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("cid", r.Cid)
		d.Set("status", netboxapi.ChoiceValue(r.Status))
		d.Set("description", r.Description)
		d.Set("provider_id", netboxapi.RefID(r.Provider))
		d.Set("type_id", netboxapi.RefID(r.Type))
		d.Set("tenant_id", netboxapi.RefID(r.Tenant))
	})
}

func resourceNetboxCircuitUpdate(d *schema.ResourceData, m interface{}) error {
	return patchObject(d, m, circuitAPIPath, circuitRequestFromResourceData(d), resourceNetboxCircuitRead)
}

func resourceNetboxCircuitDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, circuitAPIPath)
}
