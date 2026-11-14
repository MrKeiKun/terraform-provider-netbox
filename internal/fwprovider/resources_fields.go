package fwprovider

// Every read-only field of the NetBox object, such as when it was created and how many objects refer to
// it, ends up as a computed attribute of the resource. tools/gendatasources lists them in
// resourceExtraAttrs, and this adds them.
func init() {
	for _, spec := range standardResources {
		spec.Attrs = append(spec.Attrs, resourceExtraAttrs[spec.TypeName]...)
	}
	for _, spec := range manualResources {
		spec.Attrs = append(spec.Attrs, resourceExtraAttrs[spec.TypeName]...)
	}
}
