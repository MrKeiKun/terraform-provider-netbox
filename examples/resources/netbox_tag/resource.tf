resource "netbox_tag" "dmz" {
  name      = "DMZ"
  color_hex = "ff00ff"
}

# Example with object types limitation
resource "netbox_tag" "device_only" {
  name         = "Device Only"
  slug         = "device-only"
  color_hex    = "00ff00"
  description  = "Tag that can only be applied to devices"
  object_types = ["dcim.device"]
}

# Example with multiple object types
resource "netbox_tag" "network_equipment" {
  name         = "Network Equipment"
  slug         = "network-equipment"
  color_hex    = "0000ff"
  description  = "Tag for network-related equipment"
  object_types = ["dcim.device", "virtualization.virtualmachine"]
}
