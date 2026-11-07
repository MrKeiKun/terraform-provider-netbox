package netbox

import (
	"context"
	"strconv"
	"time"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxToken() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetboxTokenCreate,
		ReadContext:   resourceNetboxTokenRead,
		UpdateContext: resourceNetboxTokenUpdate,
		DeleteContext: resourceNetboxTokenDelete,

		Description: `:meta:subcategory:Authentication:From the [official documentation](https://docs.netbox.dev/en/stable/rest-api/authentication/#tokens):

> A token is a unique identifier mapped to a NetBox user account. Each user may have one or more tokens which he or she can use for authentication when making REST API requests. To create a token, navigate to the API tokens page under your user profile.`,

		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"key": {
				Type:         schema.TypeString,
				Sensitive:    true,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(40, 256),
			},
			"allowed_ips": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsCIDR,
				},
			},
			"write_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"last_used": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"expires": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const tokenAPIPath = "users/tokens/"

// tokenTimeFormat is how NetBox timestamps are written to the state.
const tokenTimeFormat = "2006-01-02T15:04:05.000Z"

type tokenRequest struct {
	User         int64         `json:"user"`
	Key          string        `json:"key,omitempty"`
	AllowedIps   []interface{} `json:"allowed_ips"`
	WriteEnabled bool          `json:"write_enabled"`
	Description  string        `json:"description,omitempty"`
	Expires      *time.Time    `json:"expires,omitempty"`
}

type tokenResponse struct {
	User         *netboxapi.Ref `json:"user"`
	Key          string         `json:"key"`
	LastUsed     *string        `json:"last_used"`
	Expires      *string        `json:"expires"`
	AllowedIps   []string       `json:"allowed_ips"`
	WriteEnabled bool           `json:"write_enabled"`
	Description  string         `json:"description"`
}

func tokenRequestFromResourceData(d *schema.ResourceData) (*tokenRequest, error) {
	allowedIps := d.Get("allowed_ips").([]interface{})

	data := &tokenRequest{
		User:         int64(d.Get("user_id").(int)),
		Key:          d.Get("key").(string),
		AllowedIps:   make([]interface{}, len(allowedIps)),
		WriteEnabled: d.Get("write_enabled").(bool),
		Description:  d.Get("description").(string),
	}
	copy(data.AllowedIps, allowedIps)

	if expiresRaw, ok := d.GetOk("expires"); ok {
		expires, err := time.Parse(time.RFC3339, expiresRaw.(string))
		if err != nil {
			return nil, err
		}
		data.Expires = &expires
	}
	return data, nil
}

// formatTokenTime rewrites a NetBox timestamp the way the state has always stored it.
func formatTokenTime(s string) string {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	return t.UTC().Format(tokenTimeFormat)
}

func resourceNetboxTokenCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*providerState)

	data, err := tokenRequestFromResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := netboxapi.Create[netboxapi.Ref](ctx, api.rest, tokenAPIPath, data)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(strconv.FormatInt(res.ID, 10))

	return resourceNetboxTokenUpdate(ctx, d, m)
}

func resourceNetboxTokenRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*providerState)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)

	token, err := netboxapi.Get[tokenResponse](ctx, api.rest, tokenAPIPath, id)
	if err != nil {
		if netboxapi.IsNotFound(err) {
			// If the ID is updated to blank, this tells Terraform the resource no longer exists (maybe it was destroyed out of band). Just like the destroy callback, the Read function should gracefully handle this case. https://www.terraform.io/docs/extend/writing-custom-providers.html
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if token.User != nil {
		d.Set("user_id", token.User.ID)
	}

	// Since NetBox 4.3.0, ALLOW_TOKEN_RETRIEVAL is disabled by default
	// This means we will usually not get a Key value from the API
	if token.Key != "" {
		d.Set("key", token.Key)
	}
	if token.LastUsed != nil {
		d.Set("last_used", formatTokenTime(*token.LastUsed))
	}
	if token.Expires != nil {
		d.Set("expires", formatTokenTime(*token.Expires))
	}
	d.Set("allowed_ips", token.AllowedIps)
	d.Set("write_enabled", token.WriteEnabled)
	d.Set("description", token.Description)

	return nil
}

func resourceNetboxTokenUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*providerState)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)

	data, err := tokenRequestFromResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	if _, err := netboxapi.Update[netboxapi.Ref](ctx, api.rest, tokenAPIPath, id, data); err != nil {
		return diag.FromErr(err)
	}
	return resourceNetboxTokenRead(ctx, d, m)
}

func resourceNetboxTokenDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*providerState)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)

	if err := netboxapi.Delete(ctx, api.rest, tokenAPIPath, id); err != nil {
		if netboxapi.IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}
