# Unreleased

NOTES

* The provider supports Netbox 4.4.0 to 4.7.4. Older Netbox versions are not supported.
* The provider now talks to Netbox with its own small REST client. The generated `go-netbox` client is no longer used, so new Netbox releases do not need a client update.

ENHANCEMENTS

* Add the resources and data sources the provider was missing for NetBox 4.4.0 to 4.7.4 ([#2](https://github.com/MrKeiKun/terraform-provider-netbox/pull/2)):
  * 12 objects new in NetBox 4.5 to 4.7, for example `netbox_rack_group`, `netbox_cable_bundle`, `netbox_virtual_machine_type`, `netbox_owner`, `netbox_owner_group`, `netbox_module_bay_type` and the cooling objects.
  * 29 objects from NetBox 4.4 the provider never had, for example l2vpn, the IKE and IPsec objects, the virtual circuit objects, `netbox_provider_account`, `netbox_custom_link`, `netbox_export_template` and `netbox_wireless_link`.
  * About 70 existing resources get `owner_id`, `comments`, `custom_fields` and other fields that were missing.
  * Front and rear ports support the new `rear_ports` and `front_ports` mappings, and services support `port_mappings`.
  * NetBox 4.6.6 to 4.6.10 are added to the supported versions.

BUG FIXES

* Fix `custom_fields` state drift and wrong value types on all resources that support it ([#1](https://github.com/MrKeiKun/terraform-provider-netbox/pull/1)):
  * Custom fields that are not in the configuration (defaults, empty values) no longer show up as a diff on every plan.
  * Object references, multi-select, multi-object and `{value, label}` select values no longer break the state write.
  * Integer, boolean, decimal, JSON, select and object fields are now sent with the right JSON type instead of always as strings.
  * JSON values with different spacing or key order, and lists of ids written as strings, no longer show as drift.
  * Removing a custom field definition now clears the old value from the state.
  * Removing a key from the configuration no longer clears it in NetBox, it is just no longer managed. To clear a field, set it to `""`.
