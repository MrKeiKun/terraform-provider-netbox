# Unreleased

NOTES

* The provider supports Netbox 4.4.0 to 4.7.4. Older Netbox versions are not supported.
* The provider now talks to Netbox with its own small REST client. The generated `go-netbox` client is no longer used, so new Netbox releases do not need a client update.

ENHANCEMENTS

* Add the resources and data sources the provider was missing for NetBox 4.4.0 to 4.7.4 (#2):
  * 12 objects new in NetBox 4.5 to 4.7, for example `netbox_rack_group`, `netbox_cable_bundle`, `netbox_virtual_machine_type`, `netbox_owner`, `netbox_owner_group`, `netbox_module_bay_type` and the cooling objects.
  * 29 objects from NetBox 4.4 the provider never had, for example l2vpn, the IKE and IPsec objects, the virtual circuit objects, `netbox_provider_account`, `netbox_custom_link`, `netbox_export_template` and `netbox_wireless_link`.
  * About 70 existing resources get `owner_id`, `comments`, `custom_fields` and other fields that were missing.
  * Front and rear ports support the new `rear_ports` and `front_ports` mappings, and services support `port_mappings`.
  * NetBox 4.6.6 to 4.6.10 are added to the supported versions.
* Add the `max_retries` provider setting (or `NETBOX_MAX_RETRIES`). Failed GET, PUT and DELETE requests are sent again after a connection error or a 502, 503 or 504 response, 3 times by default. Requests that create objects are never repeated. Set it to `0` to turn retries off (#3).
* Add the data sources `netbox_aggregates`, `netbox_cables`, `netbox_tunnels`, `netbox_device_front_ports`, `netbox_device_rear_ports` and `netbox_device_module_bays` (#3).
* Add `airflow` on devices, tags on IPAM roles and `nat_inside` on `netbox_available_ip_address`. Status values are no longer checked by the provider, NetBox does that (#3).
* `netbox_available_vlan` can now be imported (#5).
* Add import docs for every resource, and fill in missing docs, examples and schema descriptions (#5).

BUG FIXES

* Fix `custom_fields` state drift and wrong value types on all resources that support it (#1):
  * Custom fields that are not in the configuration (defaults, empty values) no longer show up as a diff on every plan.
  * Object references, multi-select, multi-object and `{value, label}` select values no longer break the state write.
  * Integer, boolean, decimal, JSON, select and object fields are now sent with the right JSON type instead of always as strings.
  * JSON values with different spacing or key order, and lists of ids written as strings, no longer show as drift.
  * Removing a custom field definition now clears the old value from the state.
  * Removing a key from the configuration no longer clears it in NetBox, it is just no longer managed. To clear a field, set it to `""`.
* Fix `netbox_token` on NetBox 4.5 and later: the token is created in one request, and the key is no longer sent again on update (#3).
* Fix bugs in `netbox_available_ip_address` and `netbox_tenant`, and restore `prefix_id` on `netbox_available_ip_address` import (#3).
* Fix importing the primary IP of a device or virtual machine (#5).
* `netbox_available_prefix` now sends all configured fields when it allocates a prefix, so NetBox rules like a required tenant pass on the first request (#6).
* Fix updating `netbox_device_rear_port` and `netbox_rear_port_template` when a front port is mapped (#7).
