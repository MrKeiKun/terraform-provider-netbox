// Package fwresource implements NetBox resources and data sources for the Terraform Plugin
// Framework from a declarative description. Nearly every NetBox object is created, read, updated and
// deleted the same way, so a Spec says which attributes there are and how they map to the REST API,
// and this package does the rest.
package fwresource

import (
	"context"

	"github.com/MrKeiKun/terraform-provider-netbox/internal/providerstate"

	"github.com/hashicorp/terraform-plugin-framework/attr"
)

// Kind is the type of an attribute.
type Kind int

const (
	String Kind = iota
	Int
	Float
	Bool
	StringSet
	StringList
	IntSet
	IntList
	// JSON is a string attribute that holds a JSON document. It is sent as JSON and compared by content.
	JSON
	// NestedSet and NestedList hold objects whose attributes are described by Sub. Required nested
	// attributes are written as blocks, computed ones as attributes.
	NestedSet
	NestedList
	// StringMap is a map of strings.
	StringMap
	// IntLists and StringLists are lists of lists of numbers or strings.
	IntLists
	StringLists
)

// Attr describes one attribute of a resource and the field it maps to in the REST API.
type Attr struct {
	Name        string
	Kind        Kind
	Description string

	Required  bool
	Optional  bool
	Computed  bool
	Sensitive bool
	ForceNew  bool

	// Deprecated holds the deprecation message, if the attribute is deprecated.
	Deprecated string

	// Key is the field in the REST API. It defaults to Name.
	Key string
	// WriteOnly attributes are sent to NetBox but never read back, for example passwords.
	WriteOnly bool
	// Volatile marks an Optional and Computed attribute whose value is not kept from the state: it
	// is unknown in a plan with changes when it is not configured.
	Volatile bool
	// Local attributes only exist in Terraform: they are never sent to NetBox and never read back.
	Local bool
	// NoSend attributes are read from NetBox but never sent.
	NoSend bool

	// Ref marks an Int, IntSet or IntList that holds the IDs of other objects. 0 means no object.
	Ref bool
	// CaseInsensitive treats values that only differ in case as equal.
	CaseInsensitive bool
	// LooseJSON sends a string that holds valid JSON as JSON and any other string as it is.
	LooseJSON bool
	// Zero makes a missing value read as the zero value instead of null (data sources).
	Zero bool
	// Time compares the value as a point in time, so that different spellings of it are the same.
	Time bool
	// OmitZero leaves the field out when the attribute is 0, for numbers where 0 is not a valid value.
	OmitZero bool
	// Nullable fields accept null, which is how they are cleared.
	Nullable bool
	// NoZero turns off the default of an optional attribute (empty string, 0, false, empty set).
	// Without it, an optional attribute that is not set reads as that zero value, like in the SDKv2 provider.
	NoZero bool

	// Default is the value used when the attribute is not set: a string, int64, float64 or bool.
	Default any

	OneOf  []string // allowed values of a String
	MinLen int      // minimal length of a String
	MaxLen int      // maximal length of a String
	Min    *int64   // minimal value of an Int
	Max    *int64   // maximal value of an Int
	Regex  string   // pattern a String must match

	// ElemOneOf lists the allowed values of the elements of a set or list of strings.
	ElemOneOf []string

	// Query is the query parameter a data source filter is sent as. It defaults to the API field.
	Query string
	// Path is the path to the field in the REST API, when it is not a field of the object itself: the
	// names of nested objects, and [0] for the first element of a list.
	Path []string
	// IPOnly strips the prefix length of an address.
	IPOnly bool
	// Pairs reads a list of [start, end] pairs as objects with a start and an end.
	Pairs bool
	// ElemKey is the field of the elements of a StringSet, StringList, IntSet or IntList to use when
	// the API returns objects, for example name for tags.
	ElemKey string

	// Sub describes the attributes of the objects in a NestedSet or NestedList.
	Sub []Attr

	// zero is set by withZeroDefaults.
	zero bool
}

func (a Attr) key() string {
	if a.Key != "" {
		return a.Key
	}
	return a.Name
}

// Values holds the values of the attributes of an object by attribute name.
type Values map[string]attr.Value

// Spec describes a resource.
type Spec struct {
	// TypeName is the full name of the resource, for example netbox_manufacturer.
	TypeName    string
	Description string
	// Path is the API collection, for example "dcim/manufacturers/".
	Path    string
	Version int64
	Attrs   []Attr

	// Tags adds the tags and tags_all attributes.
	Tags bool
	// CustomFields adds the custom_fields attribute.
	CustomFields bool
	// SlugFrom names the attribute a slug is generated from when none is given.
	SlugFrom string
	// Scopes map groups of attributes to the type and ID fields NetBox uses for an object of any type.
	Scopes []Scope

	// NoDefaultTags keeps the default tags of the provider off the object, as for tags themselves.
	NoDefaultTags bool
	// Put updates with PUT instead of PATCH.
	Put bool

	ExactlyOneOf [][]string
	AtLeastOneOf [][]string
	// Conflicts lists groups of attributes of which at most one may be set.
	Conflicts [][]string
	// RequiredWith lists groups of attributes that must all be set when one of them is.
	RequiredWith [][]string

	// ModifyBody changes the request body, which already holds the fields of all attributes.
	// prior is nil on create.
	ModifyBody func(ctx context.Context, plan, prior Values, body map[string]any) error
	// Allocate replaces the POST of a create for objects that NetBox picks itself, for example the
	// next free IP address. It gets the request body of the attributes and returns the new object.
	// A normal update with all attributes follows.
	Allocate func(ctx context.Context, st *providerstate.State, plan Values, body map[string]any) (map[string]any, error)
	// Upgraders turn the raw state of an older schema version into the state of the current
	// version, keyed by the old version.
	Upgraders map[int64]func(ctx context.Context, state map[string]any) (map[string]any, error)
	// AfterWrite runs after an object was created (prior is nil) or updated. It tells whether it
	// changed the object, which is then read again.
	AfterWrite func(ctx context.Context, st *providerstate.State, id int64, plan, prior Values) (bool, error)
	// BeforeDelete runs before an object is deleted.
	BeforeDelete func(ctx context.Context, st *providerstate.State, id int64, state Values) error
	// Finish changes the values after they were taken from the API response. Unlike ModifyState it
	// can call the API.
	Finish func(ctx context.Context, st *providerstate.State, obj map[string]any, prior, out Values) error
	// ImportID splits an import ID into the ID of the object and the values of attributes that
	// cannot be read from the API (Int and String values only).
	ImportID func(id string) (objectID string, attrs map[string]any, err error)
	// NoReload skips reading the object again after a write. Use it when the response to the write has
	// something a read does not, like the key of a token that NetBox only shows once.
	NoReload bool
	// ModifyPlan returns planned values to replace, given the configuration and the plan. It lets a
	// computed attribute become null when nothing in the configuration can set it any more.
	ModifyPlan func(config, plan Values) map[string]attr.Value
	// ModifyState changes the values taken from the API response before they go into the state.
	ModifyState func(ctx context.Context, obj map[string]any, prior Values, out Values) error
}

// emptyCollection marks the default of a set or list attribute: the empty collection.
type emptyCollection struct{}

// zeroDefault tells whether an attribute reads as its zero value while it is not set.
func (a Attr) zeroDefault() bool {
	return a.Optional && !a.Required && !a.Computed && a.Default == nil && !a.NoZero && !a.WriteOnly && !a.Local && a.Kind != JSON && a.Kind != NestedSet && a.Kind != NestedList
}

// withZeroDefaults returns a copy of the spec in which optional attributes that are not set read as
// their zero value: an empty string, 0, false or an empty set. Terraform could not tell these from
// unset attributes with the SDKv2 provider either, and the API reports them this way.
func (s *Spec) withZeroDefaults() *Spec {
	c := *s
	c.Attrs = make([]Attr, len(s.Attrs))
	for i, a := range s.Attrs {
		if a.zeroDefault() {
			a.Computed = true
			switch a.Kind {
			case String:
				a.Default = ""
			case Int:
				a.Default = int64(0)
			case Float:
				a.Default = float64(0)
			case Bool:
				a.Default = false
			default:
				a.Default = emptyCollection{}
			}
			a.zero = true
		}
		c.Attrs[i] = a
	}
	return &c
}
