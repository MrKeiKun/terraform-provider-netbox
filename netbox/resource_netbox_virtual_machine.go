package netbox

import (
	"context"
	"encoding/json"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var resourceNetboxVirtualMachineStatusOptions = []string{"offline", "active", "planned", "staged", "failed", "decommissioning"}

func resourceNetboxVirtualMachine() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetboxVirtualMachineCreate,
		ReadContext:   resourceNetboxVirtualMachineRead,
		UpdateContext: resourceNetboxVirtualMachineUpdate,
		DeleteContext: resourceNetboxVirtualMachineDelete,

		Description: `:meta:subcategory:Virtualization:From the [official documentation](https://docs.netbox.dev/en/stable/features/virtualization/#virtual-machines):

> A virtual machine is a virtualized compute instance. These behave in NetBox very similarly to device objects, but without any physical attributes. For example, a VM may have interfaces assigned to it with IP addresses and VLANs, however its interfaces cannot be connected via cables (because they are virtual). Each VM may also define its compute, memory, and storage resources as well.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the virtual machine.",
				Required:    true,
			},
			"cluster_id": {
				Type:         schema.TypeInt,
				Description:  "The ID of the cluster associated with the virtual machine.",
				Optional:     true,
				AtLeastOneOf: []string{"site_id", "cluster_id"},
			},
			"tenant_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the tenant associated with the virtual machine.",
				Optional:    true,
			},
			"device_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the device associated with the virtual machine.",
				Optional:    true,
			},
			"platform_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the platform associated with the virtual machine.",
				Optional:    true,
			},
			"role_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the role associated with the virtual machine.",
				Optional:    true,
			},
			"site_id": {
				Type:         schema.TypeInt,
				Description:  "The ID of the site associated with the virtual machine.",
				Optional:     true,
				AtLeastOneOf: []string{"site_id", "cluster_id"},
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the virtual machine.",
				Optional:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the virtual machine.",
				Optional:    true,
			},
			"memory_mb": {
				Type:        schema.TypeInt,
				Description: "The memory mb of the virtual machine.",
				Optional:    true,
			},
			"vcpus": {
				Type:        schema.TypeFloat,
				Description: "The vcpus of the virtual machine.",
				Optional:    true,
			},
			"disk_size_mb": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The disk size in MB. When virtual disks are attached to this VM, NetBox automatically computes this as the aggregate of those disks and rejects manual values. In that case, omit this field from the configuration and let it be computed.",
			},
			"status": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "active",
				Description: buildValidValueDescription(resourceNetboxVirtualMachineStatusOptions),
			},
			tagsKey: tagsSchema,
			"primary_ipv4": {
				Type:        schema.TypeInt,
				Description: "The primary ipv4 of the virtual machine.",
				Computed:    true,
			},
			"primary_ipv6": {
				Type:        schema.TypeInt,
				Description: "The primary ipv6 of the virtual machine.",
				Computed:    true,
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
		SchemaVersion: 2,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceNetboxVirtualMachineResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceNetboxVirtualMachineStateUpgradeV0,
				Version: 0,
			},
			{
				Type:    resourceNetboxVirtualMachineResourceV1().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceNetboxVirtualMachineStateUpgradeV1,
				Version: 1,
			},
		},
	}
}

type virtualMachineResponse struct {
	Name             string                 `json:"name"`
	Cluster          *netboxapi.Ref         `json:"cluster"`
	PrimaryIp4       *netboxapi.Ref         `json:"primary_ip4"`
	PrimaryIp6       *netboxapi.Ref         `json:"primary_ip6"`
	Tenant           *netboxapi.Ref         `json:"tenant"`
	Device           *netboxapi.Ref         `json:"device"`
	Role             *netboxapi.Ref         `json:"role"`
	Platform         *netboxapi.Ref         `json:"platform"`
	Site             *netboxapi.Ref         `json:"site"`
	LocalContextData interface{}            `json:"local_context_data"`
	Comments         string                 `json:"comments"`
	Description      string                 `json:"description"`
	Vcpus            *float64               `json:"vcpus"`
	Memory           *int64                 `json:"memory"`
	Disk             *int64                 `json:"disk"`
	Status           *netboxapi.Choice      `json:"status"`
	Tags             []netboxapi.TagRef     `json:"tags"`
	CustomFields     map[string]interface{} `json:"custom_fields"`
}

// virtualMachineBody builds the request for create and update. References that
// are not set are sent as null, same as before.
func virtualMachineBody(api *providerState, d *schema.ResourceData, isUpdate bool) (map[string]interface{}, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}

	body := map[string]interface{}{
		"name":        d.Get("name").(string),
		"cluster":     nullableID(d.Get("cluster_id").(int)),
		"site":        nullableID(d.Get("site_id").(int)),
		"tenant":      nullableID(d.Get("tenant_id").(int)),
		"device":      nullableID(d.Get("device_id").(int)),
		"platform":    nullableID(d.Get("platform_id").(int)),
		"role":        nullableID(d.Get("role_id").(int)),
		"vcpus":       getOptionalFloat(d, "vcpus"),
		"memory":      getOptionalInt(d, "memory_mb"),
		"disk":        getOptionalInt(d, "disk_size_mb"),
		"primary_ip4": nil,
		"primary_ip6": nil,
		"tags":        tags,
	}

	if localContext, ok := d.GetOk("local_context_data"); ok {
		var jsonObj any
		if err := json.Unmarshal([]byte(localContext.(string)), &jsonObj); err == nil {
			body["local_context_data"] = jsonObj
		}
	}
	if cf, ok := managedCustomFields(d); ok {
		body["custom_fields"] = cf
	}

	// an empty comment or description that changed is sent as a space, to clear it
	if comments := getOptionalStr(d, "comments", true); comments != "" {
		body["comments"] = comments
	}
	if description := getOptionalStr(d, "description", true); description != "" {
		body["description"] = description
	}
	if status := d.Get("status").(string); status != "" {
		body["status"] = status
	}

	if isUpdate {
		body["primary_ip4"] = getOptionalInt(d, "primary_ipv4")
		body["primary_ip6"] = getOptionalInt(d, "primary_ipv6")

		// Only send disk_size_mb when explicitly set in config. When virtual disks
		// are attached, NetBox computes this as their aggregate and rejects any
		// value that does not match, including stale prior-state values.
		if d.GetRawConfig().GetAttr("disk_size_mb").IsNull() {
			body["disk"] = nil
		} else {
			body["disk"] = int64(d.Get("disk_size_mb").(int))
		}
	}

	return body, nil
}

func resourceNetboxVirtualMachineCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	body, err := virtualMachineBody(m.(*providerState), d, false)
	if err != nil {
		return diag.FromErr(err)
	}
	return createObjectCtx(ctx, d, m, virtualMachineAPIPath, body, resourceNetboxVirtualMachineRead)
}

func resourceNetboxVirtualMachineRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*providerState)
	return readObjectCtx(ctx, d, m, virtualMachineAPIPath, func(vm *virtualMachineResponse) {
		d.Set("name", vm.Name)
		d.Set("cluster_id", netboxapi.RefID(vm.Cluster))
		d.Set("primary_ipv4", netboxapi.RefID(vm.PrimaryIp4))
		d.Set("primary_ipv6", netboxapi.RefID(vm.PrimaryIp6))
		d.Set("tenant_id", netboxapi.RefID(vm.Tenant))
		d.Set("device_id", netboxapi.RefID(vm.Device))
		d.Set("role_id", netboxapi.RefID(vm.Role))
		d.Set("platform_id", netboxapi.RefID(vm.Platform))
		d.Set("site_id", netboxapi.RefID(vm.Site))

		if vm.LocalContextData != nil {
			if jsonArr, err := json.Marshal(vm.LocalContextData); err == nil {
				d.Set("local_context_data", string(jsonArr))
			}
		} else {
			d.Set("local_context_data", nil)
		}

		d.Set("comments", vm.Comments)
		d.Set("description", vm.Description)
		d.Set("vcpus", vm.Vcpus)
		d.Set("memory_mb", vm.Memory)
		d.Set("disk_size_mb", vm.Disk)
		if vm.Status != nil {
			d.Set("status", vm.Status.Value)
		} else {
			d.Set("status", nil)
		}

		api.readTagRefs(d, vm.Tags)
		d.Set(customFieldsKey, getCustomFields(vm.CustomFields))
	})
}

func resourceNetboxVirtualMachineUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	body, err := virtualMachineBody(m.(*providerState), d, true)
	if err != nil {
		return diag.FromErr(err)
	}
	return putObjectCtx(ctx, d, m, virtualMachineAPIPath, body, resourceNetboxVirtualMachineRead)
}

func resourceNetboxVirtualMachineDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return deleteObjectCtx(ctx, d, m, virtualMachineAPIPath)
}
