package fwprovider

import (
	"context"
	"log"
	"strconv"
)

// upgradeVirtualDiskV0 moves the size of a virtual disk from gigabytes to megabytes.
func upgradeVirtualDiskV0(_ context.Context, state map[string]any) (map[string]any, error) {
	return sizeGBToMB(state, "size_gb", "size_mb"), nil
}

// upgradeVirtualMachineV0 turns the vcpus of a virtual machine from a string into a number and
// moves the disk size to megabytes.
func upgradeVirtualMachineV0(ctx context.Context, state map[string]any) (map[string]any, error) {
	if s, ok := state["vcpus"].(string); ok {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			f = 0
		}
		log.Printf("[DEBUG] Schema upgrade: vcpus has been migrated to %g", f)
		state["vcpus"] = f
	}
	// since the provider was already released without this state migration, a vcpus that is not
	// a string is accepted as it is
	return upgradeVirtualMachineV1(ctx, state)
}

// upgradeVirtualMachineV1 moves the disk size of a virtual machine from gigabytes to megabytes.
func upgradeVirtualMachineV1(_ context.Context, state map[string]any) (map[string]any, error) {
	return sizeGBToMB(state, "disk_size_gb", "disk_size_mb"), nil
}

func sizeGBToMB(state map[string]any, from, to string) map[string]any {
	gb, ok := state[from].(float64)
	if !ok {
		log.Printf("[DEBUG] size before migration is not a number: %#v", state[from])
		state[to] = float64(0)
	} else {
		state[to] = gb * 1000
	}
	delete(state, from)
	return state
}
