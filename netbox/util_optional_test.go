package netbox

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func optionalTestResource(t *testing.T, raw map[string]interface{}) *schema.ResourceData {
	t.Helper()
	s := map[string]*schema.Schema{
		"text":  {Type: schema.TypeString, Optional: true},
		"ids":   {Type: schema.TypeSet, Optional: true, Elem: &schema.Schema{Type: schema.TypeInt}},
		"names": {Type: schema.TypeSet, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
		"doc":   {Type: schema.TypeString, Optional: true},
	}
	return schema.TestResourceDataRaw(t, s, raw)
}

func TestGetOptionalStrPtr(t *testing.T) {
	d := optionalTestResource(t, map[string]interface{}{"text": "abc"})
	if v := getOptionalStrPtr(d, "text"); assert.NotNil(t, v) {
		assert.Equal(t, "abc", *v)
	}

	d = optionalTestResource(t, map[string]interface{}{})
	assert.Nil(t, getOptionalStrPtr(d, "text"))
}

func TestGetOptionalStrClear(t *testing.T) {
	d := optionalTestResource(t, map[string]interface{}{"text": "abc"})
	if v := getOptionalStrClear(d, "text"); assert.NotNil(t, v) {
		assert.Equal(t, "abc", *v)
	}

	// Not set and not changed: the field is left out of the request.
	d = optionalTestResource(t, map[string]interface{}{})
	assert.Nil(t, getOptionalStrClear(d, "text"))
}

func TestSetBodyStr(t *testing.T) {
	body := map[string]interface{}{}
	setBodyStr(body, optionalTestResource(t, map[string]interface{}{}), "text", "api_text")
	assert.NotContains(t, body, "api_text")

	setBodyStr(body, optionalTestResource(t, map[string]interface{}{"text": "x"}), "text", "api_text")
	assert.Equal(t, "x", body["api_text"])
}

func TestGetIDSet(t *testing.T) {
	d := optionalTestResource(t, map[string]interface{}{"ids": []interface{}{3, 1, 2}})
	assert.ElementsMatch(t, []int64{1, 2, 3}, getIDSet(d, "ids"))

	// An empty set is an empty list and not nil, so that it clears the relation.
	d = optionalTestResource(t, map[string]interface{}{})
	ids := getIDSet(d, "ids")
	assert.NotNil(t, ids)
	assert.Empty(t, ids)
}

func TestGetStringSet(t *testing.T) {
	d := optionalTestResource(t, map[string]interface{}{"names": []interface{}{"b", "a"}})
	assert.ElementsMatch(t, []string{"a", "b"}, getStringSet(d, "names"))

	d = optionalTestResource(t, map[string]interface{}{})
	out := getStringSet(d, "names")
	assert.NotNil(t, out)
	assert.Empty(t, out)
}

func TestJSONRawFromResourceData(t *testing.T) {
	d := optionalTestResource(t, map[string]interface{}{"doc": `{"a": 1}`})
	raw, err := jsonRawFromResourceData(d, "doc")
	assert.NoError(t, err)
	if assert.NotNil(t, raw) {
		assert.JSONEq(t, `{"a": 1}`, string(*raw))
	}

	// No value is sent as null, which clears the field.
	d = optionalTestResource(t, map[string]interface{}{})
	raw, err = jsonRawFromResourceData(d, "doc")
	assert.NoError(t, err)
	assert.Nil(t, raw)
}

func TestJSONString(t *testing.T) {
	assert.Equal(t, "", jsonString(nil))
	assert.Equal(t, "", jsonString(json.RawMessage("null")))
	assert.Equal(t, `{"a":1}`, jsonString(json.RawMessage(`{"a":1}`)))
}

func TestSuppressEquivalentJSON(t *testing.T) {
	assert.True(t, suppressEquivalentJSON("", `{"a": 1}`, `{"a":1}`, nil))
	assert.False(t, suppressEquivalentJSON("", `{"a": 1}`, `{"a":2}`, nil))
}
