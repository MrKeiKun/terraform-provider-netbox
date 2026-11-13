package fwprovider

import (
	"context"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwresource"
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/MrKeiKun/terraform-provider-netbox/internal/providerstate"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func init() {
	manualResources = append(manualResources, device())
}

func boolValue(v fwresource.Values, name string) bool {
	if b, ok := v[name].(types.Bool); ok && !b.IsNull() && !b.IsUnknown() {
		return b.ValueBool()
	}
	return false
}

// setVirtualChassisMaster sets or clears (master == nil) the master of a virtual chassis.
func setVirtualChassisMaster(ctx context.Context, st *providerstate.State, chassis int64, master *int64) error {
	var value any
	if master != nil {
		value = *master
	}
	_, err := netboxapi.Patch[netboxapi.Ref](ctx, st.Rest, "dcim/virtual-chassis/", chassis, map[string]any{"master": value})
	return err
}

func device() *fwresource.Spec {
	return &fwresource.Spec{
		TypeName:     "netbox_device",
		Path:         "dcim/devices/",
		Description:  ":meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/features/devices/#devices):\n> Every piece of hardware which is installed within a site or rack exists in NetBox as a device. Devices are measured in rack units (U) and can be half depth or full depth. A device may have a height of 0U: These devices do not consume vertical rack space and cannot be assigned to a particular rack unit. A common example of a 0U device is a vertically-mounted PDU.",
		Tags:         true,
		CustomFields: true,
		Put:          true,
		RequiredWith: [][]string{{"rack_face", "rack_position"}, {"virtual_chassis_id", "virtual_chassis_master"}, {"virtual_chassis_master", "virtual_chassis_id"}},
		Attrs: []fwresource.Attr{
			{Name: "airflow", Kind: fwresource.String, Nullable: true, Optional: true, Description: "Airflow direction of the device. One of [front-to-rear, rear-to-front, left-to-right, right-to-left, side-to-rear, rear-to-side, bottom-to-top, top-to-bottom, passive, mixed]"},
			{Name: "asset_tag", Kind: fwresource.String, Nullable: true, Optional: true, Description: "The asset tag of the device."},
			{Name: "cluster_id", Kind: fwresource.Int, Key: "cluster", Ref: true, Nullable: true, Optional: true, Description: "The ID of the cluster associated with the device."},
			{Name: "comments", Kind: fwresource.String, Optional: true, Description: "Free-form comments about the device."},
			{Name: "config_template_id", Kind: fwresource.Int, Key: "config_template", Ref: true, Nullable: true, Optional: true, Description: "The ID of the config template associated with the device."},
			{Name: "description", Kind: fwresource.String, Optional: true, Description: "A description of the device."},
			{Name: "device_type_id", Kind: fwresource.Int, Key: "device_type", Ref: true, Required: true, Description: "The ID of the device type associated with the device."},
			{Name: "local_context_data", Kind: fwresource.JSON, Optional: true, Description: "This is best managed through the use of `jsonencode` and a map of settings."},
			{Name: "location_id", Kind: fwresource.Int, Key: "location", Ref: true, Nullable: true, Optional: true, Computed: true, Description: "The ID of the location associated with the device. NetBox takes it from the rack when it is not set."},
			{Name: "name", Kind: fwresource.String, Required: true, Description: "The name of the device."},
			{Name: "platform_id", Kind: fwresource.Int, Key: "platform", Ref: true, Nullable: true, Optional: true, Description: "The ID of the platform associated with the device."},
			{Name: "primary_ipv4", Kind: fwresource.Int, Key: "primary_ip4", Ref: true, Computed: true, Description: "The primary ipv4 of the device."},
			{Name: "primary_ipv6", Kind: fwresource.Int, Key: "primary_ip6", Ref: true, Computed: true, Description: "The primary ipv6 of the device."},
			{Name: "rack_face", Kind: fwresource.String, Key: "face", Nullable: true, Optional: true, Description: "Valid values are `front` and `rear`.", OneOf: []string{"front", "rear"}},
			{Name: "rack_id", Kind: fwresource.Int, Key: "rack", Ref: true, Nullable: true, Optional: true, Description: "The ID of the rack associated with the device."},
			{Name: "rack_position", Kind: fwresource.Float, Key: "position", Optional: true, Description: "The rack position of the device."},
			{Name: "role_id", Kind: fwresource.Int, Key: "role", Ref: true, Required: true, Description: "The ID of the role associated with the device."},
			{Name: "serial", Kind: fwresource.String, Optional: true, Description: "The serial number of the device."},
			{Name: "site_id", Kind: fwresource.Int, Key: "site", Ref: true, Required: true, Description: "The ID of the site associated with the device."},
			{Name: "status", Kind: fwresource.String, Optional: true, Default: "active", Description: "Valid values are `offline`, `active`, `planned`, `staged`, `failed`, `inventory` and `decommissioning`.", OneOf: []string{"offline", "active", "planned", "staged", "failed", "inventory", "decommissioning"}},
			{Name: "tenant_id", Kind: fwresource.Int, Key: "tenant", Ref: true, Nullable: true, Optional: true, Description: "The ID of the tenant associated with the device."},
			{Name: "virtual_chassis_id", Kind: fwresource.Int, Key: "virtual_chassis", Ref: true, Nullable: true, Optional: true, Description: "The ID of the virtual chassis associated with the device."},
			{Name: "virtual_chassis_master", Kind: fwresource.Bool, NoSend: true, Optional: true, Description: "Whether virtual chassis master is enabled for the device."},
			{Name: "virtual_chassis_position", Kind: fwresource.Int, Key: "vc_position", Nullable: true, Optional: true, Description: "The virtual chassis position of the device."},
			{Name: "virtual_chassis_priority", Kind: fwresource.Int, Key: "vc_priority", Nullable: true, Optional: true, Description: "The virtual chassis priority of the device."},
		},
		ModifyPlan: func(config, plan fwresource.Values) map[string]attr.Value {
			// NetBox takes the location from the rack. Without a rack and a location in the
			// configuration the device has no location.
			if config["location_id"].IsNull() && config["rack_id"].IsNull() {
				return map[string]attr.Value{"location_id": types.Int64Null()}
			}
			return nil
		},
		ModifyState: func(_ context.Context, obj map[string]any, _ fwresource.Values, out fwresource.Values) error {
			master := false
			if vc, ok := obj["virtual_chassis"].(map[string]any); ok {
				if m, ok := vc["master"].(map[string]any); ok {
					id, _ := m["id"].(float64)
					devID, _ := obj["id"].(float64)
					master = id != 0 && id == devID
				}
			}
			out["virtual_chassis_master"] = types.BoolValue(master)
			return nil
		},
		AfterWrite: func(ctx context.Context, st *providerstate.State, id int64, plan, prior fwresource.Values) (bool, error) {
			chassis := intValue(plan, "virtual_chassis_id")
			if chassis == 0 {
				return false, nil
			}
			master := boolValue(plan, "virtual_chassis_master")
			if prior == nil && !master {
				return false, nil
			}
			if prior != nil && boolValue(prior, "virtual_chassis_master") == master {
				return false, nil
			}
			if master {
				return true, setVirtualChassisMaster(ctx, st, chassis, &id)
			}
			// It was set before, but no longer set, remove it as master
			return true, setVirtualChassisMaster(ctx, st, chassis, nil)
		},
		BeforeDelete: func(ctx context.Context, st *providerstate.State, _ int64, state fwresource.Values) error {
			// A device that is the master of its virtual chassis cannot be deleted directly. It has to
			// stop being the master first.
			if chassis := intValue(state, "virtual_chassis_id"); chassis != 0 && boolValue(state, "virtual_chassis_master") {
				return setVirtualChassisMaster(ctx, st, chassis, nil)
			}
			return nil
		},
	}
}
