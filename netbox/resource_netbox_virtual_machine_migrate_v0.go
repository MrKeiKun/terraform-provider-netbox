package netbox

import (
	"context"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxVirtualMachineResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the virtual machine migrate v0.",
				Required:    true,
			},
			"cluster_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the cluster associated with the virtual machine migrate v0.",
				Required:    true,
			},
			"tenant_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the tenant associated with the virtual machine migrate v0.",
				Optional:    true,
			},
			"platform_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the platform associated with the virtual machine migrate v0.",
				Optional:    true,
			},
			"role_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the role associated with the virtual machine migrate v0.",
				Optional:    true,
			},
			"site_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the site associated with the virtual machine migrate v0.",
				Computed:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the virtual machine migrate v0.",
				Optional:    true,
			},
			"memory_mb": {
				Type:        schema.TypeInt,
				Description: "The memory mb of the virtual machine migrate v0.",
				Optional:    true,
			},
			"vcpus": {
				Type:        schema.TypeString,
				Description: "The vcpus of the virtual machine migrate v0.",
				Optional:    true,
			},
			"disk_size_gb": {
				Type:        schema.TypeInt,
				Description: "The disk size gb of the virtual machine migrate v0.",
				Optional:    true,
			},
			tagsKey: tagsSchema,
			"primary_ipv4": {
				Type:        schema.TypeInt,
				Description: "The primary ipv4 of the virtual machine migrate v0.",
				Computed:    true,
			},
		},
	}
}
func resourceNetboxVirtualMachineStateUpgradeV0(_ context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	v, ok := rawState["vcpus"]
	if !ok {
		return rawState, nil
	}

	s, ok := v.(string)
	if !ok {
		// since the provider was already released without this state migration, we have to accept that this field already contains non-string content
		return rawState, nil
	}

	log.Printf("[DEBUG] vcpus before migration: %#v", rawState["vcpus"])

	f, err := strconv.ParseFloat(s, 64)
	if err == nil {
		rawState["vcpus"] = f
	} else {
		rawState["vcpus"] = float64(0)
		log.Printf("[DEBUG] Schema upgrade: vcpus has been migrated to %g", f)
	}

	log.Printf("[DEBUG] vcpus after migration: %#v", rawState["vcpus"])
	return rawState, nil
}
