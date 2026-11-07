package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxUserCreate,
		Read:   resourceNetboxUserRead,
		Update: resourceNetboxUserUpdate,
		Delete: resourceNetboxUserDelete,

		Description: `:meta:subcategory:Authentication:This resource is used to manage users.`,

		Schema: map[string]*schema.Schema{
			"username": {
				Type:     schema.TypeString,
				Required: true,
			},
			"password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"email": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"first_name": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"last_name": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"active": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"staff": {
				Type:       schema.TypeBool,
				Optional:   true,
				Default:    false,
				Deprecated: "The staff field has been removed in NetBox v4.5 (migration users.0013_user_remove_is_staff). This attribute is ignored by NetBox v4.5 and later.",
			},
			"group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const userAPIPath = "users/users/"

type userRequest struct {
	Username  string  `json:"username"`
	Password  string  `json:"password"`
	Email     string  `json:"email,omitempty"`
	FirstName string  `json:"first_name,omitempty"`
	LastName  string  `json:"last_name,omitempty"`
	IsActive  bool    `json:"is_active"`
	IsStaff   bool    `json:"is_staff"`
	Groups    []int64 `json:"groups"`
}

type userResponse struct {
	Username  string          `json:"username"`
	Email     string          `json:"email"`
	FirstName string          `json:"first_name"`
	LastName  string          `json:"last_name"`
	IsActive  bool            `json:"is_active"`
	IsStaff   bool            `json:"is_staff"`
	Groups    []netboxapi.Ref `json:"groups"`
}

func userRequestFromResourceData(d *schema.ResourceData) *userRequest {
	return &userRequest{
		Username:  d.Get("username").(string),
		Password:  d.Get("password").(string),
		Email:     d.Get("email").(string),
		FirstName: d.Get("first_name").(string),
		LastName:  d.Get("last_name").(string),
		IsActive:  d.Get("active").(bool),
		IsStaff:   d.Get("staff").(bool),
		Groups:    toInt64List(d.Get("group_ids")),
	}
}

func resourceNetboxUserCreate(d *schema.ResourceData, m interface{}) error {
	return createObject(d, m, userAPIPath, userRequestFromResourceData(d), resourceNetboxUserRead)
}

func resourceNetboxUserRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, userAPIPath, func(user *userResponse) {
		d.Set("username", user.Username)
		d.Set("email", user.Email)
		d.Set("first_name", user.FirstName)
		d.Set("last_name", user.LastName)
		d.Set("staff", user.IsStaff)
		d.Set("active", user.IsActive)
		d.Set("group_ids", netboxapi.IDs(user.Groups))
		// Passwords cannot be set and not read
	})
}

func resourceNetboxUserUpdate(d *schema.ResourceData, m interface{}) error {
	return putObject(d, m, userAPIPath, userRequestFromResourceData(d), resourceNetboxUserRead)
}

func resourceNetboxUserDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, userAPIPath)
}
