package netbox

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var genericObjectSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"object_type": {
			Type:     schema.TypeString,
			Required: true,
			ValidateFunc: validation.StringInSlice([]string{
				// known supported generic objects (wherever the schema is used)
				// can be made into a parameter when importing this schema if desired
				"dcim.powerport",
				"dcim.poweroutlet",
				"dcim.powerfeed",
				"dcim.frontport",
				"dcim.rearport",
				"dcim.consoleserverport",
				"dcim.consoleport",
				"dcim.interface",
				"circuits.circuittermination",
			}, false),
		},
		"object_id": {
			Type:     schema.TypeInt,
			Required: true,
		},
	},
}

// genericObject points at an object of any type, for example the end of a cable.
type genericObject struct {
	ObjectID   int64  `json:"object_id"`
	ObjectType string `json:"object_type"`
}

func getGenericObjectsFromSchemaSet(schemaSet *schema.Set) []genericObject {
	retArr := make([]genericObject, 0, schemaSet.Len())
	for _, i := range schemaSet.List() {
		retArr = append(retArr, genericObject{
			ObjectID:   int64(i.(map[string]interface{})["object_id"].(int)),
			ObjectType: i.(map[string]interface{})["object_type"].(string),
		})
	}
	return retArr
}

func getSchemaSetFromGenericObjects(objects []genericObject) []map[string]interface{} {
	retArr := make([]map[string]interface{}, 0, len(objects))
	for _, obj := range objects {
		retArr = append(retArr, map[string]interface{}{
			"object_type": obj.ObjectType,
			"object_id":   obj.ObjectID,
		})
	}
	return retArr
}
