package netbox

import (
	"encoding/json"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var resourceNetboxEventRuleActionTypeOptions = []string{"webhook", "script"}

var resourceNetboxEventRuleActionTypeToObjectType = map[string]string{
	"webhook": "extras.webhook",
	"script":  "extras.script",
}

func resourceNetboxEventRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxEventRuleCreate,
		Read:   resourceNetboxEventRuleRead,
		Update: resourceNetboxEventRuleUpdate,
		Delete: resourceNetboxEventRuleDelete,

		Description: `:meta:subcategory:Extras:From the [official documentation](https://docs.netbox.dev/en/stable/features/event-rules/):

> NetBox can be configured via Event Rules to transmit outgoing webhooks to remote systems in response to internal object changes. The receiver can act on the data in these webhook messages to perform related tasks.
	Event rules can also execute custom scripts, enabling NetBox to run defined logic locally in response to object changes.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"content_types": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"event_types": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					// We do not enforce event-type validation, because plugins might extend the list of valid events
					Type: schema.TypeString,
				},
				Description: "The types of event which will trigger this rule. By default, valid values are `object_created`, `oject_updated`, `object_deleted`, `job_started`, `job_completed`, `job_failed` and `job_errored`",
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"conditions": {
				Type:     schema.TypeString,
				Optional: true,
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					equal, _ := jsonSemanticCompare(oldValue, newValue)
					return equal
				},
				DiffSuppressOnRefresh: true,
			},
			"action_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(resourceNetboxEventRuleActionTypeOptions, false),
				Description:  buildValidValueDescription(resourceNetboxEventRuleActionTypeOptions),
			},
			"action_object_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			tagsKey: tagsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const eventRuleAPIPath = "extras/event-rules/"

type eventRuleRequest struct {
	Name             string             `json:"name"`
	ActionType       string             `json:"action_type,omitempty"`
	Description      string             `json:"description,omitempty"`
	Enabled          bool               `json:"enabled"`
	ActionObjectId   *int64             `json:"action_object_id"`
	ActionObjectType *string            `json:"action_object_type"`
	EventTypes       []string           `json:"event_types"`
	ObjectTypes      []string           `json:"object_types,omitempty"`
	Conditions       interface{}        `json:"conditions,omitempty"`
	Tags             []netboxapi.TagRef `json:"tags"`
}

type eventRuleResponse struct {
	Name           string             `json:"name"`
	ActionType     *netboxapi.Choice  `json:"action_type"`
	Description    string             `json:"description"`
	Enabled        bool               `json:"enabled"`
	ActionObjectId *int64             `json:"action_object_id"`
	ObjectTypes    []string           `json:"object_types"`
	EventTypes     []string           `json:"event_types"`
	Conditions     interface{}        `json:"conditions"`
	Tags           []netboxapi.TagRef `json:"tags"`
}

func eventRuleRequestFromResourceData(api *providerState, d *schema.ResourceData) (*eventRuleRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &eventRuleRequest{
		Name:           name,
		ActionType:     d.Get("action_type").(string),
		Description:    getOptionalStr(d, "description", true),
		Enabled:        d.Get("enabled").(bool),
		ActionObjectId: getOptionalInt(d, "action_object_id"),

		Tags: tags,
	}

	if objectType, ok := resourceNetboxEventRuleActionTypeToObjectType[data.ActionType]; ok {
		data.ActionObjectType = strToPtr(objectType)
	}

	data.EventTypes = make([]string, 0)
	for _, eventType := range d.Get("event_types").(*schema.Set).List() {
		data.EventTypes = append(data.EventTypes, eventType.(string))
	}

	contentTypes := d.Get("content_types").(*schema.Set).List()
	data.ObjectTypes = make([]string, 0, len(contentTypes))
	for _, contentType := range contentTypes {
		data.ObjectTypes = append(data.ObjectTypes, contentType.(string))
	}

	if conditionsData, ok := d.GetOk("conditions"); ok {
		var conditions any
		if err := json.Unmarshal([]byte(conditionsData.(string)), &conditions); err != nil {
			return nil, err
		}
		data.Conditions = conditions
	}
	return data, nil
}

func resourceNetboxEventRuleCreate(d *schema.ResourceData, m interface{}) error {
	data, err := eventRuleRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, eventRuleAPIPath, data, resourceNetboxEventRuleRead)
}

func resourceNetboxEventRuleRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, eventRuleAPIPath, func(r *eventRuleResponse) {
		d.Set("name", r.Name)
		d.Set("action_type", netboxapi.ChoiceValue(r.ActionType))
		d.Set("description", r.Description)
		d.Set("enabled", r.Enabled)
		d.Set("action_object_id", r.ActionObjectId)

		d.Set("content_types", r.ObjectTypes)
		d.Set("event_types", r.EventTypes)
		if r.Conditions != nil {
			if conditions, err := json.Marshal(r.Conditions); err == nil {
				d.Set("conditions", string(conditions))
			}
		}
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxEventRuleUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := eventRuleRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return putObject(d, m, eventRuleAPIPath, data, resourceNetboxEventRuleRead)
}

func resourceNetboxEventRuleDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, eventRuleAPIPath)
}
