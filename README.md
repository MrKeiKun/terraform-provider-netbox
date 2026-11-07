# terraform-provider-netbox

The Terraform Netbox provider is a plugin for Terraform that allows for the full lifecycle management of [Netbox](https://netboxlabs.com/docs/netbox/) resources.
This provider is maintained by [MrKeiKun](https://github.com/MrKeiKun).

See: [Official documentation](https://registry.terraform.io/providers/mrkeikun/netbox/latest/docs) in the Terraform registry.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 0.12.x

## Supported netbox versions

Netbox often makes breaking API changes even in non-major releases. Check the table below to see which version a provider was tested against. It is generally recommended to use the provider version matching your Netbox version. We aim to always support the latest minor version of Netbox. The oldest supported release is Netbox 4.4.0, Netbox 4.3 and older are no longer supported.

Each version of the provider has a built-in list of all Netbox versions it supports at release time. Upon initialization, the provider will probe your Netbox version and include a (non-blocking) warning if the used Netbox version is not supported.

| Netbox version  | Provider version |
| --------------- | ---------------- |
| v4.4.0 - 4.7.4  | v1.0.0 and up    |

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```sh
go install
```

## Installation

Starting with Terraform 0.13, you can download the provider via the Terraform registry.

For further information on how to use third party providers, see the [Terraform documentation](https://www.terraform.io/docs/configuration/providers.html)

Releases for all major plattforms are available on the release page.

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

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `make docs`.

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

_Note:_ Acceptance tests create a docker compose stack on port 8001.

```sh
make testacc
```

If you notice a failed test, it might be due to a stale netbox data volume. Before concluding there is a problem,
refresh the docker containers by running `docker-compose down --volumes` in the `docker` directory. Then run the tests again.

If you get `too many open files` errors when running the acceptance test suite locally on Linux, your user limit for open file descriptors might be too low. You can increase that limit with `ulimit -n 2048`.

## Contribution

We focus on virtual machine management and IPAM. If you want to contribute more resources to this provider, feel free to make a PR.
