package netbox

import (
	"context"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxDeviceInterfacePrimaryMACAddress() *schema.Resource {
	return &schema.Resource{
		CreateContext: withContext(resourceNetboxDeviceInterfacePrimaryMACAddressCreate),
		ReadContext:   withContext(resourceNetboxDeviceInterfacePrimaryMACAddressRead),
		UpdateContext: withContext(resourceNetboxDeviceInterfacePrimaryMACAddressUpdate),
		DeleteContext: withContext(resourceNetboxDeviceInterfacePrimaryMACAddressDelete),

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):This resource is used to define the primary MAC for a given device interface. The primary MAC is reflected in the device interface Netbox UI.`,

		Schema: map[string]*schema.Schema{
			"interface_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the interface associated with the device interface primary mac address.",
				Required:    true,
			},
			"mac_address_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the mac address associated with the device interface primary mac address.",
				Required:    true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceNetboxDeviceInterfacePrimaryMACAddressCreate(d *schema.ResourceData, m interface{}) error {
	d.SetId(strconv.Itoa(d.Get("interface_id").(int)))

	return resourceNetboxDeviceInterfacePrimaryMACAddressUpdate(d, m)
}

func resourceNetboxDeviceInterfacePrimaryMACAddressRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)

	iface, err := netboxapi.Get[interfacePrimaryMACResponse](context.Background(), api.rest, deviceInterfaceAPIPath, id)
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

func resourceNetboxDeviceInterfacePrimaryMACAddressUpdate(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)

	interfaceID := int64(d.Get("interface_id").(int))
	macAddressID := int64(d.Get("mac_address_id").(int))

	// The interface is read first, see interfacePrimaryMACBody.
	iface, err := netboxapi.Get[interfacePrimaryMACResponse](context.Background(), api.rest, deviceInterfaceAPIPath, interfaceID)
	if err != nil {
		return err
	}
	if _, err := netboxapi.Patch[netboxapi.Ref](context.Background(), api.rest, deviceInterfaceAPIPath, interfaceID, interfacePrimaryMACBody(iface, macAddressID)); err != nil {
		return err
	}

	return resourceNetboxDeviceInterfacePrimaryMACAddressRead(d, m)
}

func resourceNetboxDeviceInterfacePrimaryMACAddressDelete(d *schema.ResourceData, m interface{}) error {
	// Set mac_address_id to minus one and go to update. Update will set nil
	d.Set("mac_address_id", -1)
	return resourceNetboxDeviceInterfacePrimaryMACAddressUpdate(d, m)
}
