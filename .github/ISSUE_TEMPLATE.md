# Issue Reporting Guide

Thank you for opening an issue. Please use the issue tracker for bug reports and feature requests. For general Terraform usage questions, see the [Terraform community page](https://www.terraform.io/community).

Before you open an issue, check that it is not already reported and that you are using the latest provider version.

## Versions

- Terraform version (`terraform -v`):
- Provider version:
- NetBox version:

## Affected Resource(s) or Data Source(s)

Please list them, for example:

- netbox_site
- netbox_prefix

## Terraform Configuration

```hcl
# Paste the smallest configuration that shows the problem.
# Remove tokens and other secrets first.
```

## Debug Output

Run with `TF_LOG=DEBUG` and share the output as a link to a GitHub Gist. Please do not paste it into the issue. Remove tokens and other secrets first.

## Panic Output

If Terraform crashed, link a Gist with the `crash.log`.

## Expected Behavior

What should have happened?

## Actual Behavior

What happened instead?

## Steps to Reproduce

1. `terraform apply`

## Additional Context

Anything unusual about your NetBox setup, such as plugins, custom fields or a proxy in front of it.

## References

Related issues or pull requests, for example `#123`.
