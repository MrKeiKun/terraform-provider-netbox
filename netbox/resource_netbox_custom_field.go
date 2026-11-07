package netbox

import (
	"encoding/json"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// customFieldDefaultToAPI - converts the tf attribute to the json netbox expects
func customFieldDefaultToAPI(d *schema.ResourceData) interface{} {
	v, ok := d.GetOk("default")
	if !ok {
		return nil
	}
	s := v.(string)
	var parsed interface{}
	if err := json.Unmarshal([]byte(s), &parsed); err != nil {
		return s
	}
	return parsed
}

// normalizeCustomFieldDefault - state/config always compare as equal,
// strings left the same
func normalizeCustomFieldDefault(v interface{}) string {
	s := v.(string)
	var parsed interface{}
	if json.Unmarshal([]byte(s), &parsed) != nil {
		return s
	}
	b, err := json.Marshal(parsed)
	if err != nil {
		return s
	}
	return string(b)
}

func resourceCustomField() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxCustomFieldCreate,
		Read:   resourceNetboxCustomFieldRead,
		Update: resourceNetboxCustomFieldUpdate,
		Delete: resourceNetboxCustomFieldDelete,

		Description: `:meta:subcategory:Extras:From the [official documentation](https://docs.netbox.dev/en/stable/customization/custom-fields/#custom-fields):

> Each model in NetBox is represented in the database as a discrete table, and each attribute of a model exists as a column within its table. For example, sites are stored in the dcim_site table, which has columns named name, facility, physical_address, and so on. As new attributes are added to objects throughout the development of NetBox, tables are expanded to include new rows.
>
> However, some users might want to store additional object attributes that are somewhat esoteric in nature, and that would not make sense to include in the core NetBox database schema. For instance, suppose your organization needs to associate each device with a ticket number correlating it with an internal support system record. This is certainly a legitimate use for NetBox, but it's not a common enough need to warrant including a field for every NetBox installation. Instead, you can create a custom field to hold this data.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"text",
					"integer",
					"boolean",
					"date",
					"url",
					"select",
					"multiselect",
					"object",
					"multiobject",
					"json",
				}, false),
			},
			"content_types": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set: schema.HashString,
			},
			"weight": {
				Type:     schema.TypeInt,
				Required: true,
				DefaultFunc: func() (interface{}, error) {
					return 100, nil
				},
			},
			"default": {
				Type:      schema.TypeString,
				Optional:  true,
				StateFunc: normalizeCustomFieldDefault,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"group_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"label": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"required": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"validation_maximum": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"validation_minimum": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"validation_regex": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"choice_set_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"related_object_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const customFieldAPIPath = "extras/custom-fields/"

type customFieldRequest struct {
	Name              string      `json:"name"`
	Type              string      `json:"type,omitempty"`
	Default           interface{} `json:"default,omitempty"`
	Description       string      `json:"description,omitempty"`
	GroupName         string      `json:"group_name,omitempty"`
	Label             string      `json:"label,omitempty"`
	Required          bool        `json:"required"`
	ValidationRegex   string      `json:"validation_regex,omitempty"`
	RelatedObjectType string      `json:"related_object_type,omitempty"`
	Weight            *int64      `json:"weight,omitempty"`
	ChoiceSet         *int64      `json:"choice_set,omitempty"`
	ObjectTypes       []string    `json:"object_types,omitempty"`
	ValidationMaximum *float64    `json:"validation_maximum,omitempty"`
	ValidationMinimum *float64    `json:"validation_minimum,omitempty"`
}

type customFieldResponse struct {
	Name              string            `json:"name"`
	Type              *netboxapi.Choice `json:"type"`
	ObjectTypes       []string          `json:"object_types"`
	ChoiceSet         *netboxapi.Ref    `json:"choice_set"`
	Weight            *int64            `json:"weight"`
	Default           interface{}       `json:"default"`
	Description       string            `json:"description"`
	GroupName         string            `json:"group_name"`
	Label             string            `json:"label"`
	Required          bool              `json:"required"`
	ValidationMaximum *float64          `json:"validation_maximum"`
	ValidationMinimum *float64          `json:"validation_minimum"`
	ValidationRegex   string            `json:"validation_regex"`
}

func customFieldRequestFromResourceData(d *schema.ResourceData) *customFieldRequest {
	data := &customFieldRequest{
		Name:              d.Get("name").(string),
		Type:              d.Get("type").(string),
		Default:           customFieldDefaultToAPI(d),
		Description:       d.Get("description").(string),
		GroupName:         d.Get("group_name").(string),
		Label:             d.Get("label").(string),
		Required:          d.Get("required").(bool),
		ValidationRegex:   d.Get("validation_regex").(string),
		RelatedObjectType: d.Get("related_object_type").(string),
		Weight:            int64ToPtr(int64(d.Get("weight").(int))),
	}

	if choiceSet, ok := d.GetOk("choice_set_id"); ok {
		data.ChoiceSet = int64ToPtr(int64(choiceSet.(int)))
	}
	if ctypes, ok := d.GetOk("content_types"); ok {
		for _, t := range ctypes.(*schema.Set).List() {
			data.ObjectTypes = append(data.ObjectTypes, t.(string))
		}
	}
	if vmax, ok := d.GetOk("validation_maximum"); ok {
		data.ValidationMaximum = float64ToPtr(float64(vmax.(int)))
	}
	if vmin, ok := d.GetOk("validation_minimum"); ok {
		data.ValidationMinimum = float64ToPtr(float64(vmin.(int)))
	}
	return data
}

func resourceNetboxCustomFieldCreate(d *schema.ResourceData, m interface{}) error {
	return createObject(d, m, customFieldAPIPath, customFieldRequestFromResourceData(d), resourceNetboxCustomFieldRead)
}

func resourceNetboxCustomFieldRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, customFieldAPIPath, func(customField *customFieldResponse) {
		d.Set("name", customField.Name)
		d.Set("type", netboxapi.ChoiceValue(customField.Type))
		d.Set("content_types", customField.ObjectTypes)
		if customField.ChoiceSet != nil {
			d.Set("choice_set_id", customField.ChoiceSet.ID)
		}
		d.Set("weight", customField.Weight)

		// JSON value (number, bool, list) is re-encoded to its JSON text.
		if customField.Default != nil {
			if s, isStr := customField.Default.(string); isStr {
				d.Set("default", s)
			} else if b, err := json.Marshal(customField.Default); err == nil {
				d.Set("default", string(b))
			}
		} else {
			d.Set("default", "")
		}

		d.Set("description", customField.Description)
		d.Set("group_name", customField.GroupName)
		d.Set("label", customField.Label)
		d.Set("required", customField.Required)
		d.Set("validation_maximum", customField.ValidationMaximum)
		d.Set("validation_minimum", customField.ValidationMinimum)
		d.Set("validation_regex", customField.ValidationRegex)
	})
}

func resourceNetboxCustomFieldUpdate(d *schema.ResourceData, m interface{}) error {
	return putObject(d, m, customFieldAPIPath, customFieldRequestFromResourceData(d), resourceNetboxCustomFieldRead)
}

func resourceNetboxCustomFieldDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, customFieldAPIPath)
}
