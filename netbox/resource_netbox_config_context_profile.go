package netbox

import (
	"encoding/json"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxConfigContextProfile() *schema.Resource {
	return &schema.Resource{
		CreateContext: withContext(resourceNetboxConfigContextProfileCreate),
		ReadContext:   withContext(resourceNetboxConfigContextProfileRead),
		UpdateContext: withContext(resourceNetboxConfigContextProfileUpdate),
		DeleteContext: withContext(resourceNetboxConfigContextProfileDelete),

		Description: `:meta:subcategory:Extras:From the [official documentation](https://docs.netbox.dev/en/stable/models/customization/configcontext/):

> A config context profile holds a JSON schema that config contexts can be validated against.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the config context profile.",
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the config context profile.",
				Optional:    true,
			},
			"schema": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentJSON,
				Description:      "A JSON schema specifying the structure of the context data for this profile",
			},
			tagsKey: tagsSchema,
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the config context profile.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the config context profile.",
				Optional:    true,
			},
			"data_source_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the data source associated with the config context profile.",
				Optional:    true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const configContextProfileAPIPath = "extras/config-context-profiles/"

type configContextProfileRequest struct {
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Schema      *json.RawMessage   `json:"schema"`
	Tags        []netboxapi.TagRef `json:"tags"`
	Owner       *int64             `json:"owner"`
	Comments    string             `json:"comments,omitempty"`
	DataSource  *int64             `json:"data_source,omitempty"`
}

type configContextProfileResponse struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Schema      json.RawMessage    `json:"schema"`
	Tags        []netboxapi.TagRef `json:"tags"`
	Owner       *netboxapi.Ref     `json:"owner"`
	Comments    string             `json:"comments"`
	DataSource  *netboxapi.Ref     `json:"data_source"`
}

func configContextProfileRequestFromResourceData(api *providerState, d *schema.ResourceData) (*configContextProfileRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	schemaRaw, err := jsonRawFromResourceData(d, "schema")
	if err != nil {
		return nil, err
	}
	data := &configContextProfileRequest{
		Name:        d.Get("name").(string),
		Description: getOptionalStr(d, "description", true),
		Schema:      schemaRaw,
		Tags:        tags,
		Owner:       getOptionalInt(d, "owner_id"),
		Comments:    getOptionalStr(d, "comments", true),
		DataSource:  getOptionalInt(d, "data_source_id"),
	}
	return data, nil
}

func resourceNetboxConfigContextProfileCreate(d *schema.ResourceData, m interface{}) error {
	data, err := configContextProfileRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, configContextProfileAPIPath, data, resourceNetboxConfigContextProfileRead)
}

func resourceNetboxConfigContextProfileRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, configContextProfileAPIPath, func(r *configContextProfileResponse) {
		d.Set("name", r.Name)
		d.Set("description", r.Description)
		d.Set("schema", jsonString(r.Schema))
		api.readTagRefs(d, r.Tags)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("comments", r.Comments)
		d.Set("data_source_id", netboxapi.RefID(r.DataSource))
	})
}

func resourceNetboxConfigContextProfileUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := configContextProfileRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, configContextProfileAPIPath, data, resourceNetboxConfigContextProfileRead)
}

func resourceNetboxConfigContextProfileDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, configContextProfileAPIPath)
}
