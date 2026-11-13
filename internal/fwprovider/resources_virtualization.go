package fwprovider

import (
	"context"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwresource"
)

func init() {
	manualResources = append(manualResources, virtualDisk(), virtualMachine())
}

func virtualDisk() *fwresource.Spec {
	return &fwresource.Spec{
		TypeName:     "netbox_virtual_disk",
		Path:         "virtualization/virtual-disks/",
		Description:  ":meta:subcategory:Virtualization:From the [official documentation](https://docs.netbox.dev/en/stable/models/virtualization/virtualdisk/):\n> A virtual disk is used to model discrete virtual hard disks assigned to virtual machines.",
		Version:      1,
		Tags:         true,
		CustomFields: true,
		Put:          true,
		Upgraders:    map[int64]func(context.Context, map[string]any) (map[string]any, error){0: upgradeVirtualDiskV0},
		Attrs: []fwresource.Attr{
			{Name: "description", Kind: fwresource.String, Optional: true, Description: "A description of the virtual disk."},
			{Name: "name", Kind: fwresource.String, Required: true, Description: "The name of the virtual disk."},
			{Name: "owner_id", Kind: fwresource.Int, Key: "owner", Ref: true, Nullable: true, Optional: true, Description: "The ID of the owner associated with the virtual disk."},
			{Name: "size_mb", Kind: fwresource.Int, Key: "size", Required: true, Description: "The size mb of the virtual disk."},
			{Name: "virtual_machine_id", Kind: fwresource.Int, Key: "virtual_machine", Ref: true, Required: true, Description: "The ID of the virtual machine associated with the virtual disk."},
		},
	}
}

func virtualMachine() *fwresource.Spec {
	return &fwresource.Spec{
		TypeName:     "netbox_virtual_machine",
		Path:         "virtualization/virtual-machines/",
		Description:  ":meta:subcategory:Virtualization:From the [official documentation](https://docs.netbox.dev/en/stable/features/virtualization/#virtual-machines):\n> A virtual machine is a virtualized compute instance. These behave in NetBox very similarly to device objects, but without any physical attributes. For example, a VM may have interfaces assigned to it with IP addresses and VLANs, however its interfaces cannot be connected via cables (because they are virtual). Each VM may also define its compute, memory, and storage resources as well.",
		Version:      2,
		Tags:         true,
		CustomFields: true,
		Put:          true,
		AtLeastOneOf: [][]string{{"site_id", "cluster_id"}},
		Upgraders: map[int64]func(context.Context, map[string]any) (map[string]any, error){
			0: upgradeVirtualMachineV0,
			1: upgradeVirtualMachineV1,
		},
		Attrs: []fwresource.Attr{
			{Name: "cluster_id", Kind: fwresource.Int, Key: "cluster", Ref: true, Nullable: true, Optional: true, Description: "The ID of the cluster associated with the virtual machine. At least one of `site_id` or `cluster_id` must be given."},
			{Name: "comments", Kind: fwresource.String, Optional: true, Description: "Free-form comments about the virtual machine."},
			{Name: "description", Kind: fwresource.String, Optional: true, Description: "A description of the virtual machine."},
			{Name: "device_id", Kind: fwresource.Int, Key: "device", Ref: true, Nullable: true, Optional: true, Description: "The ID of the device associated with the virtual machine."},
			{Name: "disk_size_mb", Kind: fwresource.Int, Key: "disk", Optional: true, Computed: true, Volatile: true, Description: "The disk size in MB. When virtual disks are attached to this VM, NetBox automatically computes this as the aggregate of those disks and rejects manual values. In that case, omit this field from the configuration and let it be computed."},
			{Name: "local_context_data", Kind: fwresource.JSON, Optional: true, Description: "This is best managed through the use of `jsonencode` and a map of settings."},
			{Name: "memory_mb", Kind: fwresource.Int, Key: "memory", Nullable: true, Optional: true, Description: "The memory mb of the virtual machine."},
			{Name: "name", Kind: fwresource.String, Required: true, Description: "The name of the virtual machine."},
			{Name: "platform_id", Kind: fwresource.Int, Key: "platform", Ref: true, Nullable: true, Optional: true, Description: "The ID of the platform associated with the virtual machine."},
			{Name: "primary_ipv4", Kind: fwresource.Int, Key: "primary_ip4", Ref: true, Computed: true, Description: "The primary ipv4 of the virtual machine."},
			{Name: "primary_ipv6", Kind: fwresource.Int, Key: "primary_ip6", Ref: true, Computed: true, Description: "The primary ipv6 of the virtual machine."},
			{Name: "role_id", Kind: fwresource.Int, Key: "role", Ref: true, Nullable: true, Optional: true, Description: "The ID of the role associated with the virtual machine."},
			{Name: "site_id", Kind: fwresource.Int, Key: "site", Ref: true, Optional: true, Computed: true, Description: "The ID of the site associated with the virtual machine. NetBox takes it from the cluster when it is not set. At least one of `site_id` or `cluster_id` must be given."},
			{Name: "status", Kind: fwresource.String, Optional: true, Default: "active", Description: "Valid values are `offline`, `active`, `planned`, `staged`, `failed` and `decommissioning`. Defaults to `active`.", OneOf: []string{"offline", "active", "planned", "staged", "failed", "decommissioning"}},
			{Name: "tenant_id", Kind: fwresource.Int, Key: "tenant", Ref: true, Nullable: true, Optional: true, Description: "The ID of the tenant associated with the virtual machine."},
			{Name: "vcpus", Kind: fwresource.Float, Optional: true, Description: "The vcpus of the virtual machine."},
		},
		ModifyBody: func(_ context.Context, plan, prior fwresource.Values, body map[string]any) error {
			// A put clears the primary IPs, so they are sent as they are.
			for _, f := range [][2]string{{"primary_ip4", "primary_ipv4"}, {"primary_ip6", "primary_ipv6"}} {
				body[f[0]] = nil
				if prior != nil {
					if id := intValue(prior, f[1]); id != 0 {
						body[f[0]] = id
					}
				}
			}
			// When virtual disks are attached NetBox computes the disk size and rejects any value
			// that does not match, including a stale one. It is only sent when it is configured.
			if prior != nil && plan["disk_size_mb"] != nil && plan["disk_size_mb"].IsUnknown() {
				body["disk"] = nil
			}
			return nil
		},
	}
}
