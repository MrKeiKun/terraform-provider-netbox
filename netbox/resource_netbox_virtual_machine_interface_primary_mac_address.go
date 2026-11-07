package netbox

import (
	"context"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxVirtualMachineInterfacePrimaryMACAddress() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxVirtualMachineInterfacePrimaryMACAddressCreate,
		Read:   resourceNetboxVirtualMachineInterfacePrimaryMACAddressRead,
		Update: resourceNetboxVirtualMachineInterfacePrimaryMACAddressUpdate,
		Delete: resourceNetboxVirtualMachineInterfacePrimaryMACAddressDelete,

		Description: `:meta:subcategory:Virtualization:This resource is used to define the primary MAC for a given virtual machine interface. The primary MAC is reflected in the Interface Netbox UI.`,

		Schema: map[string]*schema.Schema{
			"interface_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"mac_address_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

// interfacePrimaryMACResponse holds the primary MAC address of an interface.
type interfacePrimaryMACResponse struct {
	ID                int64           `json:"id"`
	PrimaryMacAddress *netboxapi.Ref  `json:"primary_mac_address"`
	TaggedVlans       []netboxapi.Ref `json:"tagged_vlans"`
	UntaggedVlan      *netboxapi.Ref  `json:"untagged_vlan"`
}

// interfacePrimaryMACBody builds the update for the primary MAC address of an
// interface. NetBox drops the VLANs of an interface on every save when its mode
// is not "tagged", so the current VLANs are sent along to keep them.
func interfacePrimaryMACBody(iface *interfacePrimaryMACResponse, macAddressID int64) map[string]interface{} {
	body := map[string]interface{}{
		"primary_mac_address": ipAssignment(macAddressID),
		"tagged_vlans":        netboxapi.IDs(iface.TaggedVlans),
	}
	if iface.UntaggedVlan != nil {
		body["untagged_vlan"] = iface.UntaggedVlan.ID
	}
	return body
}

func resourceNetboxVirtualMachineInterfacePrimaryMACAddressCreate(d *schema.ResourceData, m interface{}) error {
	d.SetId(strconv.Itoa(d.Get("interface_id").(int)))

	return resourceNetboxVirtualMachineInterfacePrimaryMACAddressUpdate(d, m)
}

func resourceNetboxVirtualMachineInterfacePrimaryMACAddressRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)

	iface, err := netboxapi.Get[interfacePrimaryMACResponse](context.Background(), api.rest, vmInterfaceAPIPath, id)
	if err != nil {
		if netboxapi.IsNotFound(err) {
			// If the ID is updated to blank, this tells Terraform the resource no longer exists (maybe it was destroyed out of band). Just like the destroy callback, the Read function should gracefully handle this case. https://www.terraform.io/docs/extend/writing-custom-providers.html
			d.SetId("")
			return nil
		}
		return err
	}

	if iface.PrimaryMacAddress != nil {
		d.Set("mac_address_id", iface.PrimaryMacAddress.ID)
		d.Set("interface_id", iface.ID)
	} else {
		d.SetId("")
	}
	return nil
}

func resourceNetboxVirtualMachineInterfacePrimaryMACAddressUpdate(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)

	interfaceID := int64(d.Get("interface_id").(int))
	macAddressID := int64(d.Get("mac_address_id").(int))

	// The interface is read first, see interfacePrimaryMACBody.
	iface, err := netboxapi.Get[interfacePrimaryMACResponse](context.Background(), api.rest, vmInterfaceAPIPath, interfaceID)
	if err != nil {
		return err
	}
	if _, err := netboxapi.Patch[netboxapi.Ref](context.Background(), api.rest, vmInterfaceAPIPath, interfaceID, interfacePrimaryMACBody(iface, macAddressID)); err != nil {
		return err
	}

	return resourceNetboxVirtualMachineInterfacePrimaryMACAddressRead(d, m)
}

func resourceNetboxVirtualMachineInterfacePrimaryMACAddressDelete(d *schema.ResourceData, m interface{}) error {
	// Set mac_address_id to minus one and go to update. Update will set nil
	d.Set("mac_address_id", -1)
	return resourceNetboxVirtualMachineInterfacePrimaryMACAddressUpdate(d, m)
}
