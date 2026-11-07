package netbox

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var resourceNetboxWebhookHTTPMethodOptions = []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

func resourceNetboxWebhook() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxWebhookCreate,
		Read:   resourceNetboxWebhookRead,
		Update: resourceNetboxWebhookUpdate,
		Delete: resourceNetboxWebhookDelete,

		Description: `:meta:subcategory:Extras:From the [official documentation](https://docs.netbox.dev/en/stable/integrations/webhooks/):

> A webhook is a mechanism for conveying to some external system a change that took place in NetBox. For example, you may want to notify a monitoring system whenever the status of a device is updated in NetBox. This can be done by creating a webhook for the device model in NetBox and identifying the webhook receiver. When NetBox detects a change to a device, an HTTP request containing the details of the change and who made it be sent to the specified receiver.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"payload_url": {
				Type:     schema.TypeString,
				Required: true,
			},
			"body_template": {
				Type:     schema.TypeString,
				Optional: true,
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
				Type:     schema.TypeString,
				Optional: true,
			},
			"ca_file_path": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const webhookAPIPath = "extras/webhooks/"

type webhookRequest struct {
	Name              string `json:"name"`
	PayloadUrl        string `json:"payload_url"`
	BodyTemplate      string `json:"body_template,omitempty"`
	HttpMethod        string `json:"http_method,omitempty"`
	HttpContentType   string `json:"http_content_type,omitempty"`
	AdditionalHeaders string `json:"additional_headers,omitempty"`
	CaFilePath        string `json:"ca_file_path"`
}

type webhookResponse struct {
	Name              string `json:"name"`
	PayloadUrl        string `json:"payload_url"`
	BodyTemplate      string `json:"body_template"`
	HttpMethod        string `json:"http_method"`
	HttpContentType   string `json:"http_content_type"`
	AdditionalHeaders string `json:"additional_headers"`
	CaFilePath        string `json:"ca_file_path"`
}

func webhookRequestFromResourceData(d *schema.ResourceData) *webhookRequest {
	name := d.Get("name").(string)
	data := &webhookRequest{
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
