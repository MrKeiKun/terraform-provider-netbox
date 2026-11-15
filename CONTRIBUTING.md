# Contributing

Thank you for helping. Bug reports, feature requests and pull requests are welcome.

- Report a problem or ask for a feature with an [issue](https://github.com/MrKeiKun/terraform-provider-netbox/issues/new/choose).
- [DEVELOPMENT.md](DEVELOPMENT.md) explains how the provider is built and how to add a resource or a data source.
- Run `make test` before you open a pull request. Run the acceptance tests for the resources you changed with `make testacc-specific-test`.
- Run `make docs` after a change to a schema. The files in `docs/` are generated, so do not edit them by hand. CI fails if they are out of date.
- Keep a pull request to one change, and describe it in plain words.

Everybody who takes part follows the [code of conduct](.github/CODE_OF_CONDUCT.md).
