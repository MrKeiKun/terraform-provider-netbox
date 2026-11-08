package netbox

import (
	"context"
	"strconv"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxPrimaryIP() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxPrimaryIPCreate,
		Read:   resourceNetboxPrimaryIPRead,
		Update: resourceNetboxPrimaryIPUpdate,
		Delete: resourceNetboxPrimaryIPDelete,

		Description: `:meta:subcategory:Virtualization:This resource is used to define the primary IP for a given virtual machine. The primary IP is reflected in the Virtual machine Netbox UI, which identifies the Primary IPv4 and IPv6 addresses.`,

		Schema: map[string]*schema.Schema{
			"virtual_machine_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the virtual machine associated with the primary ip.",
				Required:    true,
			},
			"ip_address_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the ip address associated with the primary ip.",
				Required:    true,
			},
			"ip_address_version": {
				Type:         schema.TypeInt,
				Description:  "The ip address version of the primary ip.",
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

const virtualMachineAPIPath = "virtualization/virtual-machines/"

func resourceNetboxPrimaryIPCreate(d *schema.ResourceData, m interface{}) error {
	d.SetId(strconv.Itoa(d.Get("virtual_machine_id").(int)))

	return resourceNetboxPrimaryIPUpdate(d, m)
}

func resourceNetboxPrimaryIPRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)

	vm, err := netboxapi.Get[primaryIPResponse](context.Background(), api.rest, virtualMachineAPIPath, id)
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
		if vm.PrimaryIp4 != nil {
			IPAddressVersion = 4
		} else if vm.PrimaryIp6 != nil {
			IPAddressVersion = 6
		}
	}
	d.Set("ip_address_version", IPAddressVersion)
	if IPAddressVersion == 4 && vm.PrimaryIp4 != nil {
		d.Set("ip_address_id", vm.PrimaryIp4.ID)
	} else if IPAddressVersion == 6 && vm.PrimaryIp6 != nil {
		d.Set("ip_address_id", vm.PrimaryIp6.ID)
	} else {
		// if the vm exists, but has no primary ip, consider this element deleted
		d.SetId("")
		return nil
	}
	d.Set("virtual_machine_id", vm.ID)
	return nil
}

func resourceNetboxPrimaryIPUpdate(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)

	virtualMachineID := int64(d.Get("virtual_machine_id").(int))
	IPAddressID := int64(d.Get("ip_address_id").(int))
	IPAddressVersion := int64(d.Get("ip_address_version").(int))

	field := "primary_ip6"
	if IPAddressVersion == 4 {
		field = "primary_ip4"
	}

	// Only send the one field, so everything else on the virtual machine is left alone.
	body := map[string]interface{}{field: ipAssignment(IPAddressID)}
	if _, err := netboxapi.Patch[netboxapi.Ref](context.Background(), api.rest, virtualMachineAPIPath, virtualMachineID, body); err != nil {
		return err
	}

	return resourceNetboxPrimaryIPRead(d, m)
}

func resourceNetboxPrimaryIPDelete(d *schema.ResourceData, m interface{}) error {
	// Set ip_address_id to minus one and go to update. Update will set nil
	d.Set("ip_address_id", -1)
	return resourceNetboxPrimaryIPUpdate(d, m)
}
