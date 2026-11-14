package fwprovider

// Every field of the NetBox object ends up as an attribute of the data sources that are written by hand.
// tools/gendatasources lists in dataSourceExtraAttrs the fields that the hand written attributes do not
// cover, and this adds them.
func init() {
	for _, spec := range HandwrittenDataSources() {
		spec.Attrs = append(spec.Attrs, dataSourceExtraAttrs[spec.TypeName]...)
	}
}
