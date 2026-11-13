package fwprovider

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwresource"
	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/MrKeiKun/terraform-provider-netbox/internal/providerstate"
)

// Resources that are known by a second name, so that configurations written for either provider
// keep working, and the resource that lets NetBox pick a free AS number.

func init() {
	// netbox_virtual_machine_interface is the name of the resource of the other NetBox provider. The
	// name netbox_interface stays, but new configurations should use the new one.
	vmInterface := derive("netbox_interface")
	vmInterface.TypeName = "netbox_virtual_machine_interface"
	vmInterface.MoveFrom = []string{"netbox_interface"}
	manualResources = append(manualResources, vmInterface)
	for _, s := range standardResources {
		if s.TypeName == "netbox_interface" {
			s.Deprecated = "Use netbox_virtual_machine_interface instead. The two resources are the same."
		}
	}

	account := derive("netbox_provider_account")
	account.TypeName = "netbox_circuit_provider_account"
	account.MoveFrom = []string{"netbox_provider_account"}
	manualResources = append(manualResources, account, availableASN())
}

func availableASN() *fwresource.Spec {
	s := derive("netbox_asn")
	s.TypeName = "netbox_available_asn"
	s.Description = ":meta:subcategory:IP Address Management (IPAM):Allocates the next free AS number from an ASN range. Import with \"<asn_range_id> <asn_id>\"."
	s.Put = true
	setAttr(s, "asn", func(a *fwresource.Attr) {
		a.Required, a.Optional, a.Computed = false, false, true
		a.OmitZero = false
		a.Min, a.Max = nil, nil
		a.Description = "The AS number that NetBox picked."
	})
	setAttr(s, "rir_id", func(a *fwresource.Attr) {
		a.Required, a.Optional, a.Computed, a.Nullable = false, false, true, false
		a.Description = "The ID of the RIR of the ASN range."
	})
	rename(s, "asn", "available asn")
	s.Attrs = append(s.Attrs, fwresource.Attr{Name: "asn_range_id", Kind: fwresource.Int, Required: true, Local: true, ForceNew: true, Description: "The ID of the ASN range the AS number is taken from."})
	s.ModifyBody = func(_ context.Context, _, prior fwresource.Values, body map[string]any) error {
		if prior != nil {
			body["asn"] = intValue(prior, "asn")
			body["rir"] = intValue(prior, "rir_id")
		}
		return nil
	}
	s.Allocate = func(ctx context.Context, st *providerstate.State, plan fwresource.Values, body map[string]any) (map[string]any, error) {
		rangeID := intValue(plan, "asn_range_id")
		if rangeID == 0 {
			return nil, fmt.Errorf("asn_range_id is required to allocate an AS number")
		}
		entry := map[string]any{}
		for k, v := range body {
			if k != "asn" && k != "rir" && v != nil && v != "" {
				entry[k] = v
			}
		}
		path := fmt.Sprintf("ipam/asn-ranges/%d/available-asns/", rangeID)
		created, err := netboxapi.Request[[]map[string]any](ctx, st.Rest, http.MethodPost, path, nil, []any{entry})
		if err != nil {
			return nil, err
		}
		if len(*created) == 0 {
			return nil, fmt.Errorf("no available AS numbers in ASN range %d", rangeID)
		}
		return (*created)[0], nil
	}
	s.ImportID = func(id string) (string, map[string]any, error) {
		parts := strings.SplitN(id, " ", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return "", nil, fmt.Errorf("unexpected format of (%s), expected 'asn_range_id asn_id'", id)
		}
		rangeID, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return "", nil, fmt.Errorf("asn_range_id (%s) is not an integer", parts[0])
		}
		if _, err := strconv.ParseInt(parts[1], 10, 64); err != nil {
			return "", nil, fmt.Errorf("asn_id (%s) is not an integer", parts[1])
		}
		return parts[1], map[string]any{"asn_range_id": rangeID}, nil
	}
	return s
}
