# Security policy

## Supported versions

Security fixes go into the latest release of the provider. Update to the latest release before you report a problem.

## Reporting a vulnerability

Please do not open a public issue for a security problem.

Report it privately with [GitHub private vulnerability reporting](https://github.com/MrKeiKun/terraform-provider-netbox/security/advisories/new). Say which provider version and NetBox version you use, and how to reproduce the problem.

When the report is confirmed, the fix is released first and the problem is made public afterwards, with credit to you if you want it.

## What counts

The provider handles the NetBox API token and sends it to the server you configure. Reports about these are most useful:

- the token or another secret ending up in logs, state or error messages
- the token being sent to a host other than the configured server
- TLS settings that are not applied as documented

## What the provider does to protect the token

- `api_token` and `headers` are marked sensitive, as are the attributes that hold a secret of a NetBox object, for example `preshared_key`, `auth_psk`, `auth_key`, the `secret` and `additional_headers` of a webhook and the `parameters` of a data source. Terraform does not print them in a plan. They are still stored in the state, so protect the state.
- A redirect is only followed when it stays on the host in `server_url` and does not go from https to http. The token and the custom headers are never sent anywhere else. A redirect that would turn a request that writes into a GET is refused as well, and the provider warns at startup when `server_url` itself redirects.
- The link to the next page of a list is not followed as it is. Only the page position is taken from it, the host always comes from `server_url`.
- TLS 1.2 is the lowest version the provider accepts. `allow_insecure_https` turns off certificate checks and should only be used for testing. Use `ca_cert_file` for a private certificate authority.
- A `server_url` with `http://` sends the token without encryption. The provider writes a warning to the log (`TF_LOG=WARN`) unless the server is on the same machine.
- Error messages hold the request URL and the answer of NetBox, never the token or the headers.

## Checking a release

The checksum file of every release is signed with GPG, which Terraform checks when it installs the provider. Each release file also has a build provenance attestation, which tells that the file was built by the release workflow of this repository from the tagged commit. Check a downloaded file with the GitHub CLI:

```sh
gh attestation verify terraform-provider-netbox_<version>_linux_amd64.zip --repo MrKeiKun/terraform-provider-netbox
```

The workflows in this repository pin every action to a commit, run `govulncheck` every day and on every pull request, and run `gosec` as part of the linter.
