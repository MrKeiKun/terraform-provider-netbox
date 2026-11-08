package netbox

import (
	"encoding/json"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxModuleTypeProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxModuleTypeProfileCreate,
		Read:   resourceNetboxModuleTypeProfileRead,
		Update: resourceNetboxModuleTypeProfileUpdate,
		Delete: resourceNetboxModuleTypeProfileDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/models/dcim/moduletypeprofile/):

> A module type profile defines the attributes that modules of a module type can have, with a JSON schema.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the module type profile.",
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the module type profile.",
				Optional:    true,
			},
			"schema": {
				Type:             schema.TypeString,
				Description:      "The schema of the module type profile.",
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentJSON,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the module type profile.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the module type profile.",
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

const moduleTypeProfileAPIPath = "dcim/module-type-profiles/"

type moduleTypeProfileRequest struct {
	Name         string             `json:"name"`
	Description  string             `json:"description,omitempty"`
	Schema       *json.RawMessage   `json:"schema"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
}

type moduleTypeProfileResponse struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Schema       json.RawMessage        `json:"schema"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	Tags         []netboxapi.TagRef     `json:"tags"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func moduleTypeProfileRequestFromResourceData(api *providerState, d *schema.ResourceData) (*moduleTypeProfileRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	schemaRaw, err := jsonRawFromResourceData(d, "schema")
	if err != nil {
		return nil, err
	}
	data := &moduleTypeProfileRequest{
		Name:        d.Get("name").(string),
		Description: getOptionalStr(d, "description", true),
		Schema:      schemaRaw,
		Owner:       getOptionalInt(d, "owner_id"),
		Comments:    getOptionalStr(d, "comments", true),
		Tags:        tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxModuleTypeProfileCreate(d *schema.ResourceData, m interface{}) error {
	data, err := moduleTypeProfileRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, moduleTypeProfileAPIPath, data, resourceNetboxModuleTypeProfileRead)
}

func resourceNetboxModuleTypeProfileRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, moduleTypeProfileAPIPath, func(r *moduleTypeProfileResponse) {
		d.Set("name", r.Name)
		d.Set("description", r.Description)
		d.Set("schema", jsonString(r.Schema))
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("comments", r.Comments)
		api.readTagRefs(d, r.Tags)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
	})
}

func resourceNetboxModuleTypeProfileUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := moduleTypeProfileRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, moduleTypeProfileAPIPath, data, resourceNetboxModuleTypeProfileRead)
}

func resourceNetboxModuleTypeProfileDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, moduleTypeProfileAPIPath)
}
