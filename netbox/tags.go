package netbox

import (
	"context"
	"fmt"
	"net/url"
	"slices"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	tagsKey    = "tags"
	tagsAllKey = "tags_all"
)

var tagsSchema = &schema.Schema{
	Type:        schema.TypeSet,
	Description: "Tags assigned to the object, by name.",
	Elem: &schema.Schema{
		Type: schema.TypeString,
	},
	Optional: true,
	Set:      schema.HashString,
}

var tagsAllSchema = &schema.Schema{
	Type:        schema.TypeSet,
	Description: "All tags assigned to the object, including those not managed by Terraform.",
	Elem: &schema.Schema{
		Type: schema.TypeString,
	},
	Computed: true,
	Set:      schema.HashString,
}

var tagsSchemaRead = &schema.Schema{
	Type:        schema.TypeSet,
	Description: "Tags assigned to the object, by name.",
	Elem: &schema.Schema{
		Type: schema.TypeString,
	},
	Computed: true,
	Set:      schema.HashString,
}

func (s *providerState) readTagNames(d *schema.ResourceData, apiTags []string) {
	allTags := schema.NewSet(schema.HashString, nil)
	for _, t := range apiTags {
		allTags.Add(t)
	}
	d.Set(tagsAllKey, allTags.List())

	configTags := make([]string, len(apiTags))
	cf := d.GetRawConfig()
	if cf.IsNull() || !cf.IsKnown() {
		cf = d.GetRawState() // config is missing during refresh
	}
	if !cf.IsNull() && cf.IsKnown() { // there is some config
		c := cf.GetAttr(tagsKey)
		if !c.IsNull() && c.IsKnown() { // tags are configured
			for _, t := range c.AsValueSet().Values() {
				configTags = append(configTags, t.AsString())
			}
		}
	}

	resourceTags := schema.NewSet(schema.HashString, nil)
	// remove default tags (except when configured on the resource)
	for _, tag := range apiTags {
		if !s.defaultTags.Contains(tag) || slices.Contains(configTags, tag) {
			resourceTags.Add(tag)
		}
	}

	d.Set(tagsKey, resourceTags.List())
}

// tagRef returns the tag with the given name, using the cache of default tags
// when possible.
func (s *providerState) tagRef(name string) (netboxapi.TagRef, error) {
	if ref, ok := s.tagCache[name]; ok {
		return ref, nil
	}
	return findTagRef(s.rest, name)
}

// getTagRefsFromResourceDataSet resolves a set of tag names to tags NetBox
// understands in requests.
func getTagRefsFromResourceDataSet(state *providerState, d interface{}) ([]netboxapi.TagRef, error) {
	tagList := d.(*schema.Set).List()
	tags := make([]netboxapi.TagRef, 0, len(tagList))
	for _, tag := range tagList {
		ref, err := state.tagRef(tag.(string))
		if err != nil {
			return tags, err
		}
		tags = append(tags, ref)
	}

	return tags, nil
}

func findTagRef(c *netboxapi.Client, name string) (netboxapi.TagRef, error) {
	// We search for a unique tag. Having two hits suffices to know its not unique.
	query := url.Values{"name": {name}, "limit": {"2"}}
	tags, err := netboxapi.List[netboxapi.TagRef](context.Background(), c, "extras/tags/", query)
	if err != nil {
		return netboxapi.TagRef{}, fmt.Errorf("API Error trying to retrieve tag %q from netbox: %w", name, err)
	}

	switch len(tags) {
	case 0:
		return netboxapi.TagRef{}, fmt.Errorf("could not locate referenced tag %q in netbox, no results", name)
	case 1:
		return tags[0], nil
	default:
		return netboxapi.TagRef{}, fmt.Errorf("could not map tag %q to unique tag in netbox, %d results", name, len(tags))
	}
}

// tagNames returns the names of the given tags.
func tagNames(refs []netboxapi.TagRef) []string {
	names := make([]string, 0, len(refs))
	for _, ref := range refs {
		names = append(names, ref.Name)
	}
	return names
}

// readTagRefs is readTags for tags returned by the REST client.
func (s *providerState) readTagRefs(d *schema.ResourceData, apiTags []netboxapi.TagRef) {
	s.readTagNames(d, tagNames(apiTags))
}
