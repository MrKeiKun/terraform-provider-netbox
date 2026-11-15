---
page_title: "Changing the provider source"
subcategory: ""
description: |-
  How to move existing Terraform state to this provider without recreating any object.
---

# Changing the provider source

If your state was created with another build of the NetBox provider, you can switch to this one without recreating any object in NetBox. The resource names and arguments are the same, and the state is upgraded when you run Terraform the first time.

## Before you start

- Back up your state file, or use a backend that keeps versions.
- Check the [NetBox versions guide](netbox-versions), and read the [upgrade guide](upgrading-to-plugin-framework) if the state comes from a provider built on the plugin SDK.
- Run `terraform plan` once with the old provider, and apply or discard what it shows. You then start from a plan without changes.

## Steps

1. Change the source in `required_providers`:

   ```hcl
   terraform {
     required_providers {
       netbox = {
         source  = "mrkeikun/netbox"
         version = "~> 1.0"
       }
     }
   }
   ```

2. Move the state to the new source. Replace `<old source>` with the source your configuration used before, for example `registry.terraform.io/<namespace>/netbox`:

   ```sh
   terraform state replace-provider "<old source>" "registry.terraform.io/mrkeikun/netbox"
   ```

   Terraform shows what it changes and asks for confirmation.

3. Install the new provider:

   ```sh
   terraform init -upgrade
   ```

4. Run `terraform plan`. Read it before you apply. Read-only attributes that are new can show up in the state, and the `rear_ports` and `front_ports` blocks have to be written as attributes. No object should be replaced.

## If the plan wants to replace something

Do not apply. Open an issue with the resource, the attribute that causes the replacement, and the provider version you came from. Until it is fixed, you can keep using the old provider for that state.

## Many workspaces

Run the same steps in each workspace, one at a time, starting with one that has few objects. Each workspace has its own state, so each one needs its own `replace-provider`.
