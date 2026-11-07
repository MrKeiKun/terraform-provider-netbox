package netbox

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var resourceNetboxDeviceStatusOptions = []string{"offline", "active", "planned", "staged", "failed", "inventory", "decommissioning"}
var resourceNetboxDeviceRackFaceOptions = []string{"front", "rear"}

func resourceNetboxDevice() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetboxDeviceCreate,
		ReadContext:   resourceNetboxDeviceRead,
		UpdateContext: resourceNetboxDeviceUpdate,
		DeleteContext: resourceNetboxDeviceDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/features/devices/#devices):

> Every piece of hardware which is installed within a site or rack exists in NetBox as a device. Devices are measured in rack units (U) and can be half depth or full depth. A device may have a height of 0U: These devices do not consume vertical rack space and cannot be assigned to a particular rack unit. A common example of a 0U device is a vertically-mounted PDU.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"device_type_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"tenant_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"cluster_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"platform_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"location_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"role_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"serial": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"site_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"config_template_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"comments": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"asset_tag": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			tagsKey: tagsSchema,
			"primary_ipv4": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"primary_ipv6": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(resourceNetboxDeviceStatusOptions, false),
				Description:  buildValidValueDescription(resourceNetboxDeviceStatusOptions),
				Default:      "active",
			},
			"rack_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"rack_face": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"rack_position"},
				ValidateFunc: validation.StringInSlice(resourceNetboxDeviceRackFaceOptions, false),
				Description:  buildValidValueDescription(resourceNetboxDeviceRackFaceOptions),
			},
			"rack_position": {
				Type:     schema.TypeFloat,
				Optional: true,
			},
			"virtual_chassis_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				RequiredWith: []string{"virtual_chassis_master", "virtual_chassis_id"},
			},
			"virtual_chassis_position": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"virtual_chassis_priority": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"virtual_chassis_master": {
				Type:         schema.TypeBool,
				Optional:     true,
				RequiredWith: []string{"virtual_chassis_master", "virtual_chassis_id"},
			},
			"local_context_data": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "This is best managed through the use of `jsonencode` and a map of settings.",
			},
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

type deviceResponse struct {
	ID             int64             `json:"id"`
	Name           string            `json:"name"`
	DeviceType     *netboxapi.Ref    `json:"device_type"`
	PrimaryIp4     *netboxapi.Ref    `json:"primary_ip4"`
	PrimaryIp6     *netboxapi.Ref    `json:"primary_ip6"`
	Tenant         *netboxapi.Ref    `json:"tenant"`
	Platform       *netboxapi.Ref    `json:"platform"`
	Location       *netboxapi.Ref    `json:"location"`
	Cluster        *netboxapi.Ref    `json:"cluster"`
	Role           *netboxapi.Ref    `json:"role"`
	Site           *netboxapi.Ref    `json:"site"`
	ConfigTemplate *netboxapi.Ref    `json:"config_template"`
	Rack           *netboxapi.Ref    `json:"rack"`
	AssetTag       *string           `json:"asset_tag"`
	Comments       string            `json:"comments"`
	Description    string            `json:"description"`
	Serial         string            `json:"serial"`
	Status         *netboxapi.Choice `json:"status"`
	Face           *netboxapi.Choice `json:"face"`
	Position       *float64          `json:"position"`
	VirtualChassis *struct {
		ID     int64          `json:"id"`
		Master *netboxapi.Ref `json:"master"`
	} `json:"virtual_chassis"`
	VcPosition       *int64                 `json:"vc_position"`
	VcPriority       *int64                 `json:"vc_priority"`
	LocalContextData interface{}            `json:"local_context_data"`
	Tags             []netboxapi.TagRef     `json:"tags"`
	CustomFields     map[string]interface{} `json:"custom_fields"`
}

// deviceBody builds the request for create and update. Optional values that are
// not set are left out.
func deviceBody(api *providerState, d *schema.ResourceData, isUpdate bool) (map[string]interface{}, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}

	body := map[string]interface{}{
		"name": d.Get("name").(string),
		"tags": tags,
		// these three are required by NetBox, so they are always sent
		"device_type": nullableID(d.Get("device_type_id").(int)),
		"role":        nullableID(d.Get("role_id").(int)),
		"site":        nullableID(d.Get("site_id").(int)),
	}

	if status := d.Get("status").(string); status != "" {
		body["status"] = status
	}

	for key, field := range map[string]string{
		"tenant_id":          "tenant",
		"platform_id":        "platform",
		"location_id":        "location",
		"cluster_id":         "cluster",
		"config_template_id": "config_template",
	} {
		if v, ok := d.GetOk(key); ok {
			body[field] = int64(v.(int))
		}
	}

	if rack := getOptionalInt(d, "rack_id"); rack != nil {
		body["rack"] = *rack
	}
	if face := getOptionalStr(d, "rack_face", false); face != "" {
		body["face"] = face
	}
	if vc := getOptionalInt(d, "virtual_chassis_id"); vc != nil {
		body["virtual_chassis"] = *vc
	}
	if pos := getOptionalInt(d, "virtual_chassis_position"); pos != nil {
		body["vc_position"] = *pos
	}
	if priority := getOptionalInt(d, "virtual_chassis_priority"); priority != nil {
		body["vc_priority"] = *priority
	}

	if localContext, ok := d.GetOk("local_context_data"); ok {
		var jsonObj any
		if err := json.Unmarshal([]byte(localContext.(string)), &jsonObj); err == nil {
			body["local_context_data"] = jsonObj
		}
	}
	if cf, ok := d.GetOk(customFieldsKey); ok {
		body["custom_fields"] = cf
	}

	if isUpdate {
		if position := getOptionalFloat(d, "rack_position"); position != nil {
			body["position"] = *position
		}

		// A value that is cleared has to be sent as a space, to make NetBox clear it.
		if d.HasChanges("asset_tag") {
			if assetTag, ok := d.GetOk("asset_tag"); ok {
				body["asset_tag"] = assetTag.(string)
			} else {
				body["asset_tag"] = " "
			}
		}
		if comments := getOptionalStr(d, "comments", true); comments != "" {
			body["comments"] = comments
		}
		if description := getOptionalStr(d, "description", true); description != "" {
			body["description"] = description
		}
		if serial := getOptionalStr(d, "serial", true); serial != "" {
			body["serial"] = serial
		}
	} else {
		if assetTag, ok := d.GetOk("asset_tag"); ok {
			body["asset_tag"] = assetTag.(string)
		}
		if comments := d.Get("comments").(string); comments != "" {
			body["comments"] = comments
		}
		if description := d.Get("description").(string); description != "" {
			body["description"] = description
		}
		if serial := d.Get("serial").(string); serial != "" {
			body["serial"] = serial
		}
		if rackPosition, ok := d.GetOk("rack_position"); ok && rackPosition.(float64) > 0 {
			body["position"] = rackPosition.(float64)
		}
	}

	return body, nil
}

func resourceNetboxDeviceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*providerState)

	body, err := deviceBody(api, d, false)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := netboxapi.Create[netboxapi.Ref](ctx, api.rest, deviceAPIPath, body)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(strconv.FormatInt(res.ID, 10))

	if vcMaster, ok := d.GetOk("virtual_chassis_master"); ok {
		if vcID := getOptionalInt(d, "virtual_chassis_id"); vcID != nil {
			var err error
			if vcMaster.(bool) {
				err = virtualChassisUpdateMaster(api, *vcID, &res.ID)
			} else {
				err = virtualChassisUpdateMaster(api, *vcID, nil)
			}
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return resourceNetboxDeviceRead(ctx, d, m)
}

func resourceNetboxDeviceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*providerState)
	return readObjectCtx(ctx, d, m, deviceAPIPath, func(device *deviceResponse) {
		d.Set("name", device.Name)
		if device.DeviceType != nil {
			d.Set("device_type_id", device.DeviceType.ID)
		}
		d.Set("primary_ipv4", netboxapi.RefID(device.PrimaryIp4))
		d.Set("primary_ipv6", netboxapi.RefID(device.PrimaryIp6))
		d.Set("tenant_id", netboxapi.RefID(device.Tenant))
		d.Set("platform_id", netboxapi.RefID(device.Platform))
		d.Set("location_id", netboxapi.RefID(device.Location))
		d.Set("cluster_id", netboxapi.RefID(device.Cluster))
		d.Set("role_id", netboxapi.RefID(device.Role))
		d.Set("site_id", netboxapi.RefID(device.Site))
		d.Set("config_template_id", netboxapi.RefID(device.ConfigTemplate))

		if cf := getCustomFields(device.CustomFields); cf != nil {
			d.Set(customFieldsKey, cf)
		}

		d.Set("asset_tag", device.AssetTag)
		d.Set("comments", device.Comments)
		d.Set("description", device.Description)
		d.Set("serial", device.Serial)
		d.Set("status", netboxapi.ChoiceValue(device.Status))
		d.Set("rack_id", netboxapi.RefID(device.Rack))
		if device.Face != nil {
			d.Set("rack_face", device.Face.Value)
		} else {
			d.Set("rack_face", nil)
		}
		d.Set("rack_position", device.Position)

		if device.VirtualChassis != nil {
			d.Set("virtual_chassis_id", device.VirtualChassis.ID)
			d.Set(
				"virtual_chassis_master",
				device.VirtualChassis.Master != nil && device.VirtualChassis.Master.ID == device.ID,
			)
		} else {
			d.Set("virtual_chassis_id", 0)
			d.Set("virtual_chassis_master", false)
		}
		d.Set("virtual_chassis_position", device.VcPosition)
		d.Set("virtual_chassis_priority", device.VcPriority)

		if device.LocalContextData != nil {
			if jsonArr, err := json.Marshal(device.LocalContextData); err == nil {
				d.Set("local_context_data", string(jsonArr))
			}
		} else {
			d.Set("local_context_data", nil)
		}

		api.readTagRefs(d, device.Tags)
	})
}

func resourceNetboxDeviceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*providerState)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)

	body, err := deviceBody(api, d, true)
	if err != nil {
		return diag.FromErr(err)
	}

	if _, err := netboxapi.Update[netboxapi.Ref](ctx, api.rest, deviceAPIPath, id, body); err != nil {
		return diag.FromErr(err)
	}

	if vcID := getOptionalInt(d, "virtual_chassis_id"); vcID != nil && d.HasChange("virtual_chassis_master") {
		var err error
		if vcMaster, ok := d.GetOk("virtual_chassis_master"); ok && vcMaster.(bool) {
			err = virtualChassisUpdateMaster(api, *vcID, &id)
		} else {
			// It was set before, but no longer set, remove it as master
			err = virtualChassisUpdateMaster(api, *vcID, nil)
		}
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceNetboxDeviceRead(ctx, d, m)
}

func resourceNetboxDeviceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*providerState)

	// If the device is member of a virtual chassis and it's the master, we cannot
	// delete it directly. We first need to update it to not be the master.
	if virtualChassisIDValue, ok := d.GetOk("virtual_chassis_id"); ok {
		if d.Get("virtual_chassis_master").(bool) {
			if err := virtualChassisUpdateMaster(api, int64(virtualChassisIDValue.(int)), nil); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return deleteObjectCtx(ctx, d, m, deviceAPIPath)
}
