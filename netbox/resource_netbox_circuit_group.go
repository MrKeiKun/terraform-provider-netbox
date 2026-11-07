package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxCircuitGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxCircuitGroupCreate,
		Read:   resourceNetboxCircuitGroupRead,
		Update: resourceNetboxCircuitGroupUpdate,
		Delete: resourceNetboxCircuitGroupDelete,

		Description: `:meta:subcategory:Circuits:From the [official documentation](https://docs.netbox.dev/en/stable/models/circuits/circuitgroup/):

> Circuits can be grouped, for example to track the circuits that make up a redundant service.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"slug": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tenant_id": {
				Type:     schema.TypeInt,
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

const circuitGroupAPIPath = "circuits/circuit-groups/"

type circuitGroupRequest struct {
	Name         string             `json:"name"`
	Slug         string             `json:"slug"`
	Description  string             `json:"description,omitempty"`
	Tenant       *int64             `json:"tenant"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
}

type circuitGroupResponse struct {
	Name         string                 `json:"name"`
	Slug         string                 `json:"slug"`
	Description  string                 `json:"description"`
	Tenant       *netboxapi.Ref         `json:"tenant"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	Tags         []netboxapi.TagRef     `json:"tags"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func circuitGroupRequestFromResourceData(api *providerState, d *schema.ResourceData) (*circuitGroupRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &circuitGroupRequest{
		Name:        d.Get("name").(string),
		Slug:        slugFromResourceData(d, name),
		Description: getOptionalStr(d, "description", true),
		Tenant:      getOptionalInt(d, "tenant_id"),
		Owner:       getOptionalInt(d, "owner_id"),
		Comments:    getOptionalStr(d, "comments", true),
		Tags:        tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxCircuitGroupCreate(d *schema.ResourceData, m interface{}) error {
	data, err := circuitGroupRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, circuitGroupAPIPath, data, resourceNetboxCircuitGroupRead)
}

func resourceNetboxCircuitGroupRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, circuitGroupAPIPath, func(r *circuitGroupResponse) {
		d.Set("name", r.Name)
		d.Set("slug", r.Slug)
		d.Set("description", r.Description)
		d.Set("tenant_id", netboxapi.RefID(r.Tenant))
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("comments", r.Comments)
		api.readTagRefs(d, r.Tags)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
	})
}

func resourceNetboxCircuitGroupUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := circuitGroupRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, circuitGroupAPIPath, data, resourceNetboxCircuitGroupRead)
}

func resourceNetboxCircuitGroupDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, circuitGroupAPIPath)
}
