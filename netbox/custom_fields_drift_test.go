package netbox

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Custom field definitions of every type on dcim.site, %[1]s is the name prefix.
const customFieldsDriftDefinitions = `
resource "netbox_custom_field_choice_set" "cs" {
  name          = "%[1]s"
  extra_choices = [["a", "A"], ["b", "B"], ["c", "C"]]
}
resource "netbox_custom_field" "text" {
  name          = "%[1]s_text"
  type          = "text"
  content_types = ["dcim.site"]
}
resource "netbox_custom_field" "int" {
  name          = "%[1]s_int"
  type          = "integer"
  content_types = ["dcim.site"]
}
resource "netbox_custom_field" "bigint" {
  name          = "%[1]s_bigint"
  type          = "integer"
  content_types = ["dcim.site"]
}
resource "netbox_custom_field" "bool" {
  name          = "%[1]s_bool"
  type          = "boolean"
  content_types = ["dcim.site"]
}
resource "netbox_custom_field" "date" {
  name          = "%[1]s_date"
  type          = "date"
  content_types = ["dcim.site"]
}
resource "netbox_custom_field" "url" {
  name          = "%[1]s_url"
  type          = "url"
  content_types = ["dcim.site"]
}
resource "netbox_custom_field" "json" {
  name          = "%[1]s_json"
  type          = "json"
  content_types = ["dcim.site"]
}
resource "netbox_custom_field" "select" {
  name          = "%[1]s_select"
  type          = "select"
  content_types = ["dcim.site"]
  choice_set_id = netbox_custom_field_choice_set.cs.id
}
resource "netbox_custom_field" "multiselect" {
  name          = "%[1]s_multiselect"
  type          = "multiselect"
  content_types = ["dcim.site"]
  choice_set_id = netbox_custom_field_choice_set.cs.id
}
resource "netbox_custom_field" "object" {
  name                = "%[1]s_object"
  type                = "object"
  content_types       = ["dcim.site"]
  related_object_type = "dcim.site"
}
resource "netbox_custom_field" "multiobject" {
  name                = "%[1]s_multiobject"
  type                = "multiobject"
  content_types       = ["dcim.site"]
  related_object_type = "dcim.site"
}
resource "netbox_custom_field" "unused" {
  name          = "%[1]s_unused"
  type          = "text"
  content_types = ["dcim.site"]
}
resource "netbox_custom_field" "with_default" {
  name          = "%[1]s_default"
  type          = "text"
  default       = "dflt"
  content_types = ["dcim.site"]
}
resource "netbox_site" "other" {
  name = "%[1]s-other"
  slug = "%[1]s-other"
}
resource "netbox_site" "other2" {
  name = "%[1]s-other2"
  slug = "%[1]s-other2"
}
`

// Every custom field type, set on a site. After apply the framework re-plans
// and fails if the plan is not empty, so any difference between the
// configured string and what NetBox hands back shows up as drift.
func TestAccNetboxCustomFields_allTypesNoDrift(t *testing.T) {
	n := strings.ReplaceAll(testAccGetTestName("cf_drift_all"), "-", "_")
	defs := fmt.Sprintf(customFieldsDriftDefinitions, n)
	site := func(body string) string {
		return defs + fmt.Sprintf(`
resource "netbox_site" "test" {
  name = "%[1]s"
  slug = "%[1]s"
  depends_on = [
    netbox_custom_field.text, netbox_custom_field.int, netbox_custom_field.bigint,
    netbox_custom_field.bool, netbox_custom_field.date, netbox_custom_field.url,
    netbox_custom_field.json, netbox_custom_field.select, netbox_custom_field.multiselect,
    netbox_custom_field.object, netbox_custom_field.multiobject,
  ]
  custom_fields = {
%[2]s
  }
}`, n, body)
	}
	attr := func(name string) string { return "custom_fields." + n + "_" + name }

	step1 := site(fmt.Sprintf(`
    "%[1]s_text"        = "hello"
    "%[1]s_int"         = "5"
    "%[1]s_bigint"      = "1000000"
    "%[1]s_bool"        = "true"
    "%[1]s_date"        = "2026-01-02"
    "%[1]s_url"         = "https://example.com/x"
    "%[1]s_json"        = jsonencode({ a = 1, b = [1, 2] })
    "%[1]s_select"      = "b"
    "%[1]s_multiselect" = jsonencode(["a", "c"])
    "%[1]s_object"      = netbox_site.other.id
    "%[1]s_multiobject" = jsonencode([netbox_site.other.id, netbox_site.other2.id])
`, n))
	// Change some, clear others with an empty string, formatting differences
	// in the JSON value, bool and select changed, the rest left as it was.
	step2 := site(fmt.Sprintf(`
    "%[1]s_text"        = ""
    "%[1]s_int"         = "6"
    "%[1]s_bigint"      = "1000000"
    "%[1]s_bool"        = "false"
    "%[1]s_date"        = "2026-01-02"
    "%[1]s_url"         = ""
    "%[1]s_json"        = "{\"b\": [1, 2],   \"a\": 1}"
    "%[1]s_select"      = "a"
    "%[1]s_multiselect" = "[\"c\",\"a\"]"
    "%[1]s_object"      = ""
    "%[1]s_multiobject" = jsonencode([netbox_site.other2.id])
`, n))

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: step1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_site.test", attr("text"), "hello"),
					resource.TestCheckResourceAttr("netbox_site.test", attr("int"), "5"),
					resource.TestCheckResourceAttr("netbox_site.test", attr("bigint"), "1000000"),
					resource.TestCheckResourceAttr("netbox_site.test", attr("bool"), "true"),
					resource.TestCheckResourceAttr("netbox_site.test", attr("select"), "b"),
					resource.TestCheckResourceAttr("netbox_site.test", attr("json"), `{"a":1,"b":[1,2]}`),
					resource.TestCheckResourceAttrPair("netbox_site.test", attr("object"), "netbox_site.other", "id"),
					// Fields the configuration does not mention stay out of the state's way.
					resource.TestCheckNoResourceAttr("netbox_site.test", attr("unused")),
				),
			},
			{
				Config:   step1,
				PlanOnly: true,
			},
			{
				Config: step2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("netbox_site.test", attr("int"), "6"),
					resource.TestCheckResourceAttr("netbox_site.test", attr("bool"), "false"),
					resource.TestCheckResourceAttr("netbox_site.test", attr("select"), "a"),
					resource.TestCheckNoResourceAttr("netbox_site.test", attr("text")),
					resource.TestCheckNoResourceAttr("netbox_site.test", attr("object")),
				),
			},
			{
				Config:   step2,
				PlanOnly: true,
			},
		},
	})
}

// NetBox returns every custom field of an object, including ones with a
// default value. Without a custom_fields block that must not show up as a diff.
func TestAccNetboxCustomFields_unmanagedFieldsNoDrift(t *testing.T) {
	n := strings.ReplaceAll(testAccGetTestName("cf_drift_unmanaged"), "-", "_")
	cfg := fmt.Sprintf(customFieldsDriftDefinitions, n) + fmt.Sprintf(`
resource "netbox_site" "test" {
  name = "%[1]s"
  slug = "%[1]s"
  depends_on = [netbox_custom_field.with_default, netbox_custom_field.unused]
}`, n)
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{Config: cfg},
			{Config: cfg, PlanOnly: true},
		},
	})
}

func TestCustomFieldValueToString(t *testing.T) {
	tests := []struct {
		name string
		in   interface{}
		want string
	}{
		{"null", nil, ""},
		{"string", "x", "x"},
		{"bool", true, "true"},
		{"integer", float64(5), "5"},
		{"large integer", float64(1000000), "1000000"},
		{"decimal", 1.5, "1.5"},
		{"select", map[string]interface{}{"value": "a", "label": "A"}, "a"},
		{"object", map[string]interface{}{"id": float64(7), "url": "http://x/7/", "display": "x"}, "7"},
		{"json object", map[string]interface{}{"b": float64(2), "a": float64(1)}, `{"a":1,"b":2}`},
		{"json with id only", map[string]interface{}{"id": float64(7)}, `{"id":7}`},
		{"json with value only", map[string]interface{}{"value": "a"}, `{"value":"a"}`},
		{"multiselect", []interface{}{map[string]interface{}{"value": "a", "label": "A"}, map[string]interface{}{"value": "b", "label": "B"}}, `["a","b"]`},
		{"multiobject", []interface{}{map[string]interface{}{"id": float64(1), "url": "u", "display": "d"}, map[string]interface{}{"id": float64(2), "url": "u", "display": "d"}}, `[1,2]`},
		{"plain list", []interface{}{"x", "y"}, `["x","y"]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := customFieldValueToString(tt.in); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetCustomFieldsDropsUnsetValues(t *testing.T) {
	got := getCustomFields(map[string]interface{}{
		"set":   "v",
		"empty": "",
		"null":  nil,
		"obj":   map[string]interface{}{"id": float64(3), "url": "u"},
		"num":   float64(4),
	})
	want := map[string]interface{}{"set": "v", "obj": "3", "num": "4"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s: got %v, want %v", k, got[k], v)
		}
	}
}

func TestCustomFieldValuesEqual(t *testing.T) {
	if !customFieldValuesEqual(`{"a":1,"b":[1,2]}`, "{\"b\": [1, 2],  \"a\": 1}") {
		t.Error("json with different formatting should be equal")
	}
	if !customFieldValuesEqual("", "  ") {
		t.Error("empty values should be equal")
	}
	if customFieldValuesEqual("5", "6") {
		t.Error("different values should differ")
	}
	if customFieldValuesEqual(`{"a":1}`, `{"a":2}`) {
		t.Error("different json should differ")
	}
}
