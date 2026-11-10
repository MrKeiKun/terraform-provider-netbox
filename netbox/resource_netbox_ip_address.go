package netbox

import (
	"strings"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var resourceNetboxIPAddressObjectTypeOptions = []string{"virtualization.vminterface", "dcim.interface", "ipam.fhrpgroup"}
var resourceNetboxIPAddressStatusOptions = []string{"active", "reserved", "deprecated", "dhcp", "slaac"}
var resourceNetboxIPAddressRoleOptions = []string{"loopback", "secondary", "anycast", "vip", "vrrp", "hsrp", "glbp", "carp"}

func resourceNetboxIPAddress() *schema.Resource {
	return &schema.Resource{
		CreateContext: withContext(resourceNetboxIPAddressCreate),
		ReadContext:   withContext(resourceNetboxIPAddressRead),
		UpdateContext: withContext(resourceNetboxIPAddressUpdate),
		DeleteContext: withContext(resourceNetboxIPAddressDelete),

		Description: `:meta:subcategory:IP Address Management (IPAM):From the [official documentation](https://docs.netbox.dev/en/stable/features/ipam/#ip-addresses):

> An IP address comprises a single host address (either IPv4 or IPv6) and its subnet mask. Its mask should match exactly how the IP address is configured on an interface in the real world.
>
> Like a prefix, an IP address can optionally be assigned to a VRF (otherwise, it will appear in the "global" table). IP addresses are automatically arranged under parent prefixes within their respective VRFs according to the IP hierarchy.`,

		Schema: map[string]*schema.Schema{
			"ip_address": {
				Type:         schema.TypeString,
				Description:  "The IP address of the ip address, in CIDR notation.",
				Required:     true,
				ValidateFunc: validation.IsCIDR,
			},
			"interface_id": {
				Type:         schema.TypeInt,
				Description:  "The ID of the interface associated with the ip address.",
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
				Description:   "The ID of the virtual machine interface associated with the ip address.",
				Optional:      true,
				ConflictsWith: []string{"interface_id", "device_interface_id"},
			},
			"device_interface_id": {
				Type:          schema.TypeInt,
				Description:   "The ID of the device interface associated with the ip address.",
				Optional:      true,
				ConflictsWith: []string{"interface_id", "virtual_machine_interface_id"},
			},
			"vrf_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the vrf associated with the ip address.",
				Optional:    true,
			},
			"tenant_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the tenant associated with the ip address.",
				Optional:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Required:    true,
				Description: buildValidValueDescription(resourceNetboxIPAddressStatusOptions),
			},
			"dns_name": {
				Type:        schema.TypeString,
				Description: "The DNS name of the ip address.",
				Optional:    true,
				// NetBox always converts DNS names to lowercase
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.EqualFold(old, new)
				},
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the ip address.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the ip address.",
				Optional:    true,
			},
			tagsKey: tagsSchema,
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the ip address.",
				Optional:    true,
			},
			"role": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(resourceNetboxIPAddressRoleOptions, false),
				Description:  buildValidValueDescription(resourceNetboxIPAddressRoleOptions),
			},
			"nat_inside_address_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the nat inside address associated with the ip address.",
				Optional:    true,
			},
			"nat_outside_addresses": {
				Type:        schema.TypeList,
				Description: "The nat outside addresses of the ip address.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Description: "The numeric ID of the ip address.",
							Computed:    true,
						},
						"ip_address": {
							Type:        schema.TypeString,
							Description: "The IP address of the ip address, in CIDR notation.",
							Computed:    true,
						},
						"address_family": {
							Type:        schema.TypeInt,
							Description: "The address family of the ip address.",
							Computed:    true,
						},
					},
				},
			},
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const ipAddressAPIPath = "ipam/ip-addresses/"

type ipAddressRequest struct {
	Address            string             `json:"address"`
	Status             string             `json:"status,omitempty"`
	Description        string             `json:"description,omitempty"`
	Role               string             `json:"role"`
	DnsName            string             `json:"dns_name,omitempty"`
	Vrf                *int64             `json:"vrf"`
	Tenant             *int64             `json:"tenant"`
	NatInside          *int64             `json:"nat_inside"`
	AssignedObjectType *string            `json:"assigned_object_type,omitempty"`
	AssignedObjectID   *int64             `json:"assigned_object_id"`
	Owner              *int64             `json:"owner"`
	Comments           string             `json:"comments,omitempty"`
	CustomFields       interface{}        `json:"custom_fields,omitempty"`
	Tags               []netboxapi.TagRef `json:"tags"`
}

type ipAddressResponse struct {
	Address            string            `json:"address"`
	Status             *netboxapi.Choice `json:"status"`
	Description        string            `json:"description"`
	Role               *netboxapi.Choice `json:"role"`
	DnsName            string            `json:"dns_name"`
	Vrf                *netboxapi.Ref    `json:"vrf"`
	Tenant             *netboxapi.Ref    `json:"tenant"`
	NatInside          *netboxapi.Ref    `json:"nat_inside"`
	AssignedObjectType *string           `json:"assigned_object_type"`
	AssignedObjectID   *int64            `json:"assigned_object_id"`
	NatOutside         []struct {
		ID      int64  `json:"id"`
		Address string `json:"address"`
		Family  int64  `json:"family"`
	} `json:"nat_outside"`
	Owner        *netboxapi.Ref         `json:"owner"`
	Comments     string                 `json:"comments"`
	CustomFields map[string]interface{} `json:"custom_fields"`
	Tags         []netboxapi.TagRef     `json:"tags"`
}

func ipAddressRequestFromResourceData(api *providerState, d *schema.ResourceData) (*ipAddressRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &ipAddressRequest{
		Comments:    getOptionalStr(d, "comments", true),
		Owner:       getOptionalInt(d, "owner_id"),
		Address:     d.Get("ip_address").(string),
		Status:      d.Get("status").(string),
		Description: getOptionalStr(d, "description", true),
		Role:        getOptionalStr(d, "role", false),
		DnsName:     getOptionalStr(d, "dns_name", true),
		Vrf:         getOptionalInt(d, "vrf_id"),
		Tenant:      getOptionalInt(d, "tenant_id"),
		NatInside:   getOptionalInt(d, "nat_inside_address_id"),

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

func resourceNetboxIPAddressCreate(d *schema.ResourceData, m interface{}) error {
	data, err := ipAddressRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, ipAddressAPIPath, data, resourceNetboxIPAddressRead)
}

func resourceNetboxIPAddressRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, ipAddressAPIPath, func(r *ipAddressResponse) {
		d.Set("comments", r.Comments)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("ip_address", r.Address)
		d.Set("status", netboxapi.ChoiceValue(r.Status))
		d.Set("description", r.Description)
		d.Set("role", netboxapi.ChoiceValue(r.Role))
		if r.DnsName != "" {
			d.Set("dns_name", r.DnsName)
		}
		d.Set("vrf_id", netboxapi.RefID(r.Vrf))
		d.Set("tenant_id", netboxapi.RefID(r.Tenant))
		d.Set("nat_inside_address_id", netboxapi.RefID(r.NatInside))

		if r.AssignedObjectID != nil {
			vmInterfaceID := getOptionalInt(d, "virtual_machine_interface_id")
			deviceInterfaceID := getOptionalInt(d, "device_interface_id")
			interfaceID := getOptionalInt(d, "interface_id")

			switch {
			case vmInterfaceID != nil:
				d.Set("virtual_machine_interface_id", r.AssignedObjectID)
				d.Set("interface_id", nil)
				d.Set("object_type", "")
			case deviceInterfaceID != nil:
				d.Set("device_interface_id", r.AssignedObjectID)
				d.Set("interface_id", nil)
				d.Set("object_type", "")
			// if interfaceID is given, object_type must be set as well
			case interfaceID != nil:
				d.Set("interface_id", r.AssignedObjectID)
				d.Set("object_type", r.AssignedObjectType)
			default:
				// Set changes made to the ip address outside of Terraform and update the state accordingly.
				d.Set("interface_id", r.AssignedObjectID)
				d.Set("object_type", r.AssignedObjectType)
			}
		} else {
			d.Set("virtual_machine_interface_id", nil)
			d.Set("device_interface_id", nil)
			d.Set("interface_id", nil)
			d.Set("object_type", "")
		}

		if r.NatOutside != nil {
			var s []map[string]interface{}
			for _, v := range r.NatOutside {
				s = append(s, map[string]interface{}{
					"id":             v.ID,
					"ip_address":     v.Address,
					"address_family": v.Family,
				})
			}
			d.Set("nat_outside_addresses", s)
		} else {
			d.Set("nat_outside_addresses", nil)
		}
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxIPAddressUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := ipAddressRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return putObject(d, m, ipAddressAPIPath, data, resourceNetboxIPAddressRead)
}

func resourceNetboxIPAddressDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, ipAddressAPIPath)
}
