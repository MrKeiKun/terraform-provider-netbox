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
