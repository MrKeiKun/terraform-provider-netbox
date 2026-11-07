package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxNotificationGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxNotificationGroupCreate,
		Read:   resourceNetboxNotificationGroupRead,
		Update: resourceNetboxNotificationGroupUpdate,
		Delete: resourceNetboxNotificationGroupDelete,

		Description: `:meta:subcategory:Extras:From the [official documentation](https://docs.netbox.dev/en/stable/models/features/notifications/):

> A notification group is a set of users and groups that receive the same notifications.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
			"user_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const notificationGroupAPIPath = "extras/notification-groups/"

type notificationGroupRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Groups      []int64 `json:"groups"`
	Users       []int64 `json:"users"`
}

type notificationGroupResponse struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Groups      []netboxapi.Ref `json:"groups"`
	Users       []netboxapi.Ref `json:"users"`
}

func notificationGroupRequestFromResourceData(api *providerState, d *schema.ResourceData) (*notificationGroupRequest, error) {
	data := &notificationGroupRequest{
		Name:        d.Get("name").(string),
		Description: getOptionalStr(d, "description", true),
		Groups:      getIDSet(d, "group_ids"),
		Users:       getIDSet(d, "user_ids"),
	}
	return data, nil
}

func resourceNetboxNotificationGroupCreate(d *schema.ResourceData, m interface{}) error {
	data, err := notificationGroupRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, notificationGroupAPIPath, data, resourceNetboxNotificationGroupRead)
}

func resourceNetboxNotificationGroupRead(d *schema.ResourceData, m interface{}) error {
	return readObject(d, m, notificationGroupAPIPath, func(r *notificationGroupResponse) {
		d.Set("name", r.Name)
		d.Set("description", r.Description)
		d.Set("group_ids", netboxapi.IDs(r.Groups))
		d.Set("user_ids", netboxapi.IDs(r.Users))
	})
}

func resourceNetboxNotificationGroupUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := notificationGroupRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, notificationGroupAPIPath, data, resourceNetboxNotificationGroupRead)
}

func resourceNetboxNotificationGroupDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, notificationGroupAPIPath)
}
