---
page_title: "NetBox versions"
subcategory: ""
description: |-
  Which NetBox versions the provider supports, and what happens with older and newer ones.
---

# NetBox versions

Each release of the provider is tested against a list of NetBox versions. The oldest is NetBox 4.4.0. The newest is listed in the table on the provider page. The data source `netbox_supported_versions` returns the whole list of the release you use.

## What happens at startup

The provider asks NetBox for its version when it starts.

- A version in the list works without any message.
- A version that is not in the list gets a warning. It is not an error. It usually works, but nobody tested it.
- `skip_version_check = true` turns the check off, and so the warning. It also saves one request, which helps when you only run a plan without a server.

## Objects and attributes that need a newer NetBox

NetBox adds objects and fields in newer versions. When your configuration uses one that your server does not have, the plan stops with an error such as:

> The attribute owner_id of netbox_site needs NetBox 4.5.0 or later, and the server runs 4.4.10. NetBox would ignore it.

Without this check, the old NetBox would silently ignore the field, and the next plan would show the difference again.

The same check exists for filters of plural data sources. See [filtering data sources](filtering-data-sources).

This only happens for attributes you set. Leave them out and the resource works on every supported version.

## Writing configuration for several versions

If a module has to work on several NetBox versions, make the newer attributes optional in the module and pass `null` when they are not wanted. A `null` value counts as not set, so the check does not fire.

```hcl
variable "owner_id" {
  type    = number
  default = null
}

resource "netbox_site" "example" {
  name     = "example"
  owner_id = var.owner_id
}
```

## Upgrading NetBox

1. Check that the new NetBox version is in the list of your provider release. If it is not, upgrade the provider first.
2. Upgrade NetBox.
3. Run `terraform plan` and read it before you apply. New read-only fields of the newer NetBox can show up in the state.
