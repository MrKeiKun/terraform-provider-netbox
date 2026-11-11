# terraform-provider-netbox

[![CI](https://github.com/MrKeiKun/terraform-provider-netbox/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/MrKeiKun/terraform-provider-netbox/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/MrKeiKun/terraform-provider-netbox)](https://github.com/MrKeiKun/terraform-provider-netbox/releases/latest)
[![Terraform Registry](https://img.shields.io/badge/registry-mrkeikun%2Fnetbox-blue?logo=terraform)](https://registry.terraform.io/providers/mrkeikun/netbox/latest)
[![Coverage](https://codecov.io/gh/MrKeiKun/terraform-provider-netbox/graph/badge.svg)](https://codecov.io/gh/MrKeiKun/terraform-provider-netbox)
[![Go version](https://img.shields.io/github/go-mod/go-version/MrKeiKun/terraform-provider-netbox)](go.mod)
[![License: MPL 2.0](https://img.shields.io/github/license/MrKeiKun/terraform-provider-netbox)](LICENSE)

The Terraform Netbox provider is a plugin for Terraform that allows for the full lifecycle management of [Netbox](https://netboxlabs.com/docs/netbox/) resources.
This provider is maintained by [MrKeiKun](https://github.com/MrKeiKun).

See: [Official documentation](https://registry.terraform.io/providers/mrkeikun/netbox/latest/docs) in the Terraform registry.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0 (the provider uses plugin protocol 6)
- [Go](https://go.dev/dl/) 1.26 or newer, only if you build the provider yourself (see `go.mod`)

## Supported netbox versions

Netbox often makes breaking API changes even in non-major releases. Check the table below to see which version a provider was tested against. It is generally recommended to use the provider version matching your Netbox version. We aim to always support the latest minor version of Netbox. The oldest supported release is Netbox 4.4.0, Netbox 4.3 and older are no longer supported.

Each version of the provider has a built-in list of all Netbox versions it supports at release time. Upon initialization, the provider will probe your Netbox version and include a (non-blocking) warning if the used Netbox version is not supported.

| Netbox version | Provider version |
| -------------- | ---------------- |
| v4.4.0 - 4.7.4 | v1.0.0 and up    |

## Upgrading

The provider now uses the Terraform Plugin Framework. See the [upgrade guide](docs/guides/upgrading-to-plugin-framework.md) for the few changes to configurations, such as `rear_ports` and `front_ports` becoming attributes.

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```sh
go install
```

## Installation

Terraform downloads the provider from the registry. Add it to your configuration and run `terraform init`:

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

For further information on how to use third party providers, see the [Terraform documentation](https://developer.hashicorp.com/terraform/language/providers).

Signed binaries for Linux, macOS and Windows are also available on the [releases page](https://github.com/MrKeiKun/terraform-provider-netbox/releases).

## Using the provider

Here is a short example on how to use this provider:

```hcl
provider "netbox" {
  server_url = "https://demo.netbox.dev"
  api_token  = "<your api token>"
}

resource "netbox_platform" "testplatform" {
  name = "my-test-platform"
}
```

For a more examples, see the [provider documentation](https://registry.terraform.io/providers/mrkeikun/netbox/latest/docs).

## Developing the Provider

If you wish to work on the provider, you need [Go](http://www.golang.org) installed on your machine.

See [DEVELOPMENT.md](DEVELOPMENT.md) for how the provider is built and how to add a resource.

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `make docs`. The files in `docs/` are generated, so do not edit them by hand. CI fails if they are out of date.

In order to run the suite of unit tests, run `make test`.

In order to run the full suite of acceptance tests, run `make testacc`.

In order to run a single specific acceptance test, run

```sh
TEST_FUNC=<test_name> make testacc-specific-test
```

For example:

```sh
TEST_FUNC=TestAccNetboxLocationDataSource_basic make testacc-specific-test
```

_Note:_ Acceptance tests create a docker compose stack on port 8001. They run against Netbox v4.7.4 by default. To test another version, set `NETBOX_VERSION`:

```sh
make testacc NETBOX_VERSION=v4.6.0
```

If you notice a failed test, it might be due to a stale netbox data volume. Before concluding there is a problem,
refresh the docker containers by running `make docker-down`. Then run the tests again.

If you get `too many open files` errors when running the acceptance test suite locally on Linux, your user limit for open file descriptors might be too low. You can increase that limit with `ulimit -n 2048`.

## Contribution

Contributions are welcome. Before opening a PR:

- Run `make test` and `make docs`, and add or update acceptance tests for what you change.
- Run `pre-commit run --all-files` (see `.pre-commit-config.yaml`). CI runs the same checks, plus `golangci-lint`. CodeQL scans run as a separate check.
- Add a line for your change under `# Unreleased` in `CHANGELOG.md`.

## Releasing

Maintainers release by tagging the latest commit on `master` with a signed tag, for example `git tag -s v1.0.0 -m "v1.0.0"` and `git push origin v1.0.0`. The release workflow builds and signs the binaries, publishes the GitHub release with the notes from the `# Unreleased` section, then renames that section to the new version in `CHANGELOG.md`. The Terraform registry picks up the new release automatically.
