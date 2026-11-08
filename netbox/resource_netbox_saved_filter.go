package netbox

import (
	"encoding/json"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxSavedFilter() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxSavedFilterCreate,
		Read:   resourceNetboxSavedFilterRead,
		Update: resourceNetboxSavedFilterUpdate,
		Delete: resourceNetboxSavedFilterDelete,

		Description: `:meta:subcategory:Extras:From the [official documentation](https://docs.netbox.dev/en/stable/models/customization/savedfilter/):

> A saved filter stores a set of filter parameters for a list view, to be reused or shared.`,

		Schema: map[string]*schema.Schema{
			"object_types": {
				Type:        schema.TypeSet,
				Description: "The object types the saved filter applies to.",
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the saved filter.",
				Required:    true,
			},
			"slug": {
				Type:         schema.TypeString,
				Description:  "The URL-friendly slug of the saved filter.",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the saved filter.",
				Optional:    true,
			},
			"user_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the user associated with the saved filter.",
				Optional:    true,
			},
			"weight": {
				Type:         schema.TypeInt,
				Description:  "The weight of the saved filter.",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 32767),
			},
			"enabled": {
				Type:        schema.TypeBool,
				Description: "Whether the saved filter is enabled.",
				Optional:    true,
				Default:     false,
			},
			"shared": {
				Type:        schema.TypeBool,
				Description: "Whether shared is enabled for the saved filter.",
				Optional:    true,
				Default:     false,
			},
			"parameters": {
				Type:             schema.TypeString,
				Description:      "The parameters of the saved filter.",
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentJSON,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the saved filter.",
				Optional:    true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const savedFilterAPIPath = "extras/saved-filters/"

type savedFilterRequest struct {
	ObjectTypes []string         `json:"object_types"`
	Name        string           `json:"name"`
	Slug        string           `json:"slug"`
	Description string           `json:"description,omitempty"`
	User        *int64           `json:"user"`
	Weight      int64            `json:"weight,omitempty"`
	Enabled     bool             `json:"enabled"`
	Shared      bool             `json:"shared"`
	Parameters  *json.RawMessage `json:"parameters"`
	Owner       *int64           `json:"owner"`
}

type savedFilterResponse struct {
	ObjectTypes []string        `json:"object_types"`
	Name        string          `json:"name"`
	Slug        string          `json:"slug"`
	Description string          `json:"description"`
	User        *netboxapi.Ref  `json:"user"`
	Weight      *int64          `json:"weight"`
	Enabled     bool            `json:"enabled"`
	Shared      bool            `json:"shared"`
	Parameters  json.RawMessage `json:"parameters"`
	Owner       *netboxapi.Ref  `json:"owner"`
}

func savedFilterRequestFromResourceData(api *providerState, d *schema.ResourceData) (*savedFilterRequest, error) {
	name := d.Get("name").(string)
	parametersRaw, err := jsonRawFromResourceData(d, "parameters")
	if err != nil {
		return nil, err
	}
	data := &savedFilterRequest{
		ObjectTypes: getStringSet(d, "object_types"),
		Name:        d.Get("name").(string),
		Slug:        slugFromResourceData(d, name),
		Description: getOptionalStr(d, "description", true),
		User:        getOptionalInt(d, "user_id"),
		Weight:      int64(d.Get("weight").(int)),
		Enabled:     d.Get("enabled").(bool),
		Shared:      d.Get("shared").(bool),
		Parameters:  parametersRaw,
		Owner:       getOptionalInt(d, "owner_id"),
	}
	return data, nil
}

func resourceNetboxSavedFilterCreate(d *schema.ResourceData, m interface{}) error {
	data, err := savedFilterRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, savedFilterAPIPath, data, resourceNetboxSavedFilterRead)
}

func resourceNetboxSavedFilterRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, savedFilterAPIPath, func(r *savedFilterResponse) {
		d.Set("object_types", r.ObjectTypes)
		d.Set("name", r.Name)
		d.Set("slug", r.Slug)
		d.Set("description", r.Description)
		d.Set("user_id", netboxapi.RefID(r.User))
		d.Set("weight", r.Weight)
		d.Set("enabled", r.Enabled)
		d.Set("shared", r.Shared)
		d.Set("parameters", jsonString(r.Parameters))
		d.Set("owner_id", netboxapi.RefID(r.Owner))
	})
}

func resourceNetboxSavedFilterUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := savedFilterRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, savedFilterAPIPath, data, resourceNetboxSavedFilterRead)
}

func resourceNetboxSavedFilterDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, savedFilterAPIPath)
}
