package netbox

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxDeviceRenderConfig() *schema.Resource {
	return &schema.Resource{
		ReadContext: withContext(dataSourceNetboxDeviceRenderConfigRead),
		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):Render the configuration template assigned to a device using the device's config context.`,
		Schema: map[string]*schema.Schema{
			"device_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The ID of the device to render configuration for.",
			},
			"content": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The rendered configuration content.",
			},
			"config_template_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The ID of the config template that was used for rendering.",
			},
			"config_template_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the config template that was used for rendering.",
			},
		},
	}
}

type deviceRenderConfigResponse struct {
	Content        string `json:"content"`
	Configtemplate *struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"configtemplate"`
}

func dataSourceNetboxDeviceRenderConfigRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)

	deviceID := int64(d.Get("device_id").(int))

	// Rendering a config is an action on the device, NetBox wants a POST for it.
	path := fmt.Sprintf("dcim/devices/%d/render-config/", deviceID)
	result, err := netboxapi.Request[deviceRenderConfigResponse](context.Background(), api.rest, http.MethodPost, path, nil, map[string]interface{}{})
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatInt(deviceID, 10))
	d.Set("content", result.Content)
	if result.Configtemplate != nil {
		d.Set("config_template_id", result.Configtemplate.ID)
		d.Set("config_template_name", result.Configtemplate.Name)
	}
	return nil
}
