package netbox

import (
	"encoding/json"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxDataSource() *schema.Resource {
	return &schema.Resource{
		CreateContext: withContext(resourceNetboxDataSourceCreate),
		ReadContext:   withContext(resourceNetboxDataSourceRead),
		UpdateContext: withContext(resourceNetboxDataSourceUpdate),
		DeleteContext: withContext(resourceNetboxDataSourceDelete),

		Description: `:meta:subcategory:Extras:From the [official documentation](https://docs.netbox.dev/en/stable/models/core/datasource/):

> A data source is a remote location, such as a git repository, that NetBox syncs files from.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the data source.",
				Required:    true,
			},
			"type": {
				Type:        schema.TypeString,
				Description: "The type of the data source.",
				Required:    true,
			},
			"source_url": {
				Type:        schema.TypeString,
				Description: "The source url of the data source.",
				Required:    true,
			},
			"enabled": {
				Type:        schema.TypeBool,
				Description: "Whether the data source is enabled.",
				Optional:    true,
				Default:     false,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the data source.",
				Optional:    true,
			},
			"sync_interval": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "One of [1, 60, 720, 1440, 10080, 43200]",
				ValidateFunc: validation.IntInSlice([]int{1, 60, 720, 1440, 10080, 43200}),
			},
			"parameters": {
				Type:             schema.TypeString,
				Description:      "The parameters of the data source.",
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentJSON,
			},
			"ignore_rules": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Patterns (one per line) matching files or paths to ignore when syncing",
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the data source.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the data source.",
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

const dataSourceAPIPath = "core/data-sources/"

type dataSourceRequest struct {
	Name         string             `json:"name"`
	Type         *string            `json:"type,omitempty"`
	SourceURL    string             `json:"source_url"`
	Enabled      bool               `json:"enabled"`
	Description  string             `json:"description,omitempty"`
	SyncInterval *int64             `json:"sync_interval,omitempty"`
	Parameters   *json.RawMessage   `json:"parameters"`
	IgnoreRules  string             `json:"ignore_rules,omitempty"`
	Owner        *int64             `json:"owner"`
	Comments     string             `json:"comments,omitempty"`
	Tags         []netboxapi.TagRef `json:"tags"`
	CustomFields interface{}        `json:"custom_fields,omitempty"`
}

type dataSourceResponse struct {
	Name         string                 `json:"name"`
	Type         *netboxapi.Choice      `json:"type"`
	SourceURL    string                 `json:"source_url"`
	Enabled      bool                   `json:"enabled"`
	Description  string                 `json:"description"`
	SyncInterval *int64                 `json:"sync_interval"`
	Parameters   json.RawMessage        `json:"parameters"`
	IgnoreRules  string                 `json:"ignore_rules"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	Tags         []netboxapi.TagRef     `json:"tags"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func dataSourceRequestFromResourceData(api *providerState, d *schema.ResourceData) (*dataSourceRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	parametersRaw, err := jsonRawFromResourceData(d, "parameters")
	if err != nil {
		return nil, err
	}
	data := &dataSourceRequest{
		Name:         d.Get("name").(string),
		Type:         getOptionalStrClear(d, "type"),
		SourceURL:    d.Get("source_url").(string),
		Enabled:      d.Get("enabled").(bool),
		Description:  getOptionalStr(d, "description", true),
		SyncInterval: getOptionalInt(d, "sync_interval"),
		Parameters:   parametersRaw,
		IgnoreRules:  getOptionalStr(d, "ignore_rules", true),
		Owner:        getOptionalInt(d, "owner_id"),
		Comments:     getOptionalStr(d, "comments", true),
		Tags:         tags,
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxDataSourceCreate(d *schema.ResourceData, m interface{}) error {
	data, err := dataSourceRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, dataSourceAPIPath, data, resourceNetboxDataSourceRead)
}

func resourceNetboxDataSourceRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, dataSourceAPIPath, func(r *dataSourceResponse) {
		d.Set("name", r.Name)
		d.Set("type", netboxapi.ChoiceValue(r.Type))
		d.Set("source_url", r.SourceURL)
		d.Set("enabled", r.Enabled)
		d.Set("description", r.Description)
		d.Set("sync_interval", r.SyncInterval)
		d.Set("parameters", jsonString(r.Parameters))
		d.Set("ignore_rules", r.IgnoreRules)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("comments", r.Comments)
		api.readTagRefs(d, r.Tags)
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
	})
}

func resourceNetboxDataSourceUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := dataSourceRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, dataSourceAPIPath, data, resourceNetboxDataSourceRead)
}

func resourceNetboxDataSourceDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, dataSourceAPIPath)
}
