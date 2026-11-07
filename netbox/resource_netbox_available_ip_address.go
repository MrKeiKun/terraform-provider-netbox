package netbox

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxAvailableIPAddress() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxAvailableIPAddressCreate,
		Read:   resourceNetboxAvailableIPAddressRead,
		Update: resourceNetboxAvailableIPAddressUpdate,
		Delete: resourceNetboxAvailableIPAddressDelete,

		Description: `:meta:subcategory:IP Address Management (IPAM):Per [the docs](https://netbox.readthedocs.io/en/stable/models/ipam/ipaddress/):

> An IP address comprises a single host address (either IPv4 or IPv6) and its subnet mask. Its mask should match exactly how the IP address is configured on an interface in the real world.
> Like a prefix, an IP address can optionally be assigned to a VRF (otherwise, it will appear in the "global" table). IP addresses are automatically arranged under parent prefixes within their respective VRFs according to the IP hierarchya.
>
> Each IP address can also be assigned an operational status and a functional role. Statuses are hard-coded in NetBox and include the following:
> * Active
> * Reserved
> * Deprecated
> * DHCP
> * SLAAC (IPv6 Stateless Address Autoconfiguration)

This resource will retrieve the next available IP address from a given prefix or IP range (specified by ID)`,

		Schema: map[string]*schema.Schema{
			"prefix_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"prefix_id", "ip_range_id"},
			},
			"ip_range_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"prefix_id", "ip_range_id"},
			},
			"ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"interface_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				RequiredWith: []string{"object_type"},
			},
			"object_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(resourceNetboxIPAddressObjectTypeOptions, false),
				Description:  buildValidValueDescription(resourceNetboxIPAddressObjectTypeOptions),
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
			"vrf_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"tenant_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(resourceNetboxIPAddressStatusOptions, false),
				Description:  buildValidValueDescription(resourceNetboxIPAddressStatusOptions),
				Default:      "active",
			},
			"dns_name": {
				Type:     schema.TypeString,
				Optional: true,
				// NetBox always converts DNS names to lowercase
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.EqualFold(old, new)
				},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			tagsKey: tagsSchema,
			"role": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(resourceNetboxIPAddressRoleOptions, false),
				Description:  buildValidValueDescription(resourceNetboxIPAddressRoleOptions),
			},
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

type availableIPAddressRequest struct {
	Address            string             `json:"address"`
	Status             string             `json:"status,omitempty"`
	Description        string             `json:"description,omitempty"`
	Role               string             `json:"role"`
	DNSName            string             `json:"dns_name,omitempty"`
	Vrf                *int64             `json:"vrf"`
	Tenant             *int64             `json:"tenant"`
	AssignedObjectType *string            `json:"assigned_object_type,omitempty"`
	AssignedObjectID   *int64             `json:"assigned_object_id"`
	Tags               []netboxapi.TagRef `json:"tags"`
	CustomFields       interface{}        `json:"custom_fields,omitempty"`
}

type availableIPAddressResponse struct {
	ID                 int64                  `json:"id"`
	Address            string                 `json:"address"`
	Description        string                 `json:"description"`
	DNSName            string                 `json:"dns_name"`
	Status             *netboxapi.Choice      `json:"status"`
	Vrf                *netboxapi.Ref         `json:"vrf"`
	Tenant             *netboxapi.Ref         `json:"tenant"`
	AssignedObjectType *string                `json:"assigned_object_type"`
	AssignedObjectID   *int64                 `json:"assigned_object_id"`
	Tags               []netboxapi.TagRef     `json:"tags"`
	CustomFields       map[string]interface{} `json:"custom_fields"`
}

// availableIP is one entry of the list that asks NetBox for the next free IP address.
type availableIP struct {
	Vrf *netboxapi.Ref `json:"vrf,omitempty"`
}

// requestAvailableIP asks NetBox for the next free IP address below path and
// returns the created address.
func requestAvailableIP(api *providerState, path string, vrfID int64, what string) (*availableIPAddressResponse, error) {
	entry := availableIP{}
	if vrfID != 0 {
		entry.Vrf = &netboxapi.Ref{ID: vrfID}
	}

	created, err := netboxapi.Request[[]availableIPAddressResponse](context.Background(), api.rest, http.MethodPost, path, nil, []availableIP{entry})
	if err != nil {
		return nil, err
	}
	if len(*created) == 0 {
		return nil, fmt.Errorf("no available IP addresses in %s", what)
	}
	return &(*created)[0], nil
}

func resourceNetboxAvailableIPAddressCreate(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)

	prefixID := int64(d.Get("prefix_id").(int))
	vrfID := int64(d.Get("vrf_id").(int))
	rangeID := int64(d.Get("ip_range_id").(int))

	if prefixID != 0 {
		ip, err := requestAvailableIP(api, fmt.Sprintf("ipam/prefixes/%d/available-ips/", prefixID), vrfID, fmt.Sprintf("prefix %d", prefixID))
		if err != nil {
			return err
		}
		// Since we generated the ip_address, set that now
		d.SetId(strconv.FormatInt(ip.ID, 10))
		d.Set("ip_address", ip.Address)
	}

	if rangeID != 0 {
		ip, err := requestAvailableIP(api, fmt.Sprintf("ipam/ip-ranges/%d/available-ips/", rangeID), vrfID, fmt.Sprintf("IP range %d", rangeID))
		if err != nil {
			return err
		}
		// Since we generated the ip_address, set that now
		d.SetId(strconv.FormatInt(ip.ID, 10))
		d.Set("ip_address", ip.Address)
	}

	return resourceNetboxAvailableIPAddressUpdate(d, m)
}

func resourceNetboxAvailableIPAddressRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, "ipam/ip-addresses/", func(ipAddress *availableIPAddressResponse) {
		if ipAddress.AssignedObjectID != nil {
			vmInterfaceID := getOptionalInt(d, "virtual_machine_interface_id")
			deviceInterfaceID := getOptionalInt(d, "device_interface_id")
			interfaceID := getOptionalInt(d, "interface_id")

			switch {
			case vmInterfaceID != nil:
				d.Set("virtual_machine_interface_id", ipAddress.AssignedObjectID)
			case deviceInterfaceID != nil:
				d.Set("device_interface_id", ipAddress.AssignedObjectID)
			// if interfaceID is given, object_type must be set as well
			case interfaceID != nil:
				d.Set("object_type", ipAddress.AssignedObjectType)
				d.Set("interface_id", ipAddress.AssignedObjectID)
			}
		} else {
			d.Set("interface_id", nil)
			d.Set("object_type", "")
		}

		d.Set("vrf_id", netboxapi.RefID(ipAddress.Vrf))
		d.Set("tenant_id", netboxapi.RefID(ipAddress.Tenant))
		if ipAddress.DNSName != "" {
			d.Set("dns_name", ipAddress.DNSName)
		}

		d.Set("ip_address", ipAddress.Address)
		d.Set("description", ipAddress.Description)
		d.Set("status", netboxapi.ChoiceValue(ipAddress.Status))
		api.readTagRefs(d, ipAddress.Tags)

		d.Set(customFieldsKey, getCustomFields(ipAddress.CustomFields))
	})
}

func resourceNetboxAvailableIPAddressUpdate(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)

	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return err
	}

	data := &availableIPAddressRequest{
		Address:     d.Get("ip_address").(string),
		Status:      d.Get("status").(string),
		Description: getOptionalStr(d, "description", false),
		Role:        getOptionalStr(d, "role", false),
		DNSName:     getOptionalStr(d, "dns_name", false),
		Vrf:         getOptionalInt(d, "vrf_id"),
		Tenant:      getOptionalInt(d, "tenant_id"),
		Tags:        tags,
	}
	if cf, ok := managedCustomFields(d); ok {
		data.CustomFields = cf
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

	return putObject(d, m, "ipam/ip-addresses/", data, resourceNetboxAvailableIPAddressRead)
}

func resourceNetboxAvailableIPAddressDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, "ipam/ip-addresses/")
}
