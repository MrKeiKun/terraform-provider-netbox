package fwprovider

import (
	"sort"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/fwresource"
)

// Inventory returns the specs of the resources and data sources that are described declaratively. The
// drift report in tools/netboxdrift compares them with the NetBox API schema.
func Inventory() ([]*fwresource.Spec, []*fwresource.DataSourceSpec) {
	resources := make([]*fwresource.Spec, 0, len(standardResources)+len(manualResources))
	resources = append(resources, standardResources...)
	resources = append(resources, manualResources...)

	dataSources := HandwrittenDataSources()
	dataSources = append(dataSources, generatedDataSources...)
	dataSources = append(dataSources, aliasedDataSources()...)
	return resources, dataSources
}

// HandwrittenDataSources returns the data sources that are not generated. tools/gendatasources uses it
// to find out which data sources are still missing.
func HandwrittenDataSources() []*fwresource.DataSourceSpec {
	out := make([]*fwresource.DataSourceSpec, 0, len(standardDataSources)+len(customDataSources))
	out = append(out, standardDataSources...)
	out = append(out, customDataSources...)
	return out
}

// GeneratedFilters returns the filters that tools/gendatasources added to the plural data sources
// that are written by hand, by type name.
func GeneratedFilters() map[string]map[string]string {
	return generatedFilters
}

func sortStrings(s []string) { sort.Strings(s) }
