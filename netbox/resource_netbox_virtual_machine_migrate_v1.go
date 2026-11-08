package netbox

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxVirtualMachineResourceV1() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the virtual machine migrate v1.",
				Required:    true,
			},
			"cluster_id": {
				Type:         schema.TypeInt,
				Description:  "The ID of the cluster associated with the virtual machine migrate v1.",
				Optional:     true,
				AtLeastOneOf: []string{"site_id", "cluster_id"},
			},
			"tenant_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the tenant associated with the virtual machine migrate v1.",
				Optional:    true,
			},
			"device_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the device associated with the virtual machine migrate v1.",
				Optional:    true,
			},
			"platform_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the platform associated with the virtual machine migrate v1.",
				Optional:    true,
			},
			"role_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the role associated with the virtual machine migrate v1.",
				Optional:    true,
			},
			"site_id": {
				Type:         schema.TypeInt,
				Description:  "The ID of the site associated with the virtual machine migrate v1.",
				Optional:     true,
				AtLeastOneOf: []string{"site_id", "cluster_id"},
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the virtual machine migrate v1.",
				Optional:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the virtual machine migrate v1.",
				Optional:    true,
			},
			"memory_mb": {
				Type:        schema.TypeInt,
				Description: "The memory mb of the virtual machine migrate v1.",
				Optional:    true,
			},
			"vcpus": {
				Type:        schema.TypeFloat,
				Description: "The vcpus of the virtual machine migrate v1.",
				Optional:    true,
			},
			"disk_size_gb": {
				Type:        schema.TypeInt,
				Description: "The disk size gb of the virtual machine migrate v1.",
				Optional:    true,
				Computed:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Description: "The status of the virtual machine migrate v1.",
				Optional:    true,
				Default:     "active",
			},
			tagsKey: tagsSchema,
			"primary_ipv4": {
				Type:        schema.TypeInt,
				Description: "The primary ipv4 of the virtual machine migrate v1.",
				Computed:    true,
			},
			"primary_ipv6": {
				Type:        schema.TypeInt,
				Description: "The primary ipv6 of the virtual machine migrate v1.",
				Computed:    true,
			},
			"local_context_data": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "This is best managed through the use of `jsonencode` and a map of settings.",
			},
			customFieldsKey: customFieldsSchema,
		},
	}
}
func resourceNetboxVirtualMachineStateUpgradeV1(_ context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	v, ok := rawState["disk_size_gb"].(float64)
	if !ok {
		log.Printf("[DEBUG] disk size before migration isnt float64: %#v\n but a %T", rawState["disk_size_gb"], rawState["disk_size_gb"])
		rawState["disk_size_mb"] = float64(0)
		delete(rawState, "disk_size_gb")
		return rawState, nil
	}

	log.Printf("[DEBUG] disk size in GB before migration: %#v\n", rawState["disk_size_gb"])

	// set new disk size
	rawState["disk_size_mb"] = v * 1000
	log.Printf("[DEBUG] disk size in MB after migration: %#v\n", rawState["disk_size_mb"])

	delete(rawState, "disk_size_gb")

	return rawState, nil
}
