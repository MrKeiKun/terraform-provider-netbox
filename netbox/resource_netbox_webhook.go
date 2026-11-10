package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var resourceNetboxWebhookHTTPMethodOptions = []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

func resourceNetboxWebhook() *schema.Resource {
	return &schema.Resource{
		CreateContext: withContext(resourceNetboxWebhookCreate),
		ReadContext:   withContext(resourceNetboxWebhookRead),
		UpdateContext: withContext(resourceNetboxWebhookUpdate),
		DeleteContext: withContext(resourceNetboxWebhookDelete),

		Description: `:meta:subcategory:Extras:From the [official documentation](https://docs.netbox.dev/en/stable/integrations/webhooks/):

> A webhook is a mechanism for conveying to some external system a change that took place in NetBox. For example, you may want to notify a monitoring system whenever the status of a device is updated in NetBox. This can be done by creating a webhook for the device model in NetBox and identifying the webhook receiver. When NetBox detects a change to a device, an HTTP request containing the details of the change and who made it be sent to the specified receiver.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the webhook.",
				Required:    true,
			},
			"payload_url": {
				Type:        schema.TypeString,
				Description: "The payload url of the webhook.",
				Required:    true,
			},
			"body_template": {
				Type:        schema.TypeString,
				Description: "The body template of the webhook.",
				Optional:    true,
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					equal, _ := jsonSemanticCompare(oldValue, newValue)
					return equal
				},
				DiffSuppressOnRefresh: true,
			},
			"http_method": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(resourceNetboxWebhookHTTPMethodOptions, false),
				Description:  buildValidValueDescription(resourceNetboxWebhookHTTPMethodOptions),
				Default:      "POST",
			},
			"http_content_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The complete list of official content types is available [here](https://www.iana.org/assignments/media-types/media-types.xhtml).",
				Default:     "application/json",
			},
			"additional_headers": {
				Type:        schema.TypeString,
				Description: "The additional headers of the webhook.",
				Optional:    true,
			},
			"ca_file_path": {
				Type:        schema.TypeString,
				Description: "The ca file path of the webhook.",
				Optional:    true,
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the webhook.",
				Optional:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the webhook.",
				Optional:    true,
			},
			"ssl_verification": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Enable SSL certificate verification. Disable with caution!",
			},
			"timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 3600),
				Description:  "The maximum time (in seconds) to wait for a response before failing the request. Leave blank to use the system default (60 seconds).",
			},
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const webhookAPIPath = "extras/webhooks/"

type webhookRequest struct {
	Name              string      `json:"name"`
	PayloadUrl        string      `json:"payload_url"`
	BodyTemplate      string      `json:"body_template,omitempty"`
	HttpMethod        string      `json:"http_method,omitempty"`
	HttpContentType   string      `json:"http_content_type,omitempty"`
	AdditionalHeaders string      `json:"additional_headers,omitempty"`
	CaFilePath        string      `json:"ca_file_path"`
	Owner             *int64      `json:"owner"`
	Description       string      `json:"description,omitempty"`
	SslVerification   *bool       `json:"ssl_verification,omitempty"`
	Timeout           *int64      `json:"timeout"`
	CustomFields      interface{} `json:"custom_fields,omitempty"`
}

type webhookResponse struct {
	Name              string                 `json:"name"`
	PayloadUrl        string                 `json:"payload_url"`
	BodyTemplate      string                 `json:"body_template"`
	HttpMethod        string                 `json:"http_method"`
	HttpContentType   string                 `json:"http_content_type"`
	AdditionalHeaders string                 `json:"additional_headers"`
	CaFilePath        string                 `json:"ca_file_path"`
	Owner             *netboxapi.Ref         `json:"owner"`
	Description       string                 `json:"description"`
	SslVerification   bool                   `json:"ssl_verification"`
	Timeout           *int64                 `json:"timeout"`
	CustomFields      map[string]interface{} `json:"custom_fields"`
}

func webhookRequestFromResourceData(d *schema.ResourceData) *webhookRequest {
	name := d.Get("name").(string)
	data := &webhookRequest{
		CustomFields:      customFieldsRequest(d),
		Timeout:           getOptionalInt(d, "timeout"),
		SslVerification:   getOptionalBool(d, "ssl_verification"),
		Description:       getOptionalStr(d, "description", true),
		Owner:             getOptionalInt(d, "owner_id"),
		Name:              name,
		PayloadUrl:        d.Get("payload_url").(string),
		BodyTemplate:      d.Get("body_template").(string),
		HttpMethod:        d.Get("http_method").(string),
		HttpContentType:   d.Get("http_content_type").(string),
		AdditionalHeaders: d.Get("additional_headers").(string),
		CaFilePath:        d.Get("ca_file_path").(string),
	}
	return data
}

func resourceNetboxWebhookCreate(d *schema.ResourceData, m interface{}) error {
	return createObject(d, m, webhookAPIPath, webhookRequestFromResourceData(d), resourceNetboxWebhookRead)
}

func resourceNetboxWebhookRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, webhookAPIPath, func(r *webhookResponse) {
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		d.Set("timeout", r.Timeout)
		d.Set("ssl_verification", r.SslVerification)
		d.Set("description", r.Description)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("name", r.Name)
		d.Set("payload_url", r.PayloadUrl)
		d.Set("body_template", r.BodyTemplate)
		d.Set("http_method", r.HttpMethod)
		d.Set("http_content_type", r.HttpContentType)
		d.Set("additional_headers", r.AdditionalHeaders)
		d.Set("ca_file_path", r.CaFilePath)
	})
}

func resourceNetboxWebhookUpdate(d *schema.ResourceData, m interface{}) error {
	return putObject(d, m, webhookAPIPath, webhookRequestFromResourceData(d), resourceNetboxWebhookRead)
}

func resourceNetboxWebhookDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, webhookAPIPath)
}
