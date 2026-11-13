package fwprovider

import (
	"context"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwresource"
)

func init() {
	standardResources = append(standardResources, extrasResources()...)
}

// extrasResources are the objects that belong to a NetBox user: bookmarks, notifications,
// subscriptions and table configurations.
func extrasResources() []*fwresource.Spec {
	return []*fwresource.Spec{
		{
			TypeName:    "netbox_bookmark",
			Path:        "extras/bookmarks/",
			Description: ":meta:subcategory:Extras:A bookmark marks an object for a user, so that the user finds it again on the dashboard.",
			Attrs: []fwresource.Attr{
				{Name: "object_id", Kind: fwresource.Int, Required: true, ForceNew: true, Description: "The ID of the bookmarked object."},
				{Name: "object_type", Kind: fwresource.String, Required: true, ForceNew: true, Description: "The object type of the bookmarked object, for example `dcim.site`."},
				{Name: "user_id", Kind: fwresource.Int, Key: "user", Ref: true, Required: true, ForceNew: true, Description: "The ID of the user who has the bookmark."},
			},
		},
		{
			TypeName:    "netbox_notification",
			Path:        "extras/notifications/",
			Description: ":meta:subcategory:Extras:A notification tells a user about an event of an object.",
			Attrs: []fwresource.Attr{
				{Name: "event_type", Kind: fwresource.String, Required: true, Description: "The kind of event. One of [object_created, object_updated, object_deleted, job_started, job_completed, job_failed, job_errored].", OneOf: []string{"job_completed", "job_errored", "job_failed", "job_started", "object_created", "object_deleted", "object_updated"}},
				{Name: "object_id", Kind: fwresource.Int, Required: true, ForceNew: true, Description: "The ID of the object the notification is about."},
				{Name: "object_type", Kind: fwresource.String, Required: true, ForceNew: true, Description: "The object type of the object the notification is about, for example `dcim.site`."},
				{Name: "read", Kind: fwresource.String, Nullable: true, Optional: true, Time: true, Description: "When the user read the notification, as a timestamp. Not set while it is unread."},
				{Name: "user_id", Kind: fwresource.Int, Key: "user", Ref: true, Required: true, ForceNew: true, Description: "The ID of the user who gets the notification."},
			},
		},
		{
			TypeName:    "netbox_subscription",
			Path:        "extras/subscriptions/",
			Description: ":meta:subcategory:Extras:A subscription makes NetBox notify a user about changes of an object.",
			Attrs: []fwresource.Attr{
				{Name: "object_id", Kind: fwresource.Int, Required: true, ForceNew: true, Description: "The ID of the object the user subscribes to."},
				{Name: "object_type", Kind: fwresource.String, Required: true, ForceNew: true, Description: "The object type of the object the user subscribes to, for example `dcim.site`."},
				{Name: "user_id", Kind: fwresource.Int, Key: "user", Ref: true, Required: true, ForceNew: true, Description: "The ID of the user who subscribes."},
			},
		},
		{
			TypeName:    "netbox_table_config",
			Path:        "extras/table-configs/",
			Description: ":meta:subcategory:Extras:A table configuration chooses the columns and the order of the rows of a table in the NetBox web interface.",
			// NetBox 4.4 fails with an error when a table configuration is created without an ordering.
			ModifyBody: func(_ context.Context, _, _ fwresource.Values, body map[string]any) error {
				if _, ok := body["ordering"]; !ok {
					body["ordering"] = []any{}
				}
				return nil
			},
			Attrs: []fwresource.Attr{
				{Name: "columns", Kind: fwresource.StringList, Required: true, Description: "The columns of the table, in the order they are shown."},
				{Name: "description", Kind: fwresource.String, Optional: true, Description: "A description of the table configuration.", MaxLen: 200},
				{Name: "enabled", Kind: fwresource.Bool, Optional: true, Computed: true, Description: "Whether the table configuration is in use. NetBox turns it on by default."},
				{Name: "name", Kind: fwresource.String, Required: true, Description: "The name of the table configuration.", MaxLen: 100},
				{Name: "object_type", Kind: fwresource.String, Required: true, ForceNew: true, Description: "The object type of the table, for example `dcim.site`."},
				{Name: "ordering", Kind: fwresource.StringList, Optional: true, Description: "The columns the rows are sorted by. A leading minus sign sorts in descending order."},
				{Name: "shared", Kind: fwresource.Bool, Optional: true, Computed: true, Description: "Whether other users can use the table configuration. NetBox turns it on by default."},
				{Name: "table", Kind: fwresource.String, Required: true, ForceNew: true, Description: "The name of the table, for example `SiteTable`.", MaxLen: 100},
				{Name: "user_id", Kind: fwresource.Int, Key: "user", Ref: true, Nullable: true, Optional: true, Description: "The ID of the user who owns the table configuration."},
				{Name: "weight", Kind: fwresource.Int, Optional: true, Computed: true, Description: "The weight of the table configuration. Heavier ones are used last.", Min: i64(0), Max: i64(32767)},
			},
		},
	}
}
