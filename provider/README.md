# Public provider package

This package makes the NetBox provider importable as a Go library.

## Why this folder exists

The provider is implemented in `internal/fwprovider`, and Go does not let other modules import `internal` packages.
Tools that want to serve the provider in process, instead of starting it as a plugin through the Terraform CLI, need this.

## Usage

```go
import (
    "github.com/hashicorp/terraform-plugin-framework/providerserver"

    "github.com/MrKeiKun/terraform-provider-netbox/provider"
)

// version is reported as the provider version.
server, err := providerserver.NewProtocol6WithError(provider.New(version)())()
```
