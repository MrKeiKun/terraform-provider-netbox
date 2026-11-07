package netbox

import (
	"context"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxDeviceOobIP() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxDeviceOobIPCreate,
		Read:   resourceNetboxDeviceOobIPRead,
		Update: resourceNetboxDeviceOobIPUpdate,
		Delete: resourceNetboxDeviceOobIPDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):This resource is used to define the out-of-band (OOB) IP for a given device. Modelled as a separate resource (like ` + "`netbox_device_primary_ip`" + `) to avoid a dependency cycle between the device and an IP address assigned to one of its own interfaces.`,

		Schema: map[string]*schema.Schema{
			"device_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"ip_address_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceNetboxDeviceOobIPCreate(d *schema.ResourceData, m interface{}) error {
	d.SetId(strconv.Itoa(d.Get("device_id").(int)))

	return resourceNetboxDeviceOobIPUpdate(d, m)
}

func resourceNetboxDeviceOobIPRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)

	device, err := netboxapi.Get[primaryIPResponse](context.Background(), api.rest, deviceAPIPath, id)
	if err != nil {
		if netboxapi.IsNotFound(err) {
			// If the ID is updated to blank, this tells Terraform the resource no longer exists (maybe it was destroyed out of band). Just like the destroy callback, the Read function should gracefully handle this case. https://www.terraform.io/docs/extend/writing-custom-providers.html
			d.SetId("")
			return nil
		}
		return err
	}

	if device.OobIP != nil {
		d.Set("ip_address_id", device.OobIP.ID)
	} else {
		// if the device exists, but has no oob ip, consider this element deleted
		d.SetId("")
		return nil
	}
	d.Set("device_id", device.ID)
	return nil
}

func resourceNetboxDeviceOobIPUpdate(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)

	deviceID := int64(d.Get("device_id").(int))
	IPAddressID := int64(d.Get("ip_address_id").(int))

	// Only send the one field, so everything else on the device is left alone.
	body := map[string]interface{}{"oob_ip": ipAssignment(IPAddressID)}
	if _, err := netboxapi.Patch[netboxapi.Ref](context.Background(), api.rest, deviceAPIPath, deviceID, body); err != nil {
		return err
	}

	return resourceNetboxDeviceOobIPRead(d, m)
}

func resourceNetboxDeviceOobIPDelete(d *schema.ResourceData, m interface{}) error {
	// Set ip_address_id to minus one and go to update. Update will set nil
	d.Set("ip_address_id", -1)
	return resourceNetboxDeviceOobIPUpdate(d, m)
}
