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
		CreateContext: withContext(resourceNetboxCustomFieldCreate),
		ReadContext:   withContext(resourceNetboxCustomFieldRead),
		UpdateContext: withContext(resourceNetboxCustomFieldUpdate),
		DeleteContext: withContext(resourceNetboxCustomFieldDelete),

		Description: `:meta:subcategory:Extras:From the [official documentation](https://docs.netbox.dev/en/stable/customization/custom-fields/#custom-fields):

> Each model in NetBox is represented in the database as a discrete table, and each attribute of a model exists as a column within its table. For example, sites are stored in the dcim_site table, which has columns named name, facility, physical_address, and so on. As new attributes are added to objects throughout the development of NetBox, tables are expanded to include new rows.
>
> However, some users might want to store additional object attributes that are somewhat esoteric in nature, and that would not make sense to include in the core NetBox database schema. For instance, suppose your organization needs to associate each device with a ticket number correlating it with an internal support system record. This is certainly a legitimate use for NetBox, but it's not a common enough need to warrant including a field for every NetBox installation. Instead, you can create a custom field to hold this data.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the custom field.",
				Required:    true,
			},
			"type": {
				Type:        schema.TypeString,
				Description: "The type of the custom field.",
				Required:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"text",
					"longtext",
					"integer",
					"decimal",
					"boolean",
					"date",
					"datetime",
					"url",
					"select",
					"multiselect",
					"object",
					"multiobject",
					"json",
				}, false),
			},
			"content_types": {
				Type:        schema.TypeSet,
				Description: "The object types the custom field applies to.",
				Required:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set: schema.HashString,
			},
			"weight": {
				Type:        schema.TypeInt,
				Description: "The weight of the custom field.",
				Required:    true,
				DefaultFunc: func() (interface{}, error) {
					return 100, nil
				},
			},
			"default": {
				Type:        schema.TypeString,
				Description: "The default value of the custom field.",
				Optional:    true,
				StateFunc:   normalizeCustomFieldDefault,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the custom field.",
				Optional:    true,
			},
			"group_name": {
				Type:        schema.TypeString,
				Description: "The group the custom field is shown in.",
				Optional:    true,
			},
			"label": {
				Type:        schema.TypeString,
				Description: "The label of the custom field.",
				Optional:    true,
			},
			"required": {
				Type:        schema.TypeBool,
				Description: "Whether a value is required for the custom field.",
				Optional:    true,
			},
			"validation_maximum": {
				Type:        schema.TypeInt,
				Description: "The maximum allowed value.",
				Optional:    true,
			},
			"validation_minimum": {
				Type:        schema.TypeInt,
				Description: "The minimum allowed value.",
				Optional:    true,
			},
			"validation_regex": {
				Type:        schema.TypeString,
				Description: "A regular expression that values must match.",
				Optional:    true,
			},
			"choice_set_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the choice set associated with the custom field.",
				Optional:    true,
			},
			"related_object_type": {
				Type:        schema.TypeString,
				Description: "The related object type of the custom field.",
				Optional:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the custom field.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the custom field.",
				Optional:    true,
			},
			"unique": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "The value of this field must be unique for the assigned object",
			},
			"search_weight": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 32767),
				Description:  "Weighting for search. Lower values are considered more important. Fields with a search weight of zero will be ignored.",
			},
			"filter_logic": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "One of [disabled, loose, exact]",
				ValidateFunc: validation.StringInSlice([]string{"disabled", "loose", "exact"}, false),
			},
			"ui_visible": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "One of [always, if-set, hidden]",
				ValidateFunc: validation.StringInSlice([]string{"always", "if-set", "hidden"}, false),
			},
			"ui_editable": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "One of [yes, no, hidden]",
				ValidateFunc: validation.StringInSlice([]string{"yes", "no", "hidden"}, false),
			},
			"is_cloneable": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Replicate this value when cloning objects",
			},
			"nulls_first": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Sort null values before non-null values when ordering by this field",
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
	Owner             *int64      `json:"owner"`
	Comments          string      `json:"comments,omitempty"`
	Unique            *bool       `json:"unique,omitempty"`
	SearchWeight      int64       `json:"search_weight,omitempty"`
	FilterLogic       string      `json:"filter_logic,omitempty"`
	UiVisible         string      `json:"ui_visible,omitempty"`
	UiEditable        string      `json:"ui_editable,omitempty"`
	IsCloneable       *bool       `json:"is_cloneable,omitempty"`
	NullsFirst        *bool       `json:"nulls_first,omitempty"`
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
	Owner             *netboxapi.Ref    `json:"owner"`
	Comments          string            `json:"comments"`
	Unique            bool              `json:"unique"`
	SearchWeight      *int64            `json:"search_weight"`
	FilterLogic       *netboxapi.Choice `json:"filter_logic"`
	UiVisible         *netboxapi.Choice `json:"ui_visible"`
	UiEditable        *netboxapi.Choice `json:"ui_editable"`
	IsCloneable       bool              `json:"is_cloneable"`
	NullsFirst        bool              `json:"nulls_first"`
}

func customFieldRequestFromResourceData(d *schema.ResourceData) *customFieldRequest {
	data := &customFieldRequest{
		NullsFirst:        getOptionalBool(d, "nulls_first"),
		IsCloneable:       getOptionalBool(d, "is_cloneable"),
		UiEditable:        d.Get("ui_editable").(string),
		UiVisible:         d.Get("ui_visible").(string),
		FilterLogic:       d.Get("filter_logic").(string),
		SearchWeight:      int64(d.Get("search_weight").(int)),
		Unique:            getOptionalBool(d, "unique"),
		Comments:          getOptionalStr(d, "comments", true),
		Owner:             getOptionalInt(d, "owner_id"),
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
		d.Set("nulls_first", customField.NullsFirst)
		d.Set("is_cloneable", customField.IsCloneable)
		d.Set("ui_editable", netboxapi.ChoiceValue(customField.UiEditable))
		d.Set("ui_visible", netboxapi.ChoiceValue(customField.UiVisible))
		d.Set("filter_logic", netboxapi.ChoiceValue(customField.FilterLogic))
		d.Set("search_weight", customField.SearchWeight)
		d.Set("unique", customField.Unique)
		d.Set("comments", customField.Comments)
		d.Set("owner_id", netboxapi.RefID(customField.Owner))
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
