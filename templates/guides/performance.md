---
page_title: "Performance and connection settings"
subcategory: ""
description: |-
  How the provider keeps the number of requests to NetBox low, and the settings that change it.
---

# Performance and connection settings

A `terraform plan` reads every resource in the state from NetBox, so the time it takes grows with the number of resources. The provider does a few things to keep that time and the load on NetBox low. They are on by default and need no configuration.

## Reads are batched

Terraform reads resources one by one, ten at a time unless you change `-parallelism`. NetBox spends most of the time of a small request on work that is the same for every request, so asking for ten objects with one request is much faster than ten requests for one object each.

The provider reads objects of the same kind that Terraform asks for at the same time with one list request, filtered by ID. A read is never held back to wait for others: the first one is sent at once, and the reads that arrive while it is under way go out together. When an object is not in the answer, the provider reads it on its own before it reports it as gone.

In a test with 605 resources against NetBox 4.7, a plan took 155 requests in place of 606.

Set `batch_reads = false` (or `NETBOX_BATCH_READS=false`) when something in front of NetBox, for example a proxy that only allows some paths, does not let list requests through.

## Other things the provider does

- Connections to NetBox are kept open and used again. This matters most with HTTPS, where every new connection needs a TLS handshake.
- The config context of a device or virtual machine is not requested by the resources that do not show it. NetBox has to put it together from all config contexts for every read.
- After a create or update the provider uses the answer of NetBox and does not read the object again, unless the answer leaves out a field.
- Tags are looked up by name once and then remembered for the rest of the run.

## Settings

| Setting           | Default | When to change it                                                                                         |
| ----------------- | ------- | --------------------------------------------------------------------------------------------------------- |
| `batch_reads`     | `true`  | Set to `false` when list requests are not allowed through to NetBox.                                      |
| `page_size`       | `1000`  | Lower it when NetBox is slow to answer a large page of a data source that lists objects.                  |
| `request_timeout` | `10`    | Raise it when NetBox takes longer than 10 seconds to answer, for example for large data sources.          |
| `max_retries`     | `3`     | Raise it when NetBox or a proxy in front of it often answers 429, 502, 503 or 504. `0` turns retries off. |

When NetBox answers `429 Too Many Requests`, the provider waits and sends the request again. It waits as long as the `Retry-After` header says, up to 30 seconds.
