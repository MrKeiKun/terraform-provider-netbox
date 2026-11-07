terraform {
  required_providers {
    netbox = {
      source  = "mrkeikun/netbox"
      version = "~> 3.2.1"
    }
  }
}

# example provider configuration for https://demo.netbox.dev
provider "netbox" {
  server_url = "https://demo.netbox.dev"
  api_token  = "<your api key>"
}
