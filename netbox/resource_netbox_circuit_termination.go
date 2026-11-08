package netbox

import (
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var resourceNetboxCircuitTerminationTermSideOptions = []string{"A", "Z"}

func resourceNetboxCircuitTermination() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxCircuitTerminationCreate,
		Read:   resourceNetboxCircuitTerminationRead,
		Update: resourceNetboxCircuitTerminationUpdate,
		Delete: resourceNetboxCircuitTerminationDelete,

		Description: `:meta:subcategory:Circuits:From the [official documentation](https://docs.netbox.dev/en/stable/features/circuits/#circuit-terminations):

> The association of a circuit with a particular site and/or device is modeled separately as a circuit termination. A circuit may have up to two terminations, labeled A and Z. A single-termination circuit can be used when you don't know (or care) about the far end of a circuit (for example, an Internet access circuit which connects to a transit provider). A dual-termination circuit is useful for tracking circuits which connect two sites.
>
> Each circuit termination is attached to either a site or to a provider network. Site terminations may optionally be connected via a cable to a specific device interface or port within that site. Each termination must be assigned a port speed, and can optionally be assigned an upstream speed if it differs from the downstream speed (a common scenario with e.g. DOCSIS cable modems). Fields are also available to track cross-connect and patch panel details.`,

		Schema: map[string]*schema.Schema{
			"circuit_id": {
				Type:        schema.TypeInt,
				Description: "The ID of the circuit associated with the circuit termination.",
				Required:    true,
			},
			"location_id": {
				Type:         schema.TypeInt,
				Description:  "The ID of the location associated with the circuit termination.",
				Optional:     true,
				ExactlyOneOf: []string{"site_id", "site_group_id", "region_id", "provider_network_id"},
			},
			"site_id": {
				Type:         schema.TypeInt,
				Description:  "The ID of the site associated with the circuit termination.",
				Optional:     true,
				ExactlyOneOf: []string{"location_id", "site_group_id", "region_id", "provider_network_id"},
			},
			"site_group_id": {
				Type:         schema.TypeInt,
				Description:  "The ID of the site group associated with the circuit termination.",
				Optional:     true,
				ExactlyOneOf: []string{"location_id", "site_id", "region_id", "provider_network_id"},
			},
			"region_id": {
				Type:         schema.TypeInt,
				Description:  "The ID of the region associated with the circuit termination.",
				Optional:     true,
				ExactlyOneOf: []string{"location_id", "site_id", "site_group_id", "provider_network_id"},
			},
			"provider_network_id": {
				Type:         schema.TypeInt,
				Description:  "The ID of the provider network associated with the circuit termination.",
				Optional:     true,
				ExactlyOneOf: []string{"location_id", "site_id", "site_group_id", "region_id"},
			},
			"port_speed": {
				Type:        schema.TypeInt,
				Description: "The port speed of the circuit termination.",
				Optional:    true,
			},
			"upstream_speed": {
				Type:        schema.TypeInt,
				Description: "The upstream speed of the circuit termination.",
				Optional:    true,
			},
			"term_side": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(resourceNetboxCircuitTerminationTermSideOptions, false),
				Description:  buildValidValueDescription(resourceNetboxCircuitTerminationTermSideOptions),
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A description of the circuit termination.",
				Optional:    true,
			},
			"xconnect_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the local cross-connect",
			},
			"pp_info": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Patch panel ID and port number(s)",
			},
			"mark_connected": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Treat as if a cable is connected",
			},
			tagsKey:         tagsSchema,
			customFieldsKey: customFieldsSchema,
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

const circuitTerminationAPIPath = "circuits/circuit-terminations/"

type circuitTerminationRequest struct {
	TermSide        string             `json:"term_side"`
	Description     string             `json:"description,omitempty"`
	Circuit         *int64             `json:"circuit"`
	PortSpeed       *int64             `json:"port_speed,omitempty"`
	UpstreamSpeed   *int64             `json:"upstream_speed,omitempty"`
	TerminationType *string            `json:"termination_type"`
	TerminationID   *int64             `json:"termination_id"`
	XconnectID      string             `json:"xconnect_id,omitempty"`
	PpInfo          string             `json:"pp_info,omitempty"`
	MarkConnected   *bool              `json:"mark_connected,omitempty"`
	CustomFields    interface{}        `json:"custom_fields,omitempty"`
	Tags            []netboxapi.TagRef `json:"tags"`
}

type circuitTerminationResponse struct {
	TermSide        string                 `json:"term_side"`
	Description     string                 `json:"description"`
	Circuit         *netboxapi.Ref         `json:"circuit"`
	PortSpeed       *int64                 `json:"port_speed"`
	UpstreamSpeed   *int64                 `json:"upstream_speed"`
	TerminationType *string                `json:"termination_type"`
	TerminationID   *int64                 `json:"termination_id"`
	XconnectID      string                 `json:"xconnect_id"`
	PpInfo          string                 `json:"pp_info"`
	MarkConnected   bool                   `json:"mark_connected"`
	CustomFields    map[string]interface{} `json:"custom_fields"`
	Tags            []netboxapi.TagRef     `json:"tags"`
}

func circuitTerminationRequestFromResourceData(api *providerState, d *schema.ResourceData) (*circuitTerminationRequest, error) {
	tags, err := getTagRefsFromResourceDataSet(api, d.Get(tagsAllKey))
	if err != nil {
		return nil, err
	}
	data := &circuitTerminationRequest{
		MarkConnected: getOptionalBool(d, "mark_connected"),
		PpInfo:        getOptionalStr(d, "pp_info", true),
		XconnectID:    getOptionalStr(d, "xconnect_id", true),
		TermSide:      d.Get("term_side").(string),
		Description:   d.Get("description").(string),
		PortSpeed:     getOptionalInt(d, "port_speed"),
		UpstreamSpeed: getOptionalInt(d, "upstream_speed"),

		Tags: tags,
	}
	if v, ok := d.GetOk("circuit_id"); ok {
		data.Circuit = int64ToPtr(int64(v.(int)))
	}

	siteID := getOptionalInt(d, "site_id")
	siteGroupID := getOptionalInt(d, "site_group_id")
	locationID := getOptionalInt(d, "location_id")
	regionID := getOptionalInt(d, "region_id")
	providerNetworkID := getOptionalInt(d, "provider_network_id")

	switch {
	case siteID != nil:
		data.TerminationType = strToPtr("dcim.site")
		data.TerminationID = siteID
	case siteGroupID != nil:
		data.TerminationType = strToPtr("dcim.sitegroup")
		data.TerminationID = siteGroupID
	case locationID != nil:
		data.TerminationType = strToPtr("dcim.location")
		data.TerminationID = locationID
	case regionID != nil:
		data.TerminationType = strToPtr("dcim.region")
		data.TerminationID = regionID
	case providerNetworkID != nil:
		data.TerminationType = strToPtr("circuits.providernetwork")
		data.TerminationID = providerNetworkID
	default:
		data.TerminationType = nil
		data.TerminationID = nil
	}
	if ct, ok := managedCustomFields(d); ok {
		data.CustomFields = ct
	}
	return data, nil
}

func resourceNetboxCircuitTerminationCreate(d *schema.ResourceData, m interface{}) error {
	data, err := circuitTerminationRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return createObject(d, m, circuitTerminationAPIPath, data, resourceNetboxCircuitTerminationRead)
}

func resourceNetboxCircuitTerminationRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*providerState)
	return readObject(d, m, circuitTerminationAPIPath, func(r *circuitTerminationResponse) {
		d.Set("mark_connected", r.MarkConnected)
		d.Set("pp_info", r.PpInfo)
		d.Set("xconnect_id", r.XconnectID)
		d.Set("term_side", r.TermSide)
		d.Set("description", r.Description)
		d.Set("circuit_id", netboxapi.RefID(r.Circuit))
		d.Set("port_speed", r.PortSpeed)
		d.Set("upstream_speed", r.UpstreamSpeed)

		d.Set("site_id", nil)
		d.Set("site_group_id", nil)
		d.Set("location_id", nil)
		d.Set("region_id", nil)
		d.Set("provider_network_id", nil)
		if r.TerminationType != nil && r.TerminationID != nil {
			switch *r.TerminationType {
			case "dcim.site":
				d.Set("site_id", *r.TerminationID)
			case "dcim.sitegroup":
				d.Set("site_group_id", *r.TerminationID)
			case "dcim.location":
				d.Set("location_id", *r.TerminationID)
			case "dcim.region":
				d.Set("region_id", *r.TerminationID)
			case "circuits.providernetwork":
				d.Set("provider_network_id", *r.TerminationID)
			}
		}
		d.Set(customFieldsKey, getCustomFields(r.CustomFields))
		api.readTagRefs(d, r.Tags)
	})
}

func resourceNetboxCircuitTerminationUpdate(d *schema.ResourceData, m interface{}) error {
	data, err := circuitTerminationRequestFromResourceData(m.(*providerState), d)
	if err != nil {
		return err
	}
	return patchObject(d, m, circuitTerminationAPIPath, data, resourceNetboxCircuitTerminationRead)
}

func resourceNetboxCircuitTerminationDelete(d *schema.ResourceData, m interface{}) error {
	return deleteObject(d, m, circuitTerminationAPIPath)
}
