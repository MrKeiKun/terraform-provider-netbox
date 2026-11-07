package netbox

import (
	"encoding/json"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxPermission() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxPermissionCreate,
		Read:   resourceNetboxPermissionRead,
		Update: resourceNetboxPermissionUpdate,
		Delete: resourceNetboxPermissionDelete,
		Description: `:meta:subcategory:Authentication:This resource manages the object-based permissions for Netbox users, built into the application.

> Object-based permissions enable an administrator to grant users or groups the ability to perform an action on arbitrary subsets of objects in NetBox, rather than all objects of a certain type.
> For more information, see the [Netbox Object-Based Permissions Docs.](https://docs.netbox.dev/en/stable/administration/permissions/)`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the permission object.",
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "The description of the permission object.",
				Optional:    true,
			},
			"enabled": {
				Type:        schema.TypeBool,
				Description: "Whether the permission object is enabled or not.",
				Optional:    true,
				Default:     true,
			},
			"object_types": {
				Type: schema.TypeSet,
				Description: "A list of object types that the permission object allows access to. Should be in a form " +
					"the API can accept. For example: `circuits.provider`, `dcim.inventoryitem`, etc.",
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"groups": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "A list of group IDs that have been assigned to this permission object.",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"users": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "A list of user IDs that have been assigned to this permission object.",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"actions": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "A list actions that are allowed on the object types. Acceptable values are `view`, `add`, `change`, or `delete`.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"constraints": {
				Type: schema.TypeString,
				Description: "A JSON string of an arbitrary filter used to limit the granted action(s) to a specific subset of objects. " +
					"For more information on correct syntax, see https://docs.netbox.dev/en/stable/administration/permissions/#constraints ",
				Optional:     true,
				ValidateFunc: validation.StringIsJSON,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const permissionAPIPath = "users/permissions/"

type permissionRequest struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Enabled     bool        `json:"enabled"`
	ObjectTypes []string    `json:"object_types"`
	Groups      []int64     `json:"groups"`
	Users       []int64     `json:"users"`
	Actions     []string    `json:"actions"`
	Constraints interface{} `json:"constraints"`
}

type permissionResponse struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Enabled     bool            `json:"enabled"`
	ObjectTypes []string        `json:"object_types"`
	Groups      []netboxapi.Ref `json:"groups"`
	Users       []netboxapi.Ref `json:"users"`
	Actions     []string        `json:"actions"`
	Constraints interface{}     `json:"constraints"`
}

func permissionRequestFromResourceData(d *schema.ResourceData) (*permissionRequest, error) {
	data := &permissionRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Enabled:     d.Get("enabled").(bool),
		ObjectTypes: toStringList(d.Get("object_types")),
		Groups:      toInt64List(d.Get("groups")),
		Users:       toInt64List(d.Get("users")),
		Actions:     toStringList(d.Get("actions")),
	}

	if c := d.Get("constraints").(string); c != "" {
		var constraints interface{}
		if err := json.Unmarshal([]byte(c), &constraints); err != nil {
			return nil, err
		}
		switch v := constraints.(type) {
		case []interface{}:
			data.Constraints = v
		case map[string]interface{}:
			data.Constraints = v
		}
	}
	return data, nil
}

func resourceNetboxPermissionCreate(d *schema.ResourceData, m interface{}) error {
	data, err := permissionRequestFromResourceData(d)
	if err != nil {
		return err
	}
	return createObject(d, m, permissionAPIPath, data, resourceNetboxPermissionRead)
}

func resourceNetboxPermissionRead(d *schema.ResourceData, m interface{}) error {
	var readErr error
	err := readObject(d, m, permissionAPIPath, func(permission *permissionResponse) {
		d.Set("name", permission.Name)
		d.Set("description", permission.Description)
		d.Set("enabled", permission.Enabled)
		d.Set("object_types", permission.ObjectTypes)

		var groups []int
		for _, v := range permission.Groups {
			groups = append(groups, int(v.ID))
		}
		d.Set("groups", groups)

		var users []int
		for _, v := range permission.Users {
			users = append(users, int(v.ID))
		}
		d.Set("users", users)

		d.Set("actions", permission.Actions)

		if permission.Constraints == nil {
			d.Set("constraints", "")
			return
		}

		b, err := json.Marshal(permission.Constraints)
		if err != nil {
			readErr = err
			return
		}
		d.Set("constraints", string(b))
	})
	if err != nil {
		return err
	}
	return readErr
}

func resourceNetboxPermissionUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := permissionRequestFromResourceData(d)
	if err != nil {
		return err
	}
	return putObject(d, m, permissionAPIPath, data, resourceNetboxPermissionRead)
}

func resourceNetboxPermissionDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, permissionAPIPath)
}
