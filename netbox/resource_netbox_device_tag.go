package netbox

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNetboxDeviceTag() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxDeviceTagCreate,
		Read:   resourceNetboxDeviceTagRead,
		Delete: resourceNetboxDeviceTagDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):This resource attaches a single tag to a device, similar in purpose to the ` + "`aws_ec2_tag`" + ` resource in the AWS provider.

Use this when a device is created and managed outside of Terraform (e.g. via onboarding automation or the NetBox UI) but you still want Terraform to manage the presence of one or more specific tags on it, without taking ownership of the rest of the device.

Do not use this on a device that is also managed by a ` + "`netbox_device`" + ` resource. Since that resource does not itself declare the tag, it will treat the tag added here as drift and try to remove it on every subsequent plan/apply; this is deterministic, not a race. Unlike the AWS provider, which has an ` + "`ignore_tags`" + ` provider setting for exactly this situation, this provider has no way to make a resource ignore a specific tag, so the two resources cannot currently coexist on the same device.

This resource performs a read-modify-write of the device's tag list. If the same device is targeted by more than one ` + "`netbox_device_tag`" + ` resource (or has its tags modified by anything else) at the same time, concurrent writes can race and one may silently drop the other's change.`,

		Schema: map[string]*schema.Schema{
			"device_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the device associated with the device tag.",
				Required:    true,
				ForceNew:    true,
			},
			"tag": {
				Type:        schema.TypeString,
				Description: "The tag of the device tag.",
				Required:    true,
				ForceNew:    true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceNetboxDeviceTagParseID(id string) (int64, string, error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("invalid ID %q, expected format <device_id>:<tag>", id)
	}
	deviceID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("invalid device ID in %q: %w", id, err)
	}
	return deviceID, parts[1], nil
}

// deviceTagsResponse is the part of a device that this resource needs.
type deviceTagsResponse struct {
	ID   int64              `json:"id"`
	Tags []netboxapi.TagRef `json:"tags"`
}

// updateDeviceTags overwrites the tag list of a device. Only the tags are sent,
// everything else on the device is left alone.
func updateDeviceTags(api *providerState, deviceID int64, tags []netboxapi.TagRef) error {
	body := map[string]interface{}{"tags": tags}
	_, err := netboxapi.Patch[netboxapi.Ref](context.Background(), api.rest, deviceAPIPath, deviceID, body)
	return err
}

func resourceNetboxDeviceTagCreate(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)

	deviceID := int64(d.Get("device_id").(int))
	tagName := d.Get("tag").(string)

	nbTag, err := findTagRef(api.rest, tagName)
	if err != nil {
		return err
	}

	device, err := netboxapi.Get[deviceTagsResponse](context.Background(), api.rest, deviceAPIPath, deviceID)
	if err != nil {
		return err
	}

	tags := device.Tags
	alreadyPresent := false
	for _, tag := range tags {
		if tag.Name == nbTag.Name {
			alreadyPresent = true
			break
		}
	}
	if !alreadyPresent {
		tags = append(tags, nbTag)
	}

	if err := updateDeviceTags(api, deviceID, tags); err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%d:%s", deviceID, tagName))

	return resourceNetboxDeviceTagRead(d, m)
}

func resourceNetboxDeviceTagRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)

	deviceID, tagName, err := resourceNetboxDeviceTagParseID(d.Id())
	if err != nil {
		return err
	}

	device, err := netboxapi.Get[deviceTagsResponse](context.Background(), api.rest, deviceAPIPath, deviceID)
	if err != nil {
		if netboxapi.IsNotFound(err) {
			// the device is gone, so is the tag assignment
			d.SetId("")
			return nil
		}
		return err
	}

	found := false
	for _, tag := range device.Tags {
		if tag.Name == tagName {
			found = true
			break
		}
	}
	if !found {
		// the tag was removed from the device out of band
		d.SetId("")
		return nil
	}

	d.Set("device_id", deviceID)
	d.Set("tag", tagName)
	return nil
}

func resourceNetboxDeviceTagDelete(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)

	deviceID, tagName, err := resourceNetboxDeviceTagParseID(d.Id())
	if err != nil {
		return err
	}

	device, err := netboxapi.Get[deviceTagsResponse](context.Background(), api.rest, deviceAPIPath, deviceID)
	if err != nil {
		if netboxapi.IsNotFound(err) {
			// the device is already gone, nothing left to clean up
			return nil
		}
		return err
	}

	tags := []netboxapi.TagRef{}
	for _, tag := range device.Tags {
		if tag.Name == tagName {
			continue
		}
		tags = append(tags, tag)
	}

	return updateDeviceTags(api, deviceID, tags)
}
