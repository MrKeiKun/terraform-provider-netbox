package netboxapi

// Ref is a nested object in a NetBox response. Most nested objects carry
// more fields, but resources usually only need the ID.
type Ref struct {
	ID int64 `json:"id"`
}

// Choice is a choice field in a NetBox response, for example a status.
type Choice struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// TagRef identifies a tag in requests and responses.
type TagRef struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// RefID returns the ID of a nested object, or nil if it is not set. It is
// meant for d.Set, where nil clears the attribute.
func RefID(r *Ref) any {
	if r == nil {
		return nil
	}
	return r.ID
}

// IDs returns the IDs of the given nested objects.
func IDs(refs []Ref) []int64 {
	ids := make([]int64, 0, len(refs))
	for _, r := range refs {
		ids = append(ids, r.ID)
	}
	return ids
}

// ChoiceValue returns the value of a choice field, or an empty string if it is
// not set.
func ChoiceValue(c *Choice) string {
	if c == nil {
		return ""
	}
	return c.Value
}

// IntChoice is a choice field whose value is a number, for example a port speed.
type IntChoice struct {
	Value int64  `json:"value"`
	Label string `json:"label"`
}

// IntChoiceValue returns the value of a numeric choice field, or nil if it is
// not set. It is meant for d.Set, where nil clears the attribute.
func IntChoiceValue(c *IntChoice) any {
	if c == nil {
		return nil
	}
	return c.Value
}
