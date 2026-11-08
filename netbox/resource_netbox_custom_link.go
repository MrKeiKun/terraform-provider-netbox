package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxCustomLink() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxCustomLinkCreate,
		Read:   resourceNetboxCustomLinkRead,
		Update: resourceNetboxCustomLinkUpdate,
		Delete: resourceNetboxCustomLinkDelete,

		Description: `:meta:subcategory:Extras:From the [official documentation](https://docs.netbox.dev/en/stable/models/customization/customlink/):

> A custom link adds a link to the page of an object, built from Jinja2 templates.`,

		Schema: map[string]*schema.Schema{
			"object_types": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"link_text": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Jinja2 template code for link text",
			},
			"link_url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Jinja2 template code for link URL",
			},
			"weight": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 32767),
			},
			"group_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Links with the same group will appear as a dropdown menu",
			},
			"button_class": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "One of [default, blue, indigo, purple, pink, red, orange, yellow, green, teal, cyan, gray, black, white, ghost-dark]",
				ValidateFunc: validation.StringInSlice([]string{"default", "blue", "indigo", "purple", "pink", "red", "orange", "yellow", "green", "teal", "cyan", "gray", "black", "white", "ghost-dark"}, false),
			},
			"new_window": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Force link to open in a new window",
			},
			"owner_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const customLinkAPIPath = "extras/custom-links/"

type customLinkRequest struct {
	ObjectTypes []string `json:"object_types"`
	Name        string   `json:"name"`
	Enabled     bool     `json:"enabled"`
	LinkText    string   `json:"link_text"`
	LinkURL     string   `json:"link_url"`
	Weight      int64    `json:"weight,omitempty"`
	GroupName   string   `json:"group_name,omitempty"`
	ButtonClass string   `json:"button_class,omitempty"`
	NewWindow   bool     `json:"new_window"`
	Owner       *int64   `json:"owner"`
}

type customLinkResponse struct {
	ObjectTypes []string       `json:"object_types"`
	Name        string         `json:"name"`
	Enabled     bool           `json:"enabled"`
	LinkText    string         `json:"link_text"`
	LinkURL     string         `json:"link_url"`
	Weight      *int64         `json:"weight"`
	GroupName   string         `json:"group_name"`
	ButtonClass string         `json:"button_class"`
	NewWindow   bool           `json:"new_window"`
	Owner       *netboxapi.Ref `json:"owner"`
}

func customLinkRequestFromResourceData(api *providerState, d *schema.ResourceData) (*customLinkRequest, error) {
	data := &customLinkRequest{
		ObjectTypes: getStringSet(d, "object_types"),
		Name:        d.Get("name").(string),
		Enabled:     d.Get("enabled").(bool),
		LinkText:    d.Get("link_text").(string),
		LinkURL:     d.Get("link_url").(string),
		Weight:      int64(d.Get("weight").(int)),
		GroupName:   getOptionalStr(d, "group_name", true),
		ButtonClass: d.Get("button_class").(string),
		NewWindow:   d.Get("new_window").(bool),
		Owner:       getOptionalInt(d, "owner_id"),
	}
	return data, nil
}

func resourceNetboxCustomLinkCreate(d *schema.ResourceData, m interface{}) error {
	data, err := customLinkRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, customLinkAPIPath, data, resourceNetboxCustomLinkRead)
}

func resourceNetboxCustomLinkRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, customLinkAPIPath, func(r *customLinkResponse) {
		d.Set("object_types", r.ObjectTypes)
		d.Set("name", r.Name)
		d.Set("enabled", r.Enabled)
		d.Set("link_text", r.LinkText)
		d.Set("link_url", r.LinkURL)
		d.Set("weight", r.Weight)
		d.Set("group_name", r.GroupName)
		d.Set("button_class", r.ButtonClass)
		d.Set("new_window", r.NewWindow)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
	})
}

func resourceNetboxCustomLinkUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := customLinkRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, customLinkAPIPath, data, resourceNetboxCustomLinkRead)
}

func resourceNetboxCustomLinkDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, customLinkAPIPath)
}
