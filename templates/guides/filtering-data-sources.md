---
page_title: "Filtering data sources"
subcategory: ""
description: |-
  How to find objects with the singular and plural data sources.
---

# Filtering data sources

Most objects have two data sources. The singular one, for example `netbox_site`, returns exactly one object. The plural one, for example `netbox_sites`, returns a list.

## Singular data sources

Set one or more of the optional arguments. The data source sends them to NetBox as filters.

```hcl
data "netbox_site" "by_slug" {
  slug = "example-site-1"
}
```

The lookup fails when nothing matches, and when more than one object matches. Add another argument to narrow it down.

Every attribute of the object is returned, also the ones you did not use as a filter. This includes `created`, `last_updated`, `custom_fields` and `tags`.

## Plural data sources

Use `filter` blocks. Each block has a `name` and a `value`. The name is a filter of the NetBox API for that object type, as you find it in the NetBox REST API documentation.

```hcl
data "netbox_devices" "core" {
  filter {
    name  = "site"
    value = "example-site-1"
  }
  filter {
    name  = "status"
    value = "active"
  }
}

output "device_names" {
  value = data.netbox_devices.core.devices[*].name
}
```

- A name that the data source does not support is an error.
- Several blocks have to match all together.
- For filters that take a list in the NetBox API, a value can hold several choices, separated by commas. `"active,planned"` returns objects that have either status.
- An empty value is ignored.
- `limit` caps the number of results. Without it, all matching objects are returned. The provider asks NetBox for them page by page.
- Many plural data sources also have `name_regex`, which keeps the objects whose name matches a regular expression. The matching happens in Terraform, after NetBox returned the objects, so narrow the result with a `filter` first when there are many objects.

The objects are in a list attribute named after the data source, here `devices`. Each element has the same attributes as the singular data source.

## Filtering on custom fields

The data sources of objects that have custom fields take `custom_fields`, singular and plural alike. NetBox is asked for objects whose custom field has the value.

```hcl
data "netbox_sites" "owned" {
  custom_fields = {
    cost_center = "4711"
  }
}
```

See the [custom fields guide](custom-fields) for how the values are written.

## Filters that need a newer NetBox

Some filters only exist in newer versions of NetBox. When the server is too old, the data source stops with an error that names the filter and the version it needs, instead of sending a request that NetBox would refuse.

## The same lookup, different results over time

A data source is read again on every plan. If the objects that match change in NetBox, the plan shows the difference. For a singular data source that can turn into an error, when a second object starts to match.
