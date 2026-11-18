# 1.0.0-rc1 (November 2026)

NOTES

* The provider supports Netbox 4.4.0 to 4.7.5. Older Netbox versions are not supported.
* The provider now talks to Netbox with its own small REST client. The generated `go-netbox` client is no longer used, so new Netbox releases do not need a client update.
* The provider now uses the Terraform Plugin Framework (plugin protocol 6) in place of the plugin SDK. It needs Terraform 1.0 or newer. Existing state upgrades on its own, and resource names and arguments stay the same. See the [upgrade guide](https://registry.terraform.io/providers/mrkeikun/netbox/latest/docs/guides/upgrading-to-plugin-framework) (#12).
* Attributes that hold a secret are now marked sensitive: `api_token` and `headers` of the provider, `auth_key` of an FHRP group, `preshared_key` of an IKE policy, `auth_psk` of a wireless link, `additional_headers` of a webhook and `parameters` of a data source, on the resources and the data sources. An output that passes one of them on now needs `sensitive = true` (#24).
* **Breaking:** `rear_ports` on `netbox_device_front_port` and `netbox_front_port_template`, and `front_ports` on `netbox_device_rear_port` and `netbox_rear_port_template`, are now attributes instead of blocks. Write `rear_ports = [{ ... }]` in place of `rear_ports { ... }` (#12).

ENHANCEMENTS

* Add the resources and data sources the provider was missing for NetBox 4.4.0 to 4.7.5 (#2):
  * 12 objects new in NetBox 4.5 to 4.7, for example `netbox_rack_group`, `netbox_cable_bundle`, `netbox_virtual_machine_type`, `netbox_owner`, `netbox_owner_group`, `netbox_module_bay_type` and the cooling objects.
  * 29 objects from NetBox 4.4 the provider never had, for example l2vpn, the IKE and IPsec objects, the virtual circuit objects, `netbox_provider_account`, `netbox_custom_link`, `netbox_export_template` and `netbox_wireless_link`.
  * About 70 existing resources get `owner_id`, `comments`, `custom_fields` and other fields that were missing.
  * Front and rear ports support the new `rear_ports` and `front_ports` mappings, and services support `port_mappings`.
  * NetBox 4.6.6 to 4.6.10 are added to the supported versions.
* Add the `max_retries` provider setting (or `NETBOX_MAX_RETRIES`), 3 by default. GET, PUT and DELETE requests are sent again after a connection error or a 502, 503 or 504 response. A request that creates or changes an object is only sent again when NetBox did not carry it out: no connection could be made, the response is 429 or 503, or NetBox reports a database deadlock. The wait doubles with every attempt and follows a `Retry-After` header. Set it to `0` to turn retries off (#3, #11, #24, #26).
* Add the data sources `netbox_aggregates`, `netbox_cables`, `netbox_tunnels`, `netbox_device_front_ports`, `netbox_device_rear_ports` and `netbox_device_module_bays` (#3).
* Add `airflow` on devices, tags on IPAM roles and `nat_inside` on `netbox_available_ip_address` (#3).
* `netbox_available_vlan` can now be imported (#5).
* Add import docs for every resource, and fill in missing docs, examples and schema descriptions (#5).
* `server_url` and `api_token` are now optional in the provider schema, since they can come from `NETBOX_SERVER_URL` and `NETBOX_API_TOKEN`. The provider reports an error at startup when neither the provider block nor the environment sets them (#12).
* Add about 170 data sources, generated from the NetBox API, so that almost every object has a data source for one object and one that lists objects. The provider now has 274 data sources (#13).
* The data sources return every field of the NetBox object, for example `created`, `last_updated`, `tags`, `custom_fields` and the counts of related objects. 94 data sources get new attributes. Nothing is renamed or removed (#15).
* Resources have the read-only fields of the NetBox object as computed attributes, for example `created`, `last_updated`, `device_count` and `rack_count` on `netbox_site` (#16).
* Add the resources `netbox_available_asn`, `netbox_bookmark`, `netbox_notification`, `netbox_subscription` and `netbox_table_config` (#13), and `netbox_virtual_chassis_master`, which sets the master device of a virtual chassis (#17).
* `netbox_interface` is also available as `netbox_virtual_machine_interface`, and `netbox_provider_account` as `netbox_circuit_provider_account`. `moved` blocks work between the names. The old names stay, and `netbox_interface` is deprecated (#13).
* Add the data source `netbox_supported_versions`, and the data sources `netbox_virtual_disks`, `netbox_virtual_machine_interfaces` and `netbox_vpn_tunnels` as second names (#13).
* Add tags on 15 more resources and many attributes that were missing (#13), among them `comments` and `owner_id` on `netbox_available_ip_address`, `custom_fields`, `owner_id`, `qinq_role` and `qinq_svlan` on `netbox_available_vlan`, `account_ids` on `netbox_circuit_provider`, `data_source_id` on `netbox_config_context` and `netbox_config_template`, `created_by_id` on `netbox_journal_entry`, `enabled` on `netbox_token` and the read-only `master_id` on `netbox_virtual_chassis` (#14).
* The provider knows the first NetBox version of every resource, attribute, data source and filter. The docs say "Needs NetBox X or later", and a plan against an older server fails with a clear error, where the field was ignored before (#13).
* Plural data sources take the filters of the NetBox API that they were missing (#13), and a filter can end in a lookup expression, as in `name__ic`, `status__n` or `id__gt`. Every plural data source also takes the filter `id`. A filter or lookup that NetBox does not have is an error, because NetBox would ignore it and return every object (#26).
* A resource can be imported by its slug or name as well as by its ID (#19).
* The values of choice attributes are checked against the choices of NetBox. 22 attributes that had no check get one, for example `airflow` and the cooling attributes (#19).
* Reads are batched: resources of the same kind that Terraform reads at the same time are read with one request. A plan over 605 resources takes 105 requests in place of 606 and about a third less time. The new `batch_reads` provider setting (or `NETBOX_BATCH_READS`) turns it off (#24, #26).
* Send fewer requests in other places too (#23, #24):
  * The object is no longer read again after a create or update, which halves the requests of an apply.
  * Tags are looked up once per run and no longer for every resource.
  * Connections to NetBox are kept open and used again.
  * Device and virtual machine resources no longer ask NetBox for the config context, which they do not show.
* Add the `page_size` provider setting (or `NETBOX_PAGE_SIZE`), the number of objects a plural data source asks for with each request. Default 1000 (#23).
* Plural data sources read up to four pages at the same time and need less memory. Reading 30,000 IP addresses takes about a quarter less time (#26).
* A redirect is only followed when it stays on the host in `server_url` and does not go from https to http, so the API token is never sent anywhere else. A redirect that would turn a request that writes into a GET is refused with the address to use, and the provider warns at startup when `server_url` itself redirects (#24, #26).
* TLS 1.2 is the lowest version the provider accepts. A `server_url` with a user name or password, or with a scheme other than http or https, is refused (#24).
* Requests carry a `User-Agent` with the Terraform and provider version (#24).
* When the answer to a create is lost, the error says that NetBox may have created the object anyway (#26).
* Add the guides "Filtering data sources", "Custom fields", "Tags and `default_tags`" and "NetBox versions" (#18), "Changing the provider source" (#20) and "Performance and connection settings" (#24).
* The provider can be used as a Go library through the `provider` package (#13).
* Release files have a build provenance attestation, which can be checked with `gh attestation verify` (#24).

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
* `netbox_interface_template` `channels` and `netbox_webhook` `timeout` are no longer sent as `0` when they are not set. NetBox 4.7 rejects that (#12).
* `netbox_device` `location_id` is cleared when both `location_id` and `rack_id` are removed from the configuration (#12).
* Fix the API field behind `netbox_site` `timezone`, `netbox_power_feed` `max_percent_utilization` and `netbox_contact` `group_id` (#12).
* Fix the build on 32-bit platforms. On those, AS numbers above 2147483647 cannot be used (#8).
* The provider no longer rejects values that NetBox accepts: `paused` as the status of a virtual machine, the `infiniband-*` and other interface types, and the `mdc` and `mu*` port types. `netbox_available_ip_address` checks `status` again (#19).
* A create no longer fails when a proxy in front of NetBox answers 503 or NetBox is briefly unreachable. The request is sent again (#26).
* With `http://` in `server_url` and a web server that redirects to https, a create failed with "the response has no id". The provider now names the address to use (#26).
* `limit = 0` on a plural data source returns all objects, as documented. It was rejected (#26).
* When the API token may not view custom fields, the error for `custom_fields` names the missing permission (#26).
