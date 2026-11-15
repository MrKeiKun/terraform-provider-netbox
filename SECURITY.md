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
