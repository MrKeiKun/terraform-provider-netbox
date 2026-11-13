package fwresource

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
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
)

// DataSourceSpec describes a data source. A singular data source finds exactly one object with the
// filters given as optional attributes. A plural one lists the objects that match its filter blocks.
type DataSourceSpec struct {
	TypeName    string
	Description string
	// Path is the API collection, for example "dcim/sites/".
	Path string
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
	// NoFilter leaves out the filter blocks and the limit attribute.
	NoFilter bool
	// Extra are fixed query parameters.
	Extra url.Values
	// NameRegex names the attribute that holds a regular expression the names of the results have to
	// match. It is applied after all objects were fetched.
	NameRegex string
}

// DefaultPageSize is the number of objects fetched with each request of a plural data source.
const DefaultPageSize = 100

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
				Validators:          []validator.Int64{int64validator.AtLeast(1)},
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

func (d *DataSource) readPlural(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse, cfg Values) {
	p := d.spec.Plural

	query := url.Values{}
	for k, vs := range p.Extra {
		for _, v := range vs {
			query.Add(k, v)
		}
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
		param, ok := p.Filters[name]
		if !ok {
			resp.Diagnostics.AddError("Invalid filter", fmt.Sprintf("'%s' is not a supported filter parameter", name))
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
	var all []map[string]any
	if d.spec.Custom != nil {
		objs, err := d.spec.Custom(ctx, d.state, cfg, req)
		if err != nil {
			resp.Diagnostics.AddError("Could not read "+p.What, err.Error())
			return
		}
		all = objs
	}
	for d.spec.Custom == nil {
		query.Set("limit", strconv.Itoa(DefaultPageSize))
		query.Set("offset", strconv.Itoa(len(all)))
		page, err := netboxapi.ListPage[map[string]any](ctx, d.state.Rest, d.spec.Path, query)
		if err != nil {
			resp.Diagnostics.AddError("Could not read "+p.What, fmt.Sprintf("failed to fetch %s at offset %d: %s", p.What, len(all), err))
			return
		}
		all = append(all, page.Results...)
		if len(page.Results) == 0 || page.Next == nil || *page.Next == "" {
			break
		}
		if d.spec.Keep == nil && keepRegex == nil && userLimit > 0 && int64(len(all)) >= userLimit {
			break
		}
	}
	if d.spec.Keep != nil || keepRegex != nil {
		kept := all[:0]
		for _, o := range all {
			if d.spec.Keep != nil && !d.spec.Keep(cfg, o) {
				continue
			}
			if keepRegex != nil {
				name, _ := o["name"].(string)
				if !keepRegex.MatchString(name) {
					continue
				}
			}
			kept = append(kept, o)
		}
		all = kept
	}
	if userLimit > 0 && int64(len(all)) > userLimit {
		all = all[:userLimit]
	}

	elem := Attr{Name: p.Result, Kind: NestedList, Sub: d.spec.Attrs, Computed: true}
	items := make([]any, len(all))
	for i, o := range all {
		for _, sc := range d.spec.ResultScopes {
			typ, _ := o[sc.TypeKey].(string)
			for _, opt := range sc.Options {
				if opt.Type == typ && o[sc.IDKey] != nil {
					o[opt.Attr] = o[sc.IDKey]
				}
			}
		}
		items[i] = o
	}
	result := elem.fromJSONList(items)
	for _, a := range d.spec.ExtraFilters {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(a.Name), cfg[a.Name])...)
	}
	if d.spec.CustomFields {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(customFieldsKey), cfg[customFieldsKey])...)
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(uniqueID()))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(p.Result), result)...)
}

// fromJSONList converts a list of objects.
func (a Attr) fromJSONList(items []any) attr.Value {
	return a.fromJSON(items, nil)
}

var idCounter atomic.Int64

// uniqueID returns an ID for data sources that have no natural one.
func uniqueID() string {
	return fmt.Sprintf("terraform-%s-%d", time.Now().UTC().Format("20060102150405"), idCounter.Add(1))
}
