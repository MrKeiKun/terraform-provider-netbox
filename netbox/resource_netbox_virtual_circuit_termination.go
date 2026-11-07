package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxVirtualCircuitTermination() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxVirtualCircuitTerminationCreate,
		Read:   resourceNetboxVirtualCircuitTerminationRead,
		Update: resourceNetboxVirtualCircuitTerminationUpdate,
		Delete: resourceNetboxVirtualCircuitTerminationDelete,

		Description: `:meta:subcategory:Circuits:From the [official documentation](https://docs.netbox.dev/en/stable/models/circuits/virtualcircuittermination/):

> A virtual circuit termination connects an interface to a virtual circuit.`,

		Schema: map[string]*schema.Schema{
			"virtual_circuit_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"role": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "One of [peer, hub, spoke]",
				ValidateFunc: validation.StringInSlice([]string{"peer", "hub", "spoke"}, false),
			},
			"interface_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"description": {
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

const virtualCircuitTerminationAPIPath = "circuits/virtual-circuit-terminations/"

type virtualCircuitTerminationRequest struct {
	VirtualCircuit *int64             `json:"virtual_circuit"`
	Role           string             `json:"role,omitempty"`
	Interface      *int64             `json:"interface"`
	Description    string             `json:"description,omitempty"`
	Tags           []netboxapi.TagRef `json:"tags"`
	CustomFields   interface{}        `json:"custom_fields,omitempty"`
}

type virtualCircuitTerminationResponse struct {
	VirtualCircuit *netboxapi.Ref         `json:"virtual_circuit"`
	Role           *netboxapi.Choice      `json:"role"`
	Interface      *netboxapi.Ref         `json:"interface"`
	Description    string                 `json:"description"`
	Tags           []netboxapi.TagRef     `json:"tags"`
	CustomFields   map[string]interface{} `json:"custom_fields"`
}

func virtualCircuitTerminationRequestFromResourceData(api *providerState, d *schema.ResourceData) (*virtualCircuitTerminationRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &virtualCircuitTerminationRequest{
		VirtualCircuit: int64ToPtr(int64(d.Get("virtual_circuit_id").(int))),
		Role:           d.Get("role").(string),
		Interface:      int64ToPtr(int64(d.Get("interface_id").(int))),
		Description:    getOptionalStr(d, "description", true),
		Tags:           tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxVirtualCircuitTerminationCreate(d *schema.ResourceData, m interface{}) error {
	data, err := virtualCircuitTerminationRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, virtualCircuitTerminationAPIPath, data, resourceNetboxVirtualCircuitTerminationRead)
}

func resourceNetboxVirtualCircuitTerminationRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, virtualCircuitTerminationAPIPath, func(r *virtualCircuitTerminationResponse) {
		d.Set("virtual_circuit_id", netboxapi.RefID(r.VirtualCircuit))
		d.Set("role", netboxapi.ChoiceValue(r.Role))
		d.Set("interface_id", netboxapi.RefID(r.Interface))
		d.Set("description", r.Description)
		api.readTagRefs(d, r.Tags)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
	})
}

func resourceNetboxVirtualCircuitTerminationUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := virtualCircuitTerminationRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, virtualCircuitTerminationAPIPath, data, resourceNetboxVirtualCircuitTerminationRead)
}

func resourceNetboxVirtualCircuitTerminationDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, virtualCircuitTerminationAPIPath)
}
