---
page_title: "Upgrading to the Plugin Framework"
subcategory: ""
description: |-
  What changes when you upgrade from a provider version built on the Terraform plugin SDK.
---

# Upgrading to the Plugin Framework

The provider now uses the Terraform Plugin Framework (plugin protocol 6). It needs Terraform 1.0 or newer.

Existing state upgrades on its own, and resource names and arguments stay the same. These things are different.

## `rear_ports` and `front_ports` are attributes

The mapping between front and rear ports is now an attribute and no longer a block. This applies to:

- `rear_ports` on `netbox_device_front_port` and `netbox_front_port_template`
- `front_ports` on `netbox_device_rear_port` and `netbox_rear_port_template`

Add an equals sign and square brackets:

```hcl
# before
rear_ports {
  rear_port_id = netbox_device_rear_port.example.id
  position     = 1
}

# after
rear_ports = [{
  rear_port_id = netbox_device_rear_port.example.id
  position     = 1
}]
```

With more than one mapping, list each one as an object inside the brackets, separated by commas.

## Provider settings

`server_url` and `api_token` are the only provider settings whose requirement changed. They used to be required in the provider schema and are now optional there, since the Plugin Framework cannot fall back to an environment variable inside the schema. They can still come from the `NETBOX_SERVER_URL` and `NETBOX_API_TOKEN` environment variables, and the provider reports an error when neither the provider block nor the environment sets them. All other settings are unchanged.

## `tags_all` is gone from `netbox_config_context`

The tags of a config context choose which objects it applies to, so they are not tags on the config context itself. `netbox_config_context` used to have a computed `tags_all` that only repeated `tags`. It is removed. Use `tags`.

## Tags on more resources

These resources now take `tags` and have `tags_all`, like the others: `netbox_circuit`, `netbox_circuit_provider`, `netbox_circuit_type`, `netbox_cluster_group`, `netbox_cluster_type`, `netbox_contact_assignment`, `netbox_contact_group`, `netbox_contact_role`, `netbox_manufacturer`, `netbox_region`, `netbox_rir`, `netbox_site_group`, `netbox_tenant_group`, `netbox_vpn_tunnel_group` and `netbox_webhook`.

If you set `default_tags` on the provider, those tags are added to these resources on the next apply. The plan shows this as a change in `tags_all`.

## New and renamed resources

- `netbox_virtual_machine_interface` is the same resource as `netbox_interface`. `netbox_interface` keeps working but is deprecated, so Terraform warns when you use it. Both names can be used in the same configuration. To switch an existing interface to the new name without replacing it, use a `moved` block:

  ```hcl
  moved {
    from = netbox_interface.eth0
    to   = netbox_virtual_machine_interface.eth0
  }
  ```

- `netbox_circuit_provider_account` is the same resource as `netbox_provider_account`. Both names stay. A `moved` block works the same way for it.
- `netbox_available_asn` lets NetBox pick the next free AS number of an ASN range. Set `asn_range_id` and read the number from `asn`.
- `netbox_bookmark`, `netbox_notification`, `netbox_subscription` and `netbox_table_config` manage the objects that belong to a NetBox user. Image attachments are not available, since NetBox only takes them as a file upload.

New attributes were added to existing resources, for example `owner_id` on `netbox_device` and `netbox_virtual_machine`, `parent_id` on `netbox_device_role` and `netbox_platform`, and `import_targets` and `export_targets` on `netbox_vrf`. All of them are optional. The `owner_id` attribute and the module bay types need a NetBox version that has them. The resource pages say which.

## New data sources

Nearly every NetBox object now has a data source. Where a resource has a single object, the data source is named like the resource, for example `netbox_device_interface`. The data source that lists objects has a plural name, for example `netbox_device_interfaces`.

A data source for one object finds it by `name` and `slug`, or by the other values that identify it. It fails when no object, or more than one object, matches. A data source that lists objects takes `filter` blocks and a `limit`. Every filter of the NetBox API list endpoint can be used:

```hcl
data "netbox_devices" "edge" {
  filter {
    name  = "site_id"
    value = "3"
  }
  filter {
    name  = "status"
    value = "active"
  }
  limit = 50
}
```
