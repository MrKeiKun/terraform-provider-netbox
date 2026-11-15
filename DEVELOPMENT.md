# Development guide

This is for people who change the provider itself. For using the provider, see the [documentation](https://registry.terraform.io/providers/mrkeikun/netbox/latest/docs). For building and testing, see the README.

## How the provider is built

The provider uses the Terraform Plugin Framework and serves plugin protocol 6 directly from `main.go`. Almost every NetBox object is created, read, updated and deleted the same way, so resources and data sources are not written out one by one. Each one is described by a `Spec`, and a shared engine does the work.

| Path                     | What it does                                                                                          |
| ------------------------ | ----------------------------------------------------------------------------------------------------- |
| `internal/fwresource`    | The engine. It turns a `Spec` into a schema and into create, read, update, delete and import.         |
| `internal/fwprovider`    | The provider, its configuration and every `Spec`.                                                     |
| `internal/netboxapi`     | A small REST client for NetBox. Objects are passed around as `map[string]any`.                        |
| `internal/providerstate` | The shared client, the NetBox version and the default tags, handed to every resource.                 |
| `netbox/`                | Acceptance tests only. There is no provider code in this folder.                                      |
| `provider/`              | The provider as a Go library, for tools that serve it in process. It only passes on `fwprovider.New`. |
| `tools/`                 | `netboxdrift` compares the NetBox API with the specs, `gendatasources` writes data sources.           |

The specs in `internal/fwprovider` are plain source files that you edit by hand:

- `resources_standard.go` and `datasources_standard.go` hold the objects that follow the usual pattern, where attributes map one to one onto fields of the REST API.
- `overrides.go`, `overrides_ds.go` and `overrides_custom_ds.go` adjust a standard spec that needs a little more, for example another API field name or a hook.
- `resources_manual.go`, `resources_custom.go`, `resources_device.go`, `resources_virtualization.go`, `resources_available.go` and `datasources_custom.go` hold the objects that need more than that.
- `datasources_generated.go`, `datasources_filters_generated.go` and `versions_generated.go` are written by the tools in `tools/`. Do not edit them, see the next section.
- `resources_aliases.go` and `datasources_aliases.go` hold the resources and data sources that are known by a second name.

## Describing a resource

```go
{
    TypeName:     "netbox_tenant",
    Path:         "tenancy/tenants/",
    Description:  ":meta:subcategory:Tenancy:From the official documentation ...",
    Tags:         true,
    CustomFields: true,
    SlugFrom:     "name",
    Attrs: []fwresource.Attr{
        {Name: "name", Kind: fwresource.String, Required: true, Description: "The name of the tenant."},
        {Name: "group_id", Kind: fwresource.Int, Key: "group", Ref: true, Nullable: true, Optional: true, Description: "The ID of the group."},
    },
},
```

The documentation page is built from the descriptions, and `:meta:subcategory:` puts the page in its section. Settings of an attribute that are easy to get wrong:

- `Key` is the field name in the REST API when it differs from the attribute name. Check the real name in the NetBox OpenAPI schema. A wrong name sends nothing and fails silently.
- `Ref` marks the ID of another object. `0` means no object.
- An optional attribute that is not set reads as `""`, `0`, `false` or an empty set in the state. This keeps the state compatible with the earlier SDK-based provider. `NoZero` turns it off.
- `Nullable` makes an unset optional number go to NetBox as `null`. Use it wherever NetBox does not accept `0`, such as `timeout`, `channels` and `memory_mb`. `OmitZero` leaves the field out instead.
- `Computed` is for values NetBox picks. Together with `Optional`, the value in the state must equal the configured one, or Terraform reports an inconsistent result. If NetBox derives the value from another field, as it does for `location` from the rack, use `Computed` and a `ModifyPlan` hook.
- `Volatile` is an optional and computed attribute that is unknown in a plan with changes, so it is not kept from the previous state.
- `Local` exists only in Terraform and is never sent or read. `WriteOnly` is sent but never read back, for example passwords.

The hooks on a `Spec` cover the rest: `ModifyBody`, `ModifyPlan`, `ModifyState`, `Finish`, `AfterWrite`, `BeforeDelete`, `Allocate` (NetBox picks the object, as with the available IP, prefix and VLAN resources), `ImportID` and `Upgraders`. Each has a short comment in `internal/fwresource/spec.go`. Resources that assign a field of an existing object, such as a primary IP, use `AssignSpec` in `assign.go`.

State upgrades live in `Upgraders`, keyed by the schema version, and work on the raw JSON of the old state. See `upgrade.go` for the existing ones and `upgrade_server_test.go` for how to test one through the provider server.

## Adding a resource or data source

1. Add the `Spec` to `resources_standard.go` or `datasources_standard.go`, keeping the list in alphabetical order. Write more code only if the object needs a hook.
2. Add `examples/resources/<type name>/resource.tf` and `import.sh`. `scripts/ensure_docs_examples.sh` fails when the example is missing.
3. Add an acceptance test in `netbox/` that creates, updates and imports the resource. A test that only creates is not enough.
4. Run `make docs`, then commit the changes in `docs/`.
5. Add a line to `CHANGELOG.md` under `# Unreleased`.

## Testing

Unit tests for the engine and the provider run with `make test`. Acceptance tests run against a real NetBox in docker, see the README.

- Tests use `terraform-plugin-testing` with `ProtoV6ProviderFactories`, as in `netbox/provider_test.go`.
- `NETBOX_VERSION` must be the version of the NetBox that is actually running. The tests use it to skip what an older NetBox does not have. If it names a newer version than the container, those tests fail with 404 errors, and if it names an older one they skip and hide real failures. Check the container with `/api/status/`.
- Run the suite against the oldest and the newest supported NetBox. Some objects and field rules exist only in newer versions. For example NetBox 4.7 rejects `0` for `timeout`, and only 4.7 has the cooling objects.
- Leftover data from a failed test can break later runs, for example a duplicate prefix or VLAN. Delete it and run again.
- To check an upgrade from an earlier release, apply a configuration with the old provider binary, then plan and apply with the new one. The result should be "No changes". Terraform's `dev_overrides` makes this easy.

## Generated code

NetBox has more objects than are worth writing out by hand, so part of the provider comes from the API schema that every NetBox serves at `/api/schema/`. Two tools in `tools/` read it. Both need the schema of a running NetBox. Start one with `make docker-up`, which serves it at `http://localhost:8001`, or give a saved copy with `-schema file.json`.

| Tool             | What it writes                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `gendatasources` | `datasources_generated.go` (a data source for one object and one that lists objects, for every collection that has none), `datasources_filters_generated.go` (filters that the list endpoint has and the hand written data sources lack), `datasources_fields_generated.go` (the fields of the object that the attributes of a hand written data source do not cover, so every data source returns the whole object), `resources_fields_generated.go` (the read-only fields of an object, such as `created`, `last_updated` and the counts of related objects, as computed attributes of the resources), and an example for every data source that has none. |
| `netboxdrift`    | The drift report, and with `-write-versions` the file `versions_generated.go`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |

Do not edit the generated files. To change a generated data source, write a `Spec` for it by hand in `datasources_standard.go`. The generator leaves a collection alone when a hand written data source of that kind exists. A generated example is listed in `tools/gendatasources/generated-examples.txt`. Take the name out of the list, or write the example before the first run, to keep your own.

Run the generator again when you add or change a hand written data source, and when you move to a newer NetBox. CI runs it and fails when the files in the repository differ from what it writes.

```shell
go run ./tools/gendatasources -url http://localhost:8001
```

### Drift report

`make drift` lists what the API has and the provider does not: collections without a resource or data source, writable fields without an attribute, list filters without a filter, and string attributes whose NetBox field takes a fixed set of choices but that the provider does not check or checks against a list that lacks one of them (`enum`). Known gaps are in `tools/netboxdrift/baseline.txt`, and CI only fails for a gap that is not in the file. After you close a gap, or decide to leave one, write the file again:

```shell
go run ./tools/netboxdrift -url http://localhost:8001 -baseline tools/netboxdrift/baseline.txt -write-baseline
```

### Versions

NetBox adds fields and objects in every release. A field that the server does not have is silently ignored, which causes a plan that never gets empty. `versions_generated.go` records the first NetBox version that has each resource, attribute, data source and data source filter. The provider adds "Needs NetBox X or later" to the documentation and stops with an error when a configuration uses something that the server is too old for.

The file is made from the schema of the newest patch release of every minor release. Fetch them and write the file again after you add a spec and when a new minor release comes out:

```shell
tools/netboxdrift/fetch-schemas.sh /tmp/schemas
go run ./tools/netboxdrift -write-versions internal/fwprovider/versions_generated.go \
  -versions "4.4.0=/tmp/schemas/schema-v4.4.10.json,4.5.0=/tmp/schemas/schema-v4.5.10.json,4.6.0=/tmp/schemas/schema-v4.6.10.json,4.7.0=/tmp/schemas/schema-v4.7.5.json"
```

The version of an object is the first minor release that has it, so a patch release inside a minor release is not told apart.

### Attributes that NetBox also fills in from the other side

A relation that has two sides, such as the sites of an AS number and the AS numbers of a site, shows up as a change in the plan when it is set on one resource and the other one lists nothing. Make the attribute `Optional` and `Computed`, so that an unset value keeps what NetBox returns, and write in the description that the relation should be set on one side only. The acceptance tests catch this when both resources are in one configuration.
