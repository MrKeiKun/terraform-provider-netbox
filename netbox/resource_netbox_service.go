package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var resourceNetboxServiceProtocolOptions = []string{"tcp", "udp", "sctp"}

func resourceNetboxService() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxServiceCreate,
		Read:   resourceNetboxServiceRead,
		Update: resourceNetboxServiceUpdate,
		Delete: resourceNetboxServiceDelete,

		Description: `:meta:subcategory:IP Address Management (IPAM):From the [official documentation](https://docs.netbox.dev/en/stable/features/services/#services):

> A service represents a layer four TCP or UDP service available on a device or virtual machine. For example, you might want to document that an HTTP service is running on a device. Each service includes a name, protocol, and port number; for example, "SSH (TCP/22)" or "DNS (UDP/53)."
>
> A service may optionally be bound to one or more specific IP addresses belonging to its parent device or VM. (If no IP addresses are bound, the service is assumed to be reachable via any assigned IP address.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Description:  "The name of the service.",
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"virtual_machine_id": {
				Type:         schema.TypeInt,
				Description:  "The ID of the virtual machine associated with the service.",
				Optional:     true,
				ExactlyOneOf: []string{"virtual_machine_id", "device_id"},
			},
			"protocol": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ConflictsWith:    []string{"port_mappings"},
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(resourceNetboxServiceProtocolOptions, false)),
				Description:      buildValidValueDescription(resourceNetboxServiceProtocolOptions),
			},
			"port": {
				Type:         schema.TypeInt,
				Description:  "The port of the service.",
				Optional:     true,
				ExactlyOneOf: []string{"port", "ports", "port_mappings"},
				Deprecated:   "This field is deprecated. Please use the new \"ports\" attribute instead.",
			},
			"ports": {
				Type:         schema.TypeSet,
				Description:  "The ports of the service.",
				Optional:     true,
				ExactlyOneOf: []string{"port", "ports", "port_mappings"},
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"port_mappings": {
				Type:          schema.TypeSet,
				Optional:      true,
				Computed:      true,
				ExactlyOneOf:  []string{"port", "ports", "port_mappings"},
				ConflictsWith: []string{"protocol"},
				Elem:          &schema.Schema{Type: schema.TypeString},
				Description:   "The ports of the service as `protocol/port` strings, for example `tcp/80`. Use this for a service that spans more than one protocol. Needs NetBox 4.6 or newer.",
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the service.",
				Optional:    true,
			},
			"tags": {
				Type:        schema.TypeSet,
				Description: "Tags to assign to the service.",
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"device_id": {
				Type:         schema.TypeInt,
				Description:  "The ID of the device associated with the service.",
				Optional:     true,
				ExactlyOneOf: []string{"virtual_machine_id", "device_id"},
			},
			"owner_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the owner associated with the service.",
				Optional:    true,
			},
			"comments": {
				Type:        schema.TypeString,
				Description: "Free-form comments about the service.",
				Optional:    true,
			},
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const serviceAPIPath = "ipam/services/"

type serviceRequest struct {
	Name             string             `json:"name"`
	Protocol         string             `json:"protocol,omitempty"`
	Description      string             `json:"description,omitempty"`
	Ports            []int64            `json:"ports,omitempty"`
	PortMappings     []string           `json:"port_mappings,omitempty"`
	Ipaddresses      []int64            `json:"ipaddresses"`
	ParentObjectType *string            `json:"parent_object_type,omitempty"`
	ParentObjectID   *int64             `json:"parent_object_id,omitempty"`
	Owner            *int64             `json:"owner"`
	Comments         string             `json:"comments,omitempty"`
	CustomFields     interface{}        `json:"custom_fields,omitempty"`
	Tags             []netboxapi.TagRef `json:"tags"`
}

type serviceResponse struct {
	Name             string                 `json:"name"`
	Protocol         *netboxapi.Choice      `json:"protocol"`
	Description      string                 `json:"description"`
	Ports            []int64                `json:"ports"`
	PortMappings     []string               `json:"port_mappings"`
	ParentObjectType string                 `json:"parent_object_type"`
	ParentObjectID   *int64                 `json:"parent_object_id"`
	Owner            *netboxapi.Ref         `json:"owner"`
	Comments         string                 `json:"comments"`
	CustomFields     map[string]interface{} `json:"custom_fields"`
	Tags             []netboxapi.TagRef     `json:"tags"`
}

func serviceRequestFromResourceData(api *providerState, d *schema.ResourceData) (*serviceRequest, error) {
	name := d.Get("name").(string)
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &serviceRequest{
		Comments:    getOptionalStr(d, "comments", true),
		Owner:       getOptionalInt(d, "owner_id"),
		Name:        name,
		Protocol:    d.Get("protocol").(string),
		Description: d.Get("description").(string),

		Ipaddresses: []int64{},

		Tags: tags,
	}

	// for backwards compatibility, we allow either port or ports
	// the API only supports ports. We give precedence to port, if it exists.
	if raw := d.GetRawConfig(); !raw.IsNull() && !raw.GetAttr("port_mappings").IsNull() {
		data.PortMappings = getStringSet(d, "port_mappings")
		data.Protocol = ""
	} else if dataPort, ok := d.GetOk("port"); ok {
		data.Ports = []int64{int64(dataPort.(int))}
	} else if ports := d.Get("ports").(*schema.Set); ports.Len() > 0 {
		// if port is not set, ports has to be set
		for _, port := range ports.List() {
			data.Ports = append(data.Ports, int64(port.(int)))
		}
	}

	virtualMachineID := getOptionalInt(d, "virtual_machine_id")
	deviceID := getOptionalInt(d, "device_id")
	switch {
	case virtualMachineID != nil:
		data.ParentObjectType = strToPtr("virtualization.virtualmachine")
		data.ParentObjectID = virtualMachineID
	case deviceID != nil:
		data.ParentObjectType = strToPtr("dcim.device")
		data.ParentObjectID = deviceID
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxServiceCreate(d *schema.ResourceData, m interface{}) error {
	data, err := serviceRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, serviceAPIPath, data, resourceNetboxServiceUpdate)
}

func resourceNetboxServiceRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, serviceAPIPath, func(r *serviceResponse) {
		d.Set("comments", r.Comments)
		d.Set("owner_id", netboxapi.RefID(r.Owner))
		d.Set("name", r.Name)
		d.Set("protocol", netboxapi.ChoiceValue(r.Protocol))
		d.Set("description", r.Description)

		d.Set("ports", r.Ports)
		d.Set("port_mappings", r.PortMappings)
		switch r.ParentObjectType {
		case "virtualization.virtualmachine":
			d.Set("virtual_machine_id", r.ParentObjectID)
		case "dcim.device":
			d.Set("device_id", r.ParentObjectID)
		}
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxServiceUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := serviceRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return putObject(d, m, serviceAPIPath, data, resourceNetboxServiceRead)
}

func resourceNetboxServiceDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, serviceAPIPath)
}
