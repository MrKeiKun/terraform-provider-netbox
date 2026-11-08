package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxCircuitGroupAssignment() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxCircuitGroupAssignmentCreate,
		Read:   resourceNetboxCircuitGroupAssignmentRead,
		Update: resourceNetboxCircuitGroupAssignmentUpdate,
		Delete: resourceNetboxCircuitGroupAssignmentDelete,

		Description: `:meta:subcategory:Circuits:From the [official documentation](https://docs.netbox.dev/en/stable/models/circuits/circuitgroupassignment/):

> A circuit group assignment places a circuit, or a virtual circuit, in a circuit group.`,

		Schema: map[string]*schema.Schema{
			"group_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the group associated with the circuit group assignment.",
				Required:    true,
			},
			"member_type": {
				Type:        schema.TypeString,
				Description: "The member type of the circuit group assignment.",
				Required:    true,
			},
			"member_id": {
				Type:         schema.TypeInt,
				Description:  "The ID of the member associated with the circuit group assignment.",
				Required:     true,
				ValidateFunc: validation.IntBetween(0, 9223372036854775807),
			},
			"priority": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "One of [primary, secondary, tertiary, inactive]",
				ValidateFunc: validation.StringInSlice([]string{"primary", "secondary", "tertiary", "inactive"}, false),
			},
			tagsKey:         tagsSchema,
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const circuitGroupAssignmentAPIPath = "circuits/circuit-group-assignments/"

type circuitGroupAssignmentRequest struct {
	Group        *int64             `json:"group"`
	MemberType   string             `json:"member_type"`
	MemberID     int64              `json:"member_id"`
	Priority     *string            `json:"priority,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
}

type circuitGroupAssignmentResponse struct {
	Group        *netboxapi.Ref         `json:"group"`
	MemberType   string                 `json:"member_type"`
	MemberID     *int64                 `json:"member_id"`
	Priority     *netboxapi.Choice      `json:"priority"`
	Tags         []netboxapi.TagRef     `json:"tags"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func circuitGroupAssignmentRequestFromResourceData(api *providerState, d *schema.ResourceData) (*circuitGroupAssignmentRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &circuitGroupAssignmentRequest{
		Group:      int64ToPtr(int64(d.Get("group_id").(int))),
		MemberType: d.Get("member_type").(string),
		MemberID:   int64(d.Get("member_id").(int)),
		Priority:   getOptionalStrClear(d, "priority"),
		Tags:       tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxCircuitGroupAssignmentCreate(d *schema.ResourceData, m interface{}) error {
	data, err := circuitGroupAssignmentRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, circuitGroupAssignmentAPIPath, data, resourceNetboxCircuitGroupAssignmentRead)
}

func resourceNetboxCircuitGroupAssignmentRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, circuitGroupAssignmentAPIPath, func(r *circuitGroupAssignmentResponse) {
		d.Set("group_id", netboxapi.RefID(r.Group))
		d.Set("member_type", r.MemberType)
		d.Set("member_id", r.MemberID)
		d.Set("priority", netboxapi.ChoiceValue(r.Priority))
		api.readTagRefs(d, r.Tags)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
	})
}

func resourceNetboxCircuitGroupAssignmentUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := circuitGroupAssignmentRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, circuitGroupAssignmentAPIPath, data, resourceNetboxCircuitGroupAssignmentRead)
}

func resourceNetboxCircuitGroupAssignmentDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, circuitGroupAssignmentAPIPath)
}
