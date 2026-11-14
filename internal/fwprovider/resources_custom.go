package fwprovider

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwresource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// assignResources are the resources that point a field of an existing object
// at another object.
var assignResources = []fwresource.AssignSpec{
	{
		TypeName:           "netbox_primary_ip",
		Description:        ":meta:subcategory:Virtualization:This resource is used to define the primary IP for a given virtual machine. The primary IP is reflected in the Virtual machine Netbox UI, which identifies the Primary IPv4 and IPv6 addresses.",
		Path:               "virtualization/virtual-machines/",
		Target:             "virtual_machine_id",
		TargetDescription:  "The ID of the virtual machine associated with the primary ip.",
		Ref:                "ip_address_id",
		RefDescription:     "The ID of the ip address associated with the primary ip.",
		VersionDescription: "The ip address version of the primary ip.",
	},
	{
		TypeName:           "netbox_device_primary_ip",
		Description:        ":meta:subcategory:Data Center Inventory Management (DCIM):This resource is used to define the primary IP for a given device. The primary IP is reflected in the device Netbox UI, which identifies the Primary IPv4 and IPv6 addresses.",
		Path:               "dcim/devices/",
		Target:             "device_id",
		TargetDescription:  "The ID of the device associated with the device primary ip.",
		Ref:                "ip_address_id",
		RefDescription:     "The ID of the ip address associated with the device primary ip.",
		VersionDescription: "The ip address version of the device primary ip.",
	},
	{
		TypeName:          "netbox_device_oob_ip",
		Description:       ":meta:subcategory:Data Center Inventory Management (DCIM):This resource is used to define the out-of-band (OOB) IP for a given device. Modelled as a separate resource (like `netbox_device_primary_ip`) to avoid a dependency cycle between the device and an IP address assigned to one of its own interfaces.",
		Path:              "dcim/devices/",
		Target:            "device_id",
		TargetDescription: "The ID of the device associated with the device oob ip.",
		Ref:               "ip_address_id",
		RefDescription:    "The ID of the ip address associated with the device oob ip.",
		Field:             "oob_ip",
	},
	{
		TypeName:          "netbox_virtual_chassis_master",
		Description:       ":meta:subcategory:Data Center Inventory Management (DCIM):This resource is used to define the master device of a virtual chassis. Modelled as a separate resource (like `netbox_device_primary_ip`) to avoid a dependency cycle between the virtual chassis and its member devices. The device has to be a member of the virtual chassis. Do not combine it with `virtual_chassis_master` on the same device, or add `ignore_changes = [virtual_chassis_master]` to the device.",
		Path:              "dcim/virtual-chassis/",
		Target:            "virtual_chassis_id",
		TargetDescription: "The ID of the virtual chassis.",
		Ref:               "master_id",
		RefDescription:    "The ID of the device that is the master of the virtual chassis.",
		Field:             "master",
	},
	{
		TypeName:          "netbox_device_interface_primary_mac_address",
		Description:       ":meta:subcategory:Data Center Inventory Management (DCIM):This resource is used to define the primary MAC for a given device interface. The primary MAC is reflected in the device interface Netbox UI.",
		Path:              "dcim/interfaces/",
		Target:            "interface_id",
		TargetDescription: "The ID of the interface associated with the device interface primary mac address.",
		Ref:               "mac_address_id",
		RefDescription:    "The ID of the mac address associated with the device interface primary mac address.",
		Field:             "primary_mac_address",
		KeepVLANs:         true,
	},
	{
		TypeName:          "netbox_virtual_machine_interface_primary_mac_address",
		Description:       ":meta:subcategory:Virtualization:This resource is used to define the primary MAC for a given virtual machine interface. The primary MAC is reflected in the Interface Netbox UI.",
		Path:              "virtualization/interfaces/",
		Target:            "interface_id",
		TargetDescription: "The ID of the interface associated with the virtual machine interface primary mac address.",
		Ref:               "mac_address_id",
		RefDescription:    "The ID of the mac address associated with the virtual machine interface primary mac address.",
		Field:             "primary_mac_address",
		KeepVLANs:         true,
	},
}

// customResources returns the constructors of the hand written resources.
func customResources() []func() resource.Resource {
	var out []func() resource.Resource
	out = append(out, fwresource.NewDeviceTag)
	for _, spec := range assignResources {
		out = append(out, fwresource.NewAssign(spec))
	}
	return out
}
