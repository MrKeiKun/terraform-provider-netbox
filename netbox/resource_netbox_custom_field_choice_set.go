package netbox

import (
	"errors"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var resourceNetboxCustomFieldChoiceSetBaseChoicesOptions = []string{"IATA", "ISO_3166", "UN_LOCODE"}

func resourceNetboxCustomFieldChoiceSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxCustomFieldChoiceSetCreate,
		Read:   resourceNetboxCustomFieldChoiceSetRead,
		Update: resourceNetboxCustomFieldChoiceSetUpdate,
		Delete: resourceNetboxCustomFieldChoiceSetDelete,

		Description: `:meta:subcategory:Extras:From the [official documentation](https://docs.netbox.dev/en/stable/models/extras/customfieldchoiceset/):

Single- and multi-selection custom fields must define a set of valid choices from which the user may choose when defining the field value. These choices are defined as sets that may be reused among multiple custom fields.

A choice set must define a base choice set and/or a set of arbitrary extra choices.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"base_choices": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(resourceNetboxCustomFieldChoiceSetBaseChoicesOptions, false),
				Description:  buildValidValueDescription(resourceNetboxCustomFieldChoiceSetBaseChoicesOptions),
				AtLeastOneOf: []string{"base_choices", "extra_choices"},
			},
			"extra_choices": {
				Type:        schema.TypeList,
				Description: "This length of the inner lists must be exactly two, where the first value is the value of a choice and the second value is the label of the choice.",
				Elem: &schema.Schema{
					Type: schema.TypeList,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				Optional:     true,
				AtLeastOneOf: []string{"base_choices", "extra_choices"},
				//				ValidateFunc: func (i interface{}, k string) (warnings []string, _errors []error) {
				//	// Outer list length must be > 0
				//	extraChoiceListList := i.([][]string)
				//	if len(extraChoiceListList) == 0 {
				//		return nil, []error{errors.New("length of list must be > 0")}
				//	}
				//
				//	// Inner list length must be exactly 2
				//	for _, innerList := range extraChoiceListList {
				//		if len(innerList) != 2 {
				//			return nil, []error{errors.New("length of list must be > 0")}
				//		}
				//	}
				//
				//	return warnings, _errors
				//},
			},
			"order_alphabetically": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "experimental",
				Default:     false,
			},
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const customFieldChoiceSetAPIPath = "extras/custom-field-choice-sets/"

type customFieldChoiceSetRequest struct {
	Name                string     `json:"name"`
	Description         string     `json:"description,omitempty"`
	OrderAlphabetically bool       `json:"order_alphabetically,omitempty"`
	ExtraChoices        [][]string `json:"extra_choices"`
}

type customFieldChoiceSetResponse struct {
	Name                string            `json:"name"`
	Description         string            `json:"description"`
	ExtraChoices        [][]string        `json:"extra_choices"`
	OrderAlphabetically bool              `json:"order_alphabetically"`
	BaseChoices         *netboxapi.Choice `json:"base_choices"`
}

func customFieldChoiceSetRequestFromResourceData(d *schema.ResourceData) (*customFieldChoiceSetRequest, error) {
	data := &customFieldChoiceSetRequest{
		Name:                d.Get("name").(string),
		OrderAlphabetically: d.Get("order_alphabetically").(bool),
		Description:         getOptionalStr(d, "description", true),
	}

	if extraChoices, ok := d.GetOk("extra_choices"); ok {
		var extraChoiceListList [][]string
		for _, innerList := range extraChoices.([]interface{}) {
			tmp := innerList.([]interface{})
			if len(tmp) != 2 {
				return nil, errors.New("length of inner lists must be exactly two for custom field choice sets")
			}
			extraChoiceListList = append(extraChoiceListList, []string{tmp[0].(string), tmp[1].(string)})
		}
		data.ExtraChoices = extraChoiceListList
	}
	return data, nil
}

func resourceNetboxCustomFieldChoiceSetCreate(d *schema.ResourceData, m interface{}) error {
	data, err := customFieldChoiceSetRequestFromResourceData(d)
	if err != nil {
		return err
	}
	return createObject(d, m, customFieldChoiceSetAPIPath, data, resourceNetboxCustomFieldChoiceSetRead)
}

func resourceNetboxCustomFieldChoiceSetRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, customFieldChoiceSetAPIPath, func(choiceSet *customFieldChoiceSetResponse) {
		d.Set("name", choiceSet.Name)
		if choiceSet.Description != "" {
			d.Set("description", choiceSet.Description)
		} else {
			d.Set("description", nil)
		}
		if choiceSet.ExtraChoices != nil {
			d.Set("extra_choices", choiceSet.ExtraChoices)
		} else {
			d.Set("extra_choices", nil)
		}
		d.Set("order_alphabetically", choiceSet.OrderAlphabetically)
		if choiceSet.BaseChoices != nil {
			d.Set("base_choices", choiceSet.BaseChoices.Value)
		} else {
			d.Set("base_choices", nil)
		}
	})
}

func resourceNetboxCustomFieldChoiceSetUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := customFieldChoiceSetRequestFromResourceData(d)
	if err != nil {
		return err
	}
	return patchObject(d, m, customFieldChoiceSetAPIPath, data, resourceNetboxCustomFieldChoiceSetRead)
}

func resourceNetboxCustomFieldChoiceSetDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, customFieldChoiceSetAPIPath)
}
