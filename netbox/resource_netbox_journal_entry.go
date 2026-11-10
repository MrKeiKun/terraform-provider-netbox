package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxJournalEntry() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxJournalEntryCreate,
		Read:   resourceNetboxJournalEntryRead,
		Update: resourceNetboxJournalEntryUpdate,
		Delete: resourceNetboxJournalEntryDelete,

		Description: `:meta:subcategory:Extras:From the [official documentation](https://docs.netbox.dev/en/stable/models/features/journaling/):

> A journal entry is a note that records a change or an event for an object.`,

		Schema: map[string]*schema.Schema{
			"assigned_object_type": {
				Type:        schema.TypeString,
				Description: "The assigned object type of the journal entry.",
				Required:    true,
			},
			"assigned_object_id": {
				Type:         schema.TypeInt,
				Description:  "The ID of the assigned object associated with the journal entry.",
				Required:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"kind": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "One of [info, success, warning, danger]",
				ValidateFunc: validation.StringInSlice([]string{"info", "success", "warning", "danger"}, false),
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the journal entry.",
				Required:    true,
			},
			tagsKey:         tagsSchema,
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const journalEntryAPIPath = "extras/journal-entries/"

type journalEntryRequest struct {
	AssignedObjectType string             `json:"assigned_object_type"`
	AssignedObjectID   int64              `json:"assigned_object_id"`
	Kind               string             `json:"kind,omitempty"`
	Comments           string             `json:"comments"`
	Tags               []netboxapi.TagRef `json:"tags"`
	CustomFields       interface{}        `json:"custom_fields,omitempty"`
}

type journalEntryResponse struct {
	AssignedObjectType string                 `json:"assigned_object_type"`
	AssignedObjectID   *int64                 `json:"assigned_object_id"`
	Kind               *netboxapi.Choice      `json:"kind"`
	Comments           string                 `json:"comments"`
	Tags               []netboxapi.TagRef     `json:"tags"`
	CustomFields       map[string]interface{} `json:"custom_fields"`
}

func journalEntryRequestFromResourceData(api *providerState, d *schema.ResourceData) (*journalEntryRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &journalEntryRequest{
		AssignedObjectType: d.Get("assigned_object_type").(string),
		AssignedObjectID:   int64(d.Get("assigned_object_id").(int)),
		Kind:               d.Get("kind").(string),
		Comments:           d.Get("comments").(string),
		Tags:               tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxJournalEntryCreate(d *schema.ResourceData, m interface{}) error {
	data, err := journalEntryRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, journalEntryAPIPath, data, resourceNetboxJournalEntryRead)
}

func resourceNetboxJournalEntryRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, journalEntryAPIPath, func(r *journalEntryResponse) {
		d.Set("assigned_object_type", r.AssignedObjectType)
		d.Set("assigned_object_id", r.AssignedObjectID)
		d.Set("kind", netboxapi.ChoiceValue(r.Kind))
		d.Set("comments", r.Comments)
		api.readTagRefs(d, r.Tags)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
	})
}

func resourceNetboxJournalEntryUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := journalEntryRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, journalEntryAPIPath, data, resourceNetboxJournalEntryRead)
}

func resourceNetboxJournalEntryDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, journalEntryAPIPath)
}
