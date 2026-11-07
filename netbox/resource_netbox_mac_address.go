package netbox

import (
	"strings"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var resourceNetboxMACAddressObjectTypeOptions = []string{"virtualization.vminterface", "dcim.interface"}

func resourceNetboxMACAddress() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxMACAddressCreate,
		Read:   resourceNetboxMACAddressRead,
		Update: resourceNetboxMACAddressUpdate,
		Delete: resourceNetboxMACAddressDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://netboxlabs.com/docs/netbox/models/dcim/macaddress/):

> A MAC address object in NetBox comprises a single Ethernet link layer address, and represents a MAC address as reported by or assigned to a network interface. MAC addresses can be assigned to device and virtual machine interfaces.`,

		Schema: map[string]*schema.Schema{
			"mac_address": {
				Type:     schema.TypeString,
				Required: true,
				// Netbox converts MAC addresses always to uppercase
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.EqualFold(old, new)
				},
			},
			"interface_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				RequiredWith: []string{"object_type"},
			},
			"object_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(resourceNetboxMACAddressObjectTypeOptions, false),
				Description:  buildValidValueDescription(resourceNetboxMACAddressObjectTypeOptions),
				RequiredWith: []string{"interface_id"},
			},
			"virtual_machine_interface_id": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"interface_id", "device_interface_id"},
			},
			"device_interface_id": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"interface_id", "virtual_machine_interface_id"},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"comments": {
				Type:     schema.TypeString,
				Optional: true,
			},
			tagsKey:         tagsSchema,
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const macAddressAPIPath = "dcim/mac-addresses/"

type macAddressRequest struct {
	MacAddress         string             `json:"mac_address"`
	Description        string             `json:"description"`
	Comments           string             `json:"comments"`
	AssignedObjectType *string            `json:"assigned_object_type,omitempty"`
	AssignedObjectID   *int64             `json:"assigned_object_id"`
	CustomFields       interface{}        `json:"custom_fields,omitempty"`
	Tags               []netboxapi.TagRef `json:"tags"`
}

type macAddressResponse struct {
	MacAddress         string                 `json:"mac_address"`
	Description        string                 `json:"description"`
	Comments           string                 `json:"comments"`
	AssignedObjectType *string                `json:"assigned_object_type"`
	AssignedObjectID   *int64                 `json:"assigned_object_id"`
	CustomFields       map[string]interface{} `json:"custom_fields"`
	Tags               []netboxapi.TagRef     `json:"tags"`
}

func macAddressRequestFromResourceData(api *providerState, d *schema.ResourceData) (*macAddressRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &macAddressRequest{
		MacAddress:  d.Get("mac_address").(string),
		Description: getOptionalStr(d, "description", true),
		Comments:    getOptionalStr(d, "comments", true),

		Tags: tags,
	}

	vmInterfaceID := getOptionalInt(d, "virtual_machine_interface_id")
	deviceInterfaceID := getOptionalInt(d, "device_interface_id")
	interfaceID := getOptionalInt(d, "interface_id")

	switch {
	case vmInterfaceID != nil:
		data.AssignedObjectType = strToPtr("virtualization.vminterface")
		data.AssignedObjectID = vmInterfaceID
	case deviceInterfaceID != nil:
		data.AssignedObjectType = strToPtr("dcim.interface")
		data.AssignedObjectID = deviceInterfaceID
	// if interfaceID is given, object_type must be set as well
	case interfaceID != nil:
		data.AssignedObjectType = strToPtr(d.Get("object_type").(string))
		data.AssignedObjectID = interfaceID
	// default = ip is not linked to anything
	default:
		data.AssignedObjectType = strToPtr("")
		data.AssignedObjectID = nil
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxMACAddressCreate(d *schema.ResourceData, m interface{}) error {
	data, err := macAddressRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, macAddressAPIPath, data, resourceNetboxMACAddressRead)
}

func resourceNetboxMACAddressRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, macAddressAPIPath, func(r *macAddressResponse) {
		d.Set("mac_address", r.MacAddress)
		d.Set("description", r.Description)
		d.Set("comments", r.Comments)

		if r.AssignedObjectID != nil && r.AssignedObjectType != nil {
			vmInterfaceID := getOptionalInt(d, "virtual_machine_interface_id")
			deviceInterfaceID := getOptionalInt(d, "device_interface_id")
			interfaceID := getOptionalInt(d, "interface_id")

			switch {
			case vmInterfaceID != nil && *r.AssignedObjectType == "virtualization.vminterface":
				d.Set("virtual_machine_interface_id", r.AssignedObjectID)
			case deviceInterfaceID != nil && *r.AssignedObjectType == "dcim.interface":
				d.Set("device_interface_id", r.AssignedObjectID)
			// if interfaceID is given, object_type must be set as well
			case interfaceID != nil:
				d.Set("object_type", r.AssignedObjectType)
				d.Set("interface_id", r.AssignedObjectID)
			}
		} else {
			d.Set("interface_id", nil)
			d.Set("object_type", "")
		}
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxMACAddressUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := macAddressRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, macAddressAPIPath, data, resourceNetboxMACAddressRead)
}

func resourceNetboxMACAddressDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, macAddressAPIPath)
}
