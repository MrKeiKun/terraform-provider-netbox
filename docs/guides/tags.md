---
page_title: "Tags and default_tags"
subcategory: ""
description: |-
  How tags work on resources, and how to add tags to every resource with default_tags.
---

# Tags and `default_tags`

## Tags on a resource

Resources of objects that can have tags take a `tags` set of tag names. Use the name of the tag, not the slug.

```hcl
resource "netbox_site" "example" {
  name = "example"
  tags = ["production", "europe"]
}
```

The tags have to exist in NetBox. A name that is not found, or that matches more than one tag, is an error. Create tags with `netbox_tag` and refer to them by `netbox_tag.example.name`, so Terraform creates the tag first.

## Default tags

`default_tags` on the provider adds tags to every resource that supports them.

```hcl
provider "netbox" {
  server_url   = "https://netbox.example.com"
  api_token    = var.netbox_token
  default_tags = ["managed-by-terraform"]
}
```

## `tags` and `tags_all`

- `tags` holds the tags you wrote in the configuration.
- `tags_all` holds all tags that the object has in NetBox: your tags and the default tags. It is read-only.

A default tag is not shown in `tags`, so adding or removing a default tag does not change your configuration. If you write a tag in `tags` that is also a default tag, it stays in `tags`.

When you change `default_tags`, the plan shows the change in `tags_all` of every resource, and applying it updates the objects in NetBox.

## Tags added somewhere else

A tag that somebody adds to an object in NetBox shows up in `tags_all` after a refresh. It is not in your configuration, so the plan shows a change of `tags_all`, and applying it removes the tag again.

Default tags are not added to `netbox_tag` itself. A resource whose page has no `tags` attribute does not have tags at all.
