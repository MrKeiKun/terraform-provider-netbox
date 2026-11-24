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

## Known NetBox problems

These come from NetBox itself. The provider cannot fix them.

### Creating a device fails with 403 Forbidden

Creating a `netbox_device` can fail with `403 Forbidden`, even with a token that has all permissions. This happens when the device type has inventory item templates and some of them have a parent template.

NetBox creates the inventory items of a new device before it saves any of them, so a child item cannot find its parent. NetBox then reports the failed lookup as a permission error.

- It happens on NetBox 4.6.0 to 4.7.5. NetBox 4.4.0 to 4.5.10 are not affected. We did not test 4.6.1 to 4.6.9 and 4.7.1 to 4.7.4.
- The failed lookup is tracked in [netbox-community/netbox#23743](https://github.com/netbox-community/netbox/issues/23743). The wrong 403 error is tracked in [netbox-community/netbox#23765](https://github.com/netbox-community/netbox/issues/23765).
- Until NetBox fixes it, you can use a device type whose inventory item templates have no parent, or a NetBox version that is not affected.

If you get a 403 on a device create, check this first before you change the token permissions.

### A deleted object can come back

When two requests delete a rack and its location at the same moment, NetBox can create the rack again with the same ID, and without its location. The same happens with a LAG interface and its member interfaces, and with a device that is deleted together with its virtual chassis, platform or other objects it points to. Terraform deletes objects in parallel, so it can run into this when it deletes objects that are not linked to each other in the configuration, for example imported objects that use plain IDs.

- It happens on NetBox 4.4.0 to 4.7.5. It is rare on 4.4.0 and common on newer versions.
- It is tracked in [netbox-community/netbox#23766](https://github.com/netbox-community/netbox/issues/23766).
- When a delete reports that other objects depend on an object, run `terraform destroy` again. If the object came back, delete it in NetBox.
- Objects that refer to each other with references in the configuration are deleted in the right order, so they are not affected.

### The list of tagged objects can repeat and miss objects

NetBox does not give the list `extras/tagged-objects/` a stable order. When a read needs more than one page, NetBox can return some objects twice and leave others out. The data source `netbox_tagged_objects` asks for the order of the IDs, so it is not affected. It is tracked in [netbox-community/netbox#23767](https://github.com/netbox-community/netbox/issues/23767).

## Upgrading NetBox

1. Check that the new NetBox version is in the list of your provider release. If it is not, upgrade the provider first.
2. Upgrade NetBox.
3. Run `terraform plan` and read it before you apply. New read-only fields of the newer NetBox can show up in the state.
