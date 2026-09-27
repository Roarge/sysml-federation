# The host configuration (AD-0033). OpenTofu 1.8 is the first release whose
# test framework can stand in for a provider, which `make opentofu-check`
# relies on. The lock file beside this one pins the provider builds.
terraform {
  required_version = ">= 1.8.0"

  required_providers {
    proxmox = {
      source  = "bpg/proxmox"
      version = "~> 0.114"
    }
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 5.26"
    }
  }
}

# Both providers read their credentials from the environment, so nothing
# secret is written here: PROXMOX_VE_ENDPOINT and PROXMOX_VE_API_TOKEN for the
# host, CLOUDFLARE_API_TOKEN for the tunnel and the DNS record.
provider "proxmox" {}

provider "cloudflare" {}
