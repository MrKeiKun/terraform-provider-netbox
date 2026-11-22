package fwresource

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/netboxapi"
	"github.com/MrKeiKun/terraform-provider-netbox/internal/providerstate"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// pageWorkers is how many pages of a plural data source are asked for at the same time. NetBox
// answers a page of 1000 objects in a second or more, so reading them one after the other takes long
// on a large NetBox. More than a few at once only makes NetBox slower for everyone else.
const pageWorkers = 4

// DataSourceSpec describes a data source. A singular data source finds exactly one object with the
// filters given as optional attributes. A plural one lists the objects that match its filter blocks.
type DataSourceSpec struct {
	TypeName    string
	Description string
	// Path is the API collection, for example "dcim/sites/".
	Path string
	// SinceVersion is the first NetBox version that has the object type, for example "4.5.0".
	SinceVersion string
	// FilterSince maps filters of a plural data source to the first NetBox version that has them. An older
	// NetBox would refuse the request.
	FilterSince map[string]string
	// Noun names the object in error messages, for example "site".
	Noun string
	// Attrs are the attributes. Optional ones are filters (singular) and Computed ones are results.
	// For a plural data source they are the attributes of each result.
	Attrs []Attr

	// CustomFields adds the custom_fields attribute, which filters and holds the custom fields.
	CustomFields bool

	AtLeastOneOf [][]string
	ExactlyOneOf [][]string
	Conflicts    [][]string

	// MoreThanOne and NoMatch are the error messages of a singular data source.
	MoreThanOne string
	NoMatch     string

	// Plural turns the data source into one that lists objects.
	Plural *Plural

	// Custom looks up the objects itself instead of querying Path. A singular data source expects
	// exactly one object.
	Custom func(ctx context.Context, st *providerstate.State, cfg Values, req datasource.ReadRequest) ([]map[string]any, error)
	// ModifyQuery changes the query before the request is sent.
	ModifyQuery func(ctx context.Context, cfg Values, q url.Values) error
	// Keep filters the results on the client side. Only used by plural data sources.
	Keep func(cfg Values, obj map[string]any) bool
	// ExtraFilters are attributes that only exist for filtering.
	ExtraFilters []Attr

	// FilterBlock adds a filter block whose attributes are sent as query parameters.
	FilterBlock []Attr

	// ResultScopes add, for every result, the attributes of the scopes with the ID of the object the
	// result is scoped to.
	ResultScopes []Scope
}

// Plural describes a data source that lists objects.
type Plural struct {
	// Result is the name of the attribute that holds the objects, for example tenants.
	Result string
	// What names the objects in error messages.
	What string
	// Filters maps the name of a filter to the query parameter. A parameter ending in [] takes a comma
	// separated list.
	Filters map[string]string
	// Lookups are the lookup expressions NetBox has for a query parameter: "name" -> ["ic", "isw",
	// "n"] for the parameters name__ic, name__isw and name__n. A filter block can name a filter with
	// one of its lookups, as in status__n.
	Lookups map[string][]string
	// NoFilter leaves out the filter blocks and the limit attribute.
	NoFilter bool
	// Extra are fixed query parameters.
	Extra url.Values
	// NameRegex names the attribute that holds a regular expression the names of the results have to
	// match. It is applied after all objects were fetched.
	NameRegex string
}

// lookupSince are the lookup expressions that the oldest supported NetBox does not have, with the
// first version that has them. The others are the same from NetBox 4.4.0 on.
var lookupSince = map[string]string{
	// tag__any and tag_id__any: objects with any of the tags, where tag wants all of them
	"any": "4.6.6",
}

// filterParam returns the query parameter of a filter named in a filter block. The name is a filter of
// the data source, or one with a lookup expression of NetBox behind it, as in name__ic. A parameter
// ending in [] takes a comma separated list.
//
// NetBox ignores a query parameter it does not know and returns all objects, so a name that is not
// known is an error here and never sent.
func (p *Plural) filterParam(name string) (string, error) {
	if param, ok := p.Filters[name]; ok {
		return param, nil
	}
	base, lookup, hasLookup := strings.Cut(name, "__")
	baseParam, ok := p.Filters[base]
	if !hasLookup || !ok {
		return "", fmt.Errorf("'%s' is not a supported filter parameter%s", name, p.suggestion(name))
	}
	query, isList := strings.CutSuffix(baseParam, "[]")
	lookups := p.Lookups[query]
	if !slices.Contains(lookups, lookup) {
		if len(lookups) == 0 {
			return "", fmt.Errorf("'%s' is not a supported filter parameter: the filter %s has no lookup expressions", name, base)
		}
		return "", fmt.Errorf("'%s' is not a supported filter parameter: the filter %s has the lookup expressions %s", name, base, strings.Join(lookups, ", "))
	}
	param := query + "__" + lookup
	if isList {
		param += "[]"
	}
	return param, nil
}

// suggestion names a filter that looks like the one that is not known, as ". Did you mean 'x'?",
// or is empty. These are the usual mix-ups: the name of an object where its id is wanted (site and
// site_id), and a singular where the filter is plural (tag and tags).
func (p *Plural) suggestion(name string) string {
	if p.Filters == nil {
		return ""
	}
	base, lookup, hasLookup := strings.Cut(name, "__")
	var found []string
	for _, c := range []string{base + "_id", base + "s", strings.TrimSuffix(base, "s"), strings.TrimSuffix(base, "_id")} {
		if c == base || c == "" || slices.Contains(found, c) {
			continue
		}
		if _, ok := p.Filters[c]; ok {
			found = append(found, c)
		}
	}
	if len(found) == 0 {
		return ""
	}
	if hasLookup {
		for i := range found {
			found[i] += "__" + lookup
		}
	}
	return fmt.Sprintf(". Did you mean '%s'?", strings.Join(found, "' or '"))
}

// pageSize returns the number of objects to ask for with each request of a plural data source. It is
// the page size of the provider, but never more than the user wants when nothing is filtered on the
// client side, so that a data source with a small limit does not read a whole page.
func (d *DataSource) pageSize(userLimit int64, clientFilter bool) int {
	size := providerstate.DefaultPageSize
	if d.state != nil && d.state.PageSize > 0 {
		size = d.state.PageSize
	}
	if !clientFilter && userLimit > 0 && userLimit < int64(size) {
		return int(userLimit)
	}
	return size
}

// NewDataSource returns the constructor of a data source described by spec.
func NewDataSource(spec *DataSourceSpec) func() datasource.DataSource {
	return func() datasource.DataSource {
		return &DataSource{spec: spec}
	}
}

// DataSource is a NetBox data source.
type DataSource struct {
	spec  *DataSourceSpec
	state *providerstate.State
}

var (
	_ datasource.DataSource                     = &DataSource{}
	_ datasource.DataSourceWithConfigure        = &DataSource{}
	_ datasource.DataSourceWithConfigValidators = &DataSource{}
)

func (d *DataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = d.spec.TypeName
}

func (d *DataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	state, ok := req.ProviderData.(*providerstate.State)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("expected *providerstate.State, got %T", req.ProviderData))
		return
	}
	d.state = state
}

func (d *DataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	paths := func(names []string) []path.Expression {
		out := make([]path.Expression, len(names))
		for i, n := range names {
			out[i] = path.MatchRoot(n)
		}
		return out
	}
	var v []datasource.ConfigValidator
	for _, g := range d.spec.ExactlyOneOf {
		v = append(v, datasourcevalidator.ExactlyOneOf(paths(g)...))
	}
	for _, g := range d.spec.AtLeastOneOf {
		v = append(v, datasourcevalidator.AtLeastOneOf(paths(g)...))
	}
	for _, g := range d.spec.Conflicts {
		v = append(v, datasourcevalidator.Conflicting(paths(g)...))
	}
	return v
}

// dsAttribute creates the data source schema attribute for a.
func (a Attr) dsAttribute() schema.Attribute {
	// An optional attribute is a filter, and takes the value of the object when it is not set.
	computed := a.Computed || a.Optional
	if !a.Optional && !a.Required {
		computed = true
	}
	switch a.Kind {
	case Int:
		out := schema.Int64Attribute{Required: a.Required, Optional: a.Optional, Computed: computed, MarkdownDescription: a.Description, DeprecationMessage: a.Deprecated}
		switch {
		case a.Min != nil && a.Max != nil && a.Optional:
			out.Validators = []validator.Int64{int64validator.Between(*a.Min, *a.Max)}
		case a.Min != nil && a.Optional:
			out.Validators = []validator.Int64{int64validator.AtLeast(*a.Min)}
		}
		return out
	case Float:
		return schema.Float64Attribute{Required: a.Required, Optional: a.Optional, Computed: computed, MarkdownDescription: a.Description, DeprecationMessage: a.Deprecated}
	case Bool:
		return schema.BoolAttribute{Required: a.Required, Optional: a.Optional, Computed: computed, MarkdownDescription: a.Description, DeprecationMessage: a.Deprecated}
	case StringMap:
		return schema.MapAttribute{ElementType: types.StringType, Required: a.Required, Optional: a.Optional, Computed: computed, MarkdownDescription: a.Description}
	case StringSet, IntSet:
		return schema.SetAttribute{ElementType: a.elemType(), Required: a.Required, Optional: a.Optional, Computed: computed, MarkdownDescription: a.Description}
	case StringList, IntList:
		return schema.ListAttribute{ElementType: a.elemType(), Required: a.Required, Optional: a.Optional, Computed: computed, MarkdownDescription: a.Description}
	case NestedSet:
		return schema.SetNestedAttribute{Computed: true, MarkdownDescription: a.Description, NestedObject: schema.NestedAttributeObject{Attributes: a.dsNested()}}
	case NestedList:
		return schema.ListNestedAttribute{Computed: true, MarkdownDescription: a.Description, NestedObject: schema.NestedAttributeObject{Attributes: a.dsNested()}}
	}
	out := schema.StringAttribute{Required: a.Required, Optional: a.Optional, Computed: computed, Sensitive: a.Sensitive, MarkdownDescription: a.Description, DeprecationMessage: a.Deprecated}
	var v []validator.String
	if len(a.OneOf) > 0 && a.Optional {
		v = append(v, stringvalidator.OneOf(a.OneOf...))
	}
	out.Validators = v
	return out
}

func (a Attr) dsNested() map[string]schema.Attribute {
	out := map[string]schema.Attribute{}
	for _, sub := range a.Sub {
		sub.Optional, sub.Required, sub.Computed = false, false, true
		out[sub.Name] = sub.dsAttribute()
	}
	return out
}

// hasIDAttr tells whether the data source declares its own id attribute, which then holds the number.
func (s *DataSourceSpec) hasIDAttr() bool {
	for _, a := range s.Attrs {
		if a.Name == "id" {
			return true
		}
	}
	return false
}

func (s *DataSourceSpec) buildSchema() schema.Schema {
	attrs := map[string]schema.Attribute{}
	if !s.hasIDAttr() {
		attrs["id"] = schema.StringAttribute{Computed: true}
	}
	blocks := map[string]schema.Block{}
	for _, a := range s.Attrs {
		attrs[a.Name] = a.dsAttribute()
	}
	for _, a := range s.ExtraFilters {
		attrs[a.Name] = a.dsAttribute()
	}
	if s.CustomFields {
		attrs[customFieldsKey] = schema.MapAttribute{ElementType: types.StringType, Optional: true, Computed: true, MarkdownDescription: customFieldsDescription}
	}
	if p := s.Plural; p != nil {
		// the attributes are the ones of each result
		elem := Attr{Name: p.Result, Kind: NestedList, Sub: s.Attrs, Description: "The " + p.What + " found."}
		attrs = map[string]schema.Attribute{"id": schema.StringAttribute{Computed: true}}
		for _, a := range s.ExtraFilters {
			a.Computed = false
			attrs[a.Name] = a.dsAttribute()
		}
		if s.CustomFields {
			attrs[customFieldsKey] = schema.MapAttribute{ElementType: types.StringType, Optional: true, Computed: true, MarkdownDescription: customFieldsDescription}
		}
		attrs[p.Result] = elem.dsAttribute()
		if !p.NoFilter {
			attrs["limit"] = schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "The maximum number of items to return. Defaults to 0, which returns all items.",
				Validators:          []validator.Int64{int64validator.AtLeast(0)},
			}
			blocks["filter"] = schema.SetNestedBlock{
				MarkdownDescription: "Filters that select which items are returned. Each filter is a name/value pair.",
				NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
					"name":  schema.StringAttribute{Required: true, MarkdownDescription: "The name of the field to filter by."},
					"value": schema.StringAttribute{Required: true, MarkdownDescription: "The value of the field to filter by."},
				}},
			}
		}
	}
	if len(s.FilterBlock) > 0 {
		inner := map[string]schema.Attribute{}
		for _, a := range s.FilterBlock {
			a.Computed = false
			a.Required = false
			a.Optional = true
			switch a.Kind {
			case Int:
				inner[a.Name] = schema.Int64Attribute{Optional: true, MarkdownDescription: a.Description}
			default:
				inner[a.Name] = schema.StringAttribute{Optional: true, MarkdownDescription: a.Description}
			}
		}
		blocks["filter"] = schema.SetNestedBlock{
			MarkdownDescription: "Filters that select the object.",
			NestedObject:        schema.NestedBlockObject{Attributes: inner},
		}
	}
	return schema.Schema{MarkdownDescription: s.Description, Attributes: attrs, Blocks: blocks}
}

func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = d.spec.buildSchema()
}

func (d *DataSource) load(ctx context.Context, cfg getter) (Values, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := Values{}
	get := func(name string, target any) {
		diags.Append(cfg.GetAttribute(ctx, path.Root(name), target)...)
		out[name] = deref(target)
	}
	if !d.spec.hasIDAttr() {
		get("id", &types.String{})
	}
	if d.spec.Plural == nil {
		for _, a := range d.spec.Attrs {
			get(a.Name, a.target())
		}
	}
	for _, a := range d.spec.ExtraFilters {
		get(a.Name, a.target())
	}
	if d.spec.CustomFields {
		get(customFieldsKey, &types.Map{})
	}
	return out, diags
}

func queryValue(v attr.Value) []string {
	switch x := v.(type) {
	case types.String:
		if !x.IsNull() && !x.IsUnknown() && x.ValueString() != "" {
			return []string{x.ValueString()}
		}
	case types.Int64:
		if !x.IsNull() && !x.IsUnknown() && x.ValueInt64() != 0 {
			return []string{strconv.FormatInt(x.ValueInt64(), 10)}
		}
	case types.Float64:
		if !x.IsNull() && !x.IsUnknown() {
			return []string{strconv.FormatFloat(x.ValueFloat64(), 'f', -1, 64)}
		}
	case types.Bool:
		if !x.IsNull() && !x.IsUnknown() {
			return []string{strconv.FormatBool(x.ValueBool())}
		}
	case types.Set:
		var out []string
		for _, e := range x.Elements() {
			out = append(out, queryValue(e)...)
		}
		return out
	case types.List:
		var out []string
		for _, e := range x.Elements() {
			out = append(out, queryValue(e)...)
		}
		return out
	}
	return nil
}

func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if notConfigured(&resp.Diagnostics, d.state) {
		return
	}
	cfg, diags := d.load(ctx, req.Config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if d.state != nil && serverOlderThan(d.state.Version, d.spec.SinceVersion) {
		resp.Diagnostics.AddError("NetBox is too old", fmt.Sprintf("%s needs NetBox %s or later, and the server runs %s.", d.spec.TypeName, d.spec.SinceVersion, d.state.Version))
		return
	}
	if d.spec.Plural != nil {
		d.readPlural(ctx, req, resp, cfg)
		return
	}

	query := url.Values{}
	for _, a := range d.spec.Attrs {
		if !a.Optional && !a.Required {
			continue
		}
		param := a.Query
		if param == "" {
			param = a.Name
		}
		if param == "-" {
			continue
		}
		for _, v := range queryValue(cfg[a.Name]) {
			query.Add(param, v)
		}
	}
	for _, a := range d.spec.ExtraFilters {
		param := a.Query
		if param == "" {
			param = a.Name
		}
		for _, v := range queryValue(cfg[a.Name]) {
			query.Add(param, v)
		}
	}
	if len(d.spec.FilterBlock) > 0 {
		var filters types.Set
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("filter"), &filters)...)
		for _, e := range filters.Elements() {
			obj, ok := e.(types.Object)
			if !ok {
				continue
			}
			for _, a := range d.spec.FilterBlock {
				param := a.Query
				if param == "" {
					param = a.Name
				}
				for _, v := range queryValue(obj.Attributes()[a.Name]) {
					query.Add(param, v)
				}
			}
		}
	}
	cfOnly, cfSet := stringMapValues(cfg[customFieldsKey])
	for k, v := range cfOnly {
		query.Set("cf_"+k, v)
	}
	if d.spec.ModifyQuery != nil {
		if err := d.spec.ModifyQuery(ctx, cfg, query); err != nil {
			resp.Diagnostics.AddError("Invalid filter", err.Error())
			return
		}
	}

	var page *netboxapi.Page[map[string]any]
	if d.spec.Custom != nil {
		objs, err := d.spec.Custom(ctx, d.state, cfg, req)
		if err != nil {
			resp.Diagnostics.AddError("Could not read "+d.spec.Noun, err.Error())
			return
		}
		page = &netboxapi.Page[map[string]any]{Count: int64(len(objs)), Results: objs}
	} else {
		// Two results are enough to know that the filter is not narrow enough.
		query.Set("limit", "2")
		var err error
		page, err = netboxapi.ListPage[map[string]any](ctx, d.state.Rest, d.spec.Path, query)
		if err != nil {
			resp.Diagnostics.AddError("Could not read "+d.spec.Noun, err.Error())
			return
		}
	}
	more := d.spec.MoreThanOne
	if more == "" {
		more = fmt.Sprintf("more than one %s returned, specify a more narrow filter", d.spec.Noun)
	}
	none := d.spec.NoMatch
	if none == "" {
		none = fmt.Sprintf("no %s found matching filter", d.spec.Noun)
	}
	if page.Count > 1 {
		resp.Diagnostics.AddError("Could not read "+d.spec.Noun, errors.New(more).Error())
		return
	}
	if page.Count == 0 || len(page.Results) == 0 {
		resp.Diagnostics.AddError("Could not read "+d.spec.Noun, errors.New(none).Error())
		return
	}
	obj := page.Results[0]

	out := Values{}
	id, _ := intFromJSON(obj["id"])
	if !d.spec.hasIDAttr() {
		out["id"] = types.StringValue(strconv.FormatInt(id, 10))
	}
	for _, a := range d.spec.Attrs {
		out[a.Name] = d.outputValue(a, obj, cfg[a.Name])
	}
	for _, a := range d.spec.ExtraFilters {
		out[a.Name] = cfg[a.Name]
	}
	if d.spec.CustomFields {
		fields := customFieldsFromObject(obj["custom_fields"], cfOnly, !cfSet, false)
		elems := make(map[string]attr.Value, len(fields))
		for k, v := range fields {
			elems[k] = types.StringValue(v)
		}
		out[customFieldsKey], _ = types.MapValue(types.StringType, elems)
	}
	for name, v := range out {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(name), v)...)
	}
}

// outputValue returns the value of an attribute of the object. An attribute that is a filter keeps
// the value of the configuration.
func (d *DataSource) outputValue(a Attr, obj map[string]any, cfg attr.Value) attr.Value {
	if (a.Optional || a.Required) && cfg != nil && !cfg.IsNull() && !cfg.IsUnknown() {
		return cfg
	}
	a.Computed = true
	if a.Kind == JSON {
		a.Computed = false
	}
	return a.fromJSON(a.pick(obj), nil)
}

// unorderedEndpoints are the list endpoints of NetBox that page badly. The tagged objects are sorted by
// values that many of them share, so the pages of one read can repeat some objects and leave out others
// (seen from 4.4.0 to 4.7.5, with a page size below the number of objects or more than 1000 objects).
// They are asked for in the order of their ID instead. The other endpoints keep the order that they have.
// This is a NetBox bug, reported as netbox-community/netbox#23767.
var unorderedEndpoints = map[string]bool{
	"extras/tagged-objects/": true,
}

func (d *DataSource) readPlural(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse, cfg Values) {
	p := d.spec.Plural

	query := url.Values{}
	for k, vs := range p.Extra {
		for _, v := range vs {
			query.Add(k, v)
		}
	}
	if unorderedEndpoints[d.spec.Path] && query.Get("ordering") == "" {
		query.Set("ordering", "id")
	}

	for _, a := range d.spec.ExtraFilters {
		param := a.Query
		if param == "" {
			param = a.Name
		}
		if param == "-" {
			continue
		}
		for _, v := range queryValue(cfg[a.Name]) {
			query.Add(param, v)
		}
	}
	var keepRegex *regexp.Regexp
	if p.NameRegex != "" {
		if expr := stringOf(cfg[p.NameRegex]); expr != "" {
			re, err := regexp.Compile(expr)
			if err != nil {
				resp.Diagnostics.AddAttributeError(path.Root(p.NameRegex), "Invalid regular expression", err.Error())
				return
			}
			keepRegex = re
		}
	}
	if cf, ok := stringMapValues(cfg[customFieldsKey]); ok {
		for k, v := range cf {
			query.Set("cf_"+k, v)
		}
	}

	var filters types.Set
	var limit types.Int64
	if !p.NoFilter {
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("filter"), &filters)...)
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("limit"), &limit)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	for _, e := range filters.Elements() {
		obj, ok := e.(types.Object)
		if !ok {
			continue
		}
		name := stringOf(obj.Attributes()["name"])
		value := stringOf(obj.Attributes()["value"])
		param, err := p.filterParam(name)
		if err != nil {
			resp.Diagnostics.AddError("Invalid filter", err.Error())
			return
		}
		// a filter with a lookup needs the NetBox version of the filter itself
		base, lookup, _ := strings.Cut(name, "__")
		if since := lookupSince[lookup]; since != "" && d.state != nil && serverOlderThan(d.state.Version, since) {
			resp.Diagnostics.AddError("NetBox is too old", fmt.Sprintf("The lookup %s of the filter %s needs NetBox %s or later, and the server runs %s. An older NetBox ignores it and returns all objects.", lookup, name, since, d.state.Version))
			return
		}
		if since := d.spec.FilterSince[base]; d.state != nil && serverOlderThan(d.state.Version, since) {
			resp.Diagnostics.AddError("NetBox is too old", fmt.Sprintf("The filter %s of %s needs NetBox %s or later, and the server runs %s.", name, d.spec.TypeName, since, d.state.Version))
			return
		}
		if value == "" {
			continue
		}
		if list, isList := strings.CutSuffix(param, "[]"); isList {
			for _, part := range strings.Split(value, ",") {
				query.Add(list, part)
			}
			continue
		}
		query.Add(param, value)
	}
	if d.spec.ModifyQuery != nil {
		if err := d.spec.ModifyQuery(ctx, cfg, query); err != nil {
			resp.Diagnostics.AddError("Invalid filter", err.Error())
			return
		}
	}

	userLimit := limit.ValueInt64()
	filtered := d.spec.Keep != nil || keepRegex != nil

	// Every page is turned into result values as it arrives and then let go. A large NetBox returns
	// tens of thousands of objects, and the decoded JSON of an object takes several times the memory
	// of its values.
	//
	// The values are kept in the form Terraform is sent, not as framework values. Setting a list of
	// 40,000 objects with State.SetAttribute takes about 2 GB, because the framework converts and
	// walks the whole list again.
	elem := Attr{Name: p.Result, Kind: NestedList, Sub: d.spec.Attrs, Computed: true}
	ot := elem.objectType()
	objType := ot.TerraformType(ctx)
	subs := make([]Attr, len(elem.Sub))
	for i, sub := range elem.Sub {
		sub.Computed, sub.Zero = true, true
		subs[i] = sub
	}
	var results []tftypes.Value
	var convErr error
	add := func(objs []map[string]any) {
		for _, o := range objs {
			if d.spec.Keep != nil && !d.spec.Keep(cfg, o) {
				continue
			}
			if keepRegex != nil {
				name, _ := o["name"].(string)
				if !keepRegex.MatchString(name) {
					continue
				}
			}
			for _, sc := range d.spec.ResultScopes {
				typ, _ := o[sc.TypeKey].(string)
				for _, opt := range sc.Options {
					if opt.Type == typ && o[sc.IDKey] != nil {
						o[opt.Attr] = o[sc.IDKey]
					}
				}
			}
			// The object is built from the values of its attributes directly. Going through a
			// framework object would work out the type of the object again for every result.
			vals := make(map[string]tftypes.Value, len(subs))
			for _, sub := range subs {
				sv := sub.fromJSON(sub.pick(o), nil)
				if sv == nil || sv.IsNull() && sub.Required {
					sv = nullOf(sub)
				}
				tv, err := sv.ToTerraformValue(ctx)
				if err != nil {
					convErr = fmt.Errorf("%s: %w", sub.Name, err)
					return
				}
				vals[sub.Name] = tv
			}
			results = append(results, tftypes.NewValue(objType, vals))
		}
	}

	if d.spec.Custom != nil {
		objs, err := d.spec.Custom(ctx, d.state, cfg, req)
		if err != nil {
			resp.Diagnostics.AddError("Could not read "+p.What, err.Error())
			return
		}
		add(objs)
	}
	// The first page tells how many objects there are and how many NetBox returns for a page. The
	// pages after it are asked for a few at a time, and added in order.
	fetched, per, total := 0, 0, 0
	for d.spec.Custom == nil {
		n := 1
		if per > 0 {
			n = min(pageWorkers, max(1, (total-fetched+per-1)/per))
		}
		pages := make([]*netboxapi.Page[map[string]any], n)
		errs := make([]error, n)
		var wg sync.WaitGroup
		for i := range pages {
			q := url.Values{}
			for k, v := range query {
				q[k] = v
			}
			limit := d.pageSize(userLimit, filtered)
			if per > 0 {
				// every page has to be as long as the first, or the offsets do not line up
				limit = per
			}
			q.Set("limit", strconv.Itoa(limit))
			q.Set("offset", strconv.Itoa(fetched+i*per))
			wg.Add(1)
			go func() {
				defer wg.Done()
				pages[i], errs[i] = netboxapi.ListPage[map[string]any](ctx, d.state.Rest, d.spec.Path, q)
			}()
		}
		wg.Wait()

		start := fetched
		more := false
		for i, page := range pages {
			if errs[i] != nil {
				resp.Diagnostics.AddError("Could not read "+p.What, fmt.Sprintf("failed to fetch %s at offset %d: %s", p.What, start+i*per, errs[i]))
				return
			}
			if per == 0 {
				per, total = len(page.Results), int(page.Count)
				if !filtered && userLimit > 0 && int(userLimit) < total {
					total = int(userLimit)
				}
			}
			fetched += len(page.Results)
			add(page.Results)
			more = len(page.Results) > 0 && page.Next != nil && *page.Next != ""
			pages[i] = nil // the objects of the page are no longer needed
		}
		if !more {
			break
		}
		if !filtered && userLimit > 0 && int64(fetched) >= userLimit {
			break
		}
	}
	if userLimit > 0 && int64(len(results)) > userLimit {
		results = results[:userLimit]
	}
	if convErr != nil {
		resp.Diagnostics.AddError("Could not read "+p.What, convErr.Error())
		return
	}
	for _, a := range d.spec.ExtraFilters {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(a.Name), cfg[a.Name])...)
	}
	if d.spec.CustomFields {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(customFieldsKey), cfg[customFieldsKey])...)
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(uniqueID()))...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The result goes straight into the raw state, next to the attributes set above.
	state := map[string]tftypes.Value{}
	if err := resp.State.Raw.As(&state); err != nil {
		resp.Diagnostics.AddError("Could not read "+p.What, err.Error())
		return
	}
	if results == nil {
		results = []tftypes.Value{}
	}
	state[p.Result] = tftypes.NewValue(tftypes.List{ElementType: objType}, results)
	resp.State.Raw = tftypes.NewValue(resp.State.Raw.Type(), state)
}

var idCounter atomic.Int64

// uniqueID returns an ID for data sources that have no natural one.
func uniqueID() string {
	return fmt.Sprintf("terraform-%s-%d", time.Now().UTC().Format("20060102150405"), idCounter.Add(1))
}
