# 1.0.0 (unreleased)

NOTES

* The provider supports Netbox 4.4.0 to 4.7.4. Older Netbox versions are not supported.
* The provider now talks to Netbox with its own small REST client. The generated `go-netbox` client is no longer used, so new Netbox releases do not need a client update.
