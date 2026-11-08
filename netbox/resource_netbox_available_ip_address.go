package netbox

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
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
			"nat_inside_address_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The ID of the IP address this address is the NAT outside address of.",
			},
			"tenant_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"status": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: buildValidValueDescription(resourceNetboxIPAddressStatusOptions),
				Default:     "active",
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
	NatInside          *int64             `json:"nat_inside"`
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
	NatInside          *netboxapi.Ref         `json:"nat_inside"`
	AssignedObjectType *string                `json:"assigned_object_type"`
	AssignedObjectID   *int64                 `json:"assigned_object_id"`
	Tags               []netboxapi.TagRef     `json:"tags"`
	CustomFields       map[string]interface{} `json:"custom_fields"`
}

// availableIP is one entry of the list that asks NetBox for the next free IP address.
type availableIP struct {
	Vrf          *netboxapi.Ref         `json:"vrf,omitempty"`
	Tenant       *netboxapi.Ref         `json:"tenant,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}

// requestAvailableIP asks NetBox for the next free IP address below path and
// returns the created address.
func requestAvailableIP(api *providerState, d *schema.ResourceData, path string, vrfID int64, what string) (*availableIPAddressResponse, error) {
	entry := availableIP{}
	if vrfID != 0 {
		entry.Vrf = &netboxapi.Ref{ID: vrfID}
	}
	// Send what NetBox may require when it creates the address, such as
	// mandatory custom fields. The update that follows sets everything else.
	if tenantID := getOptionalInt(d, "tenant_id"); tenantID != nil {
		entry.Tenant = &netboxapi.Ref{ID: *tenantID}
	}
	if cf, ok := managedCustomFields(d); ok {
		entry.CustomFields = cf
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
		ip, err := requestAvailableIP(api, d, fmt.Sprintf("ipam/prefixes/%d/available-ips/", prefixID), vrfID, fmt.Sprintf("prefix %d", prefixID))
		if err != nil {
			return err
		}
		// Since we generated the ip_address, set that now
		d.SetId(strconv.FormatInt(ip.ID, 10))
		d.Set("ip_address", ip.Address)
	}

	if rangeID != 0 {
		ip, err := requestAvailableIP(api, d, fmt.Sprintf("ipam/ip-ranges/%d/available-ips/", rangeID), vrfID, fmt.Sprintf("IP range %d", rangeID))
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
		d.Set("nat_inside_address_id", netboxapi.RefID(ipAddress.NatInside))
		if ipAddress.DNSName != "" {
			d.Set("dns_name", ipAddress.DNSName)
		}

		d.Set("ip_address", ipAddress.Address)
		// After an import neither parent is known. Restore the most specific
		// prefix, so the next plan does not want to replace the address.
		if d.Get("prefix_id").(int) == 0 && d.Get("ip_range_id").(int) == 0 {
			if prefixID, ok := findParentPrefixID(api, ipAddress); ok {
				d.Set("prefix_id", prefixID)
			}
		}
		d.Set("description", ipAddress.Description)
		d.Set("status", netboxapi.ChoiceValue(ipAddress.Status))
		api.readTagRefs(d, ipAddress.Tags)

		d.Set(customFieldsKey, getCustomFields(ipAddress.CustomFields))
	})
}

// findParentPrefixID returns the ID of the most specific prefix that contains
// the address and lives in the same VRF.
func findParentPrefixID(api *providerState, ip *availableIPAddressResponse) (int64, bool) {
	host, _, err := net.ParseCIDR(ip.Address)
	if err != nil {
		return 0, false
	}
	query := url.Values{}
	query.Set("contains", host.String())
	if ip.Vrf != nil {
		query.Set("vrf_id", strconv.FormatInt(ip.Vrf.ID, 10))
	} else {
		query.Set("vrf_id", "null")
	}
	prefixes, err := netboxapi.List[struct {
		ID     int64  `json:"id"`
		Prefix string `json:"prefix"`
	}](context.Background(), api.rest, "ipam/prefixes/", query)
	if err != nil {
		return 0, false
	}
	var bestID int64
	bestLen := -1
	for _, p := range prefixes {
		_, n, err := net.ParseCIDR(p.Prefix)
		if err != nil {
			continue
		}
		if ones, _ := n.Mask.Size(); ones > bestLen {
			bestLen = ones
			bestID = p.ID
		}
	}
	return bestID, bestLen >= 0
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
		Description: getOptionalStr(d, "description", true),
		Role:        getOptionalStr(d, "role", false),
		DNSName:     getOptionalStr(d, "dns_name", true),
		Vrf:         getOptionalInt(d, "vrf_id"),
		Tenant:      getOptionalInt(d, "tenant_id"),
		NatInside:   getOptionalInt(d, "nat_inside_address_id"),
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
