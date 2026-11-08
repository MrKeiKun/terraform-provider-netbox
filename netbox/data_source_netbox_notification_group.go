package netbox

import (
	"context"
	"errors"
	"net/url"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetboxNotificationGroup() *schema.Resource {
	return &schema.Resource{
		Read:        dataSourceNetboxNotificationGroupRead,
		Description: `:meta:subcategory:Extras:Retrieves a single notification group from NetBox. The lookup fails if no match or more than one match is found.`,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Description:  "The name of the notification group.",
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"name"},
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the notification group.",
				Computed:    true,
			},
			"group_ids": {
				Type:        schema.TypeSet,
				Description: "The IDs of the group associated with the notification group.",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"user_ids": {
				Type:        schema.TypeSet,
				Description: "The IDs of the user associated with the notification group.",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
		},
	}
}

func dataSourceNetboxNotificationGroupRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	name := d.Get("name").(string)

	// NetBox ignores the name filter for notification groups, so the search is
	// used to narrow the list down and the exact name is matched here.
	query := url.Values{}
	query.Set("q", name)
	groups, err := netboxapi.List[map[string]interface{}](context.Background(), api.rest, "extras/notification-groups/", query)
	if err != nil {
		return err
	}

	var matches []map[string]interface{}
	for _, g := range groups {
		if jStr(g, "name") == name {
			matches = append(matches, g)
		}
	}
	switch {
	case len(matches) > 1:
		return errors.New("more than one notification group returned, specify a more narrow filter")
	case len(matches) == 0:
		return errors.New("no notification group found matching filter")
	}
	r := matches[0]

	d.SetId(strconv.FormatInt(jInt(r, "id"), 10))
	d.Set("name", jStr(r, "name"))
	d.Set("description", jStr(r, "description"))
	d.Set("group_ids", jIDs(r, "groups"))
	d.Set("user_ids", jIDs(r, "users"))
	return nil
}
