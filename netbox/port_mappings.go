package netbox

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// A front port is mapped to positions of rear ports, and a rear port to
// positions of front ports. NetBox 4.5 and later describe this with a list of
// mappings. portKey is "rear_port" for the list on a front port and
// "front_port" for the list on a rear port.

// portMappingsSchema is the schema of a list of port mappings.
func portMappingsSchema(portKey string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Computed:    true,
		Description: "The ports and positions this port is mapped to. Needs NetBox 4.5 or newer.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"position": {
					Type:         schema.TypeInt,
					Required:     true,
					ValidateFunc: validation.IntBetween(1, 1024),
					Description:  "Position on this port.",
				},
				portKey + "_id": {
					Type:     schema.TypeInt,
					Required: true,
				},
				portKey + "_position": {
					Type:         schema.TypeInt,
					Optional:     true,
					Default:      1,
					ValidateFunc: validation.IntBetween(1, 1024),
					Description:  "Position on the other port.",
				},
			},
		},
	}
}

// portMappingsRequest returns the mappings of a list attribute for a request.
// It returns nil when the list is not configured, so that the field is left out.
func portMappingsRequest(d *schema.ResourceData, key, portKey string) []map[string]int64 {
	list, ok := d.Get(key).([]interface{})
	if !ok || len(list) == 0 {
		return nil
	}
	out := make([]map[string]int64, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		out = append(out, map[string]int64{
			"position":            int64(m["position"].(int)),
			portKey:               int64(m[portKey+"_id"].(int)),
			portKey + "_position": int64(m[portKey+"_position"].(int)),
		})
	}
	return out
}

// portMappingsState turns the mappings of a response into the value of the list
// attribute.
func portMappingsState(mappings []map[string]interface{}, portKey string) []interface{} {
	out := make([]interface{}, 0, len(mappings))
	for _, m := range mappings {
		out = append(out, map[string]interface{}{
			"position":            int(jInt(m, "position")),
			portKey + "_id":       int(jInt(m, portKey)),
			portKey + "_position": int(jInt(m, portKey+"_position")),
		})
	}
	return out
}
