package netbox

import (
	"context"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxDevicePrimaryIP() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxDevicePrimaryIPCreate,
		Read:   resourceNetboxDevicePrimaryIPRead,
		Update: resourceNetboxDevicePrimaryIPUpdate,
		Delete: resourceNetboxDevicePrimaryIPDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):This resource is used to define the primary IP for a given device. The primary IP is reflected in the device Netbox UI, which identifies the Primary IPv4 and IPv6 addresses.`,

		Schema: map[string]*schema.Schema{
			"device_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the device associated with the device primary ip.",
				Required:    true,
			},
			"ip_address_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the ip address associated with the device primary ip.",
				Required:    true,
			},
			"ip_address_version": {
				Type:         schema.TypeInt,
				Description:  "The ip address version of the device primary ip.",
				ValidateFunc: validation.IntInSlice([]int{4, 6}),
				Optional:     true,
				Default:      4,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const deviceAPIPath = "dcim/devices/"

// primaryIPResponse holds the IP assignments of a device or virtual machine.
type primaryIPResponse struct {
	ID         int64          `json:"id"`
	OobIP      *netboxapi.Ref `json:"oob_ip"`
	PrimaryIp4 *netboxapi.Ref `json:"primary_ip4"`
	PrimaryIp6 *netboxapi.Ref `json:"primary_ip6"`
}

// ipAssignment returns the value to send for an IP field: nil clears it when
// the ID is -1, which is how delete asks for the assignment to be removed.
func ipAssignment(ipAddressID int64) interface{} {
	if ipAddressID == -1 {
		return nil
	}
	return ipAddressID
}

func resourceNetboxDevicePrimaryIPCreate(d *schema.ResourceData, m interface{}) error {
	d.SetId(strconv.Itoa(d.Get("device_id").(int)))

	return resourceNetboxDevicePrimaryIPUpdate(d, m)
}

func resourceNetboxDevicePrimaryIPRead(d *schema.ResourceData, m interface{}) error {
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

	IPAddressVersion := d.Get("ip_address_version")
	// On import the version is unknown; infer it from whichever primary IP is set.
	if IPAddressVersion == 0 {
		if device.PrimaryIp4 != nil {
			IPAddressVersion = 4
		} else if device.PrimaryIp6 != nil {
			IPAddressVersion = 6
		}
	}
	d.Set("ip_address_version", IPAddressVersion)
	if IPAddressVersion == 4 && device.PrimaryIp4 != nil {
		d.Set("ip_address_id", device.PrimaryIp4.ID)
	} else if IPAddressVersion == 6 && device.PrimaryIp6 != nil {
		d.Set("ip_address_id", device.PrimaryIp6.ID)
	} else {
		// if the device exists, but has no primary ip, consider this element deleted
		d.SetId("")
		return nil
	}
	d.Set("device_id", device.ID)
	return nil
}

func resourceNetboxDevicePrimaryIPUpdate(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)

	deviceID := int64(d.Get("device_id").(int))
	IPAddressID := int64(d.Get("ip_address_id").(int))
	IPAddressVersion := int64(d.Get("ip_address_version").(int))

	field := "primary_ip6"
	if IPAddressVersion == 4 {
		field = "primary_ip4"
	}

	// Only send the one field, so everything else on the device is left alone.
	body := map[string]interface{}{field: ipAssignment(IPAddressID)}
	if _, err := netboxapi.Patch[netboxapi.Ref](context.Background(), api.rest, deviceAPIPath, deviceID, body); err != nil {
		return err
	}

	return resourceNetboxDevicePrimaryIPRead(d, m)
}

func resourceNetboxDevicePrimaryIPDelete(d *schema.ResourceData, m interface{}) error {
	// Set ip_address_id to minus one and go to update. Update will set nil
	d.Set("ip_address_id", -1)
	return resourceNetboxDevicePrimaryIPUpdate(d, m)
}
