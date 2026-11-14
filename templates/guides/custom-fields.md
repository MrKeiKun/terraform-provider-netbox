---
page_title: "Custom fields"
subcategory: ""
description: |-
  How to set, read and filter custom fields.
---

# Custom fields

Resources and data sources of objects that have custom fields in NetBox take a `custom_fields` map. The key is the name of the custom field, as you defined it in NetBox. The custom field has to exist and has to apply to that object type, otherwise NetBox refuses the request.

```hcl
resource "netbox_site" "example" {
  name = "example"

  custom_fields = {
    cost_center = "4711"
    reviewed    = "true"
    contacts    = jsonencode(["anna", "ben"])
  }
}
```

## Values are strings

Every value is a string, whatever the type of the custom field is.

| Type of the custom field | Write the value as                               |
| ------------------------ | ------------------------------------------------ |
| Text, long text, URL     | The text                                         |
| Integer, decimal         | `"42"`, `"3.5"`                                  |
| Boolean                  | `"true"` or `"false"`                            |
| Date                     | `"2026-11-14"`                                   |
| Selection                | The value of the choice (not its label)          |
| Multiple selection, JSON | `jsonencode(...)`                                |
| Object, multiple objects | The ID, or `jsonencode([...])` for a list of IDs |

JSON values are compared by what they contain. Different spacing or key order is not a change.

## Which fields Terraform manages

- Only the fields in your configuration are managed. A field in NetBox that is not in `custom_fields` is left alone, and no difference is shown for it.
- When you remove a field from `custom_fields`, it is cleared in NetBox.
- When you do not write `custom_fields` at all, no field is changed. The state then holds the fields that have a value in NetBox, so you can read them.
- An empty string means the field is not set.

## Reading custom fields

Data sources return `custom_fields` as a map of field name to value. Only the fields that have a value are in it.

```hcl
data "netbox_site" "example" {
  slug = "example"
}

output "cost_center" {
  value = data.netbox_site.example.custom_fields["cost_center"]
}
```

## Finding objects by a custom field

Give `custom_fields` to a data source to filter. See [filtering data sources](filtering-data-sources).
