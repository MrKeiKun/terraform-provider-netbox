---
page_title: "Importing existing objects"
subcategory: ""
description: |-
  How to bring objects that already exist in NetBox under Terraform, and what to check before you apply.
---

# Importing existing objects

You can bring objects that already exist in NetBox under Terraform with `import` blocks. Terraform can write the configuration for you.

```terraform
import {
  to = netbox_site.main
  id = "12"
}
```

```shell
terraform plan -generate-config-out=generated.tf
```

Read `generated.tf`, move what you want to keep into your configuration, and run `terraform plan` again. The plan should show only the import. Then apply.

Use Terraform 1.13 or newer for `-generate-config-out` on a large number of imports. Terraform 1.5 can crash on it.

## Check the plan before you apply

The generated configuration only holds what Terraform can read back, so the first plan can show changes. Read them before you apply.

- **Secrets are not in the generated configuration, and they stay in NetBox.** Terraform does not write sensitive values, and the attributes that hold secrets keep the value NetBox returns while your configuration does not mention them. This covers `netbox_ike_policy` `preshared_key`, `netbox_fhrp_group` `auth_key`, `netbox_webhook` `additional_headers`, `netbox_wireless_lan` and `netbox_wireless_link` `auth_psk`, and `netbox_data_source` `parameters`. To change one, write the new value. To clear one, write `""`. `netbox_webhook` `secret` is never returned by NetBox, so it is only set when you write it.
- **A user is imported without its password.** NetBox never returns it. `password` is only needed to create a user, so an imported user needs no password in the configuration.
- **A rack with a rack type does not hold its dimensions.** NetBox takes `form_factor`, `width` and `u_height` from the rack type, so they are empty in the state and the generated configuration. Read them from the data source.
- **Fields that are not set show as empty values.** An unset number is written as `0`, an unset flag as `false` and an unset text as `""`. The provider reads these as not set, so you can keep them or remove them.
- **Custom fields are adopted.** After an import the state holds every custom field that has a value, so you can read them. Only the fields in your configuration are managed. See [Custom fields](custom-fields).
- **Some computed values show as `(known after apply)`.** The first plan after an import can show an update for an object without a visible change. One apply settles it, and the next plan is clean.

## Moving an object to another address

Use a `moved` block or `terraform state mv` to rename an object or put it in a module. Neither one sends a request to NetBox.

```terraform
moved {
  from = netbox_site.main
  to   = module.dc.netbox_site.main
}
```
