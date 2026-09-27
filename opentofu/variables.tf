# ------------------------------------------------------------- the switch --

variable "run_on_proxmox" {
  description = "Run the demo and the tunnel's connector on the Proxmox host. Off, the tunnel routes to the compose session's demo, as it did before this configuration existed."
  type        = bool
  default     = false
}

# ---------------------------------------------------------------- the host --

variable "proxmox_node" {
  description = "The name of the Proxmox VE node that runs the two containers. Version 9.1 or later, for application containers from OCI images."
  type        = string
  default     = null
}

variable "template_datastore" {
  description = "The storage the two images are pulled into. It must allow container templates, as the default local storage does."
  type        = string
  default     = "local"
}

variable "datastore" {
  description = "The storage for the two containers' root filesystems. It must allow containers, as local-lvm or local-zfs do on a default installation."
  type        = string
  default     = null
}

variable "bridge" {
  description = "The network bridge the two containers join."
  type        = string
  default     = "vmbr0"
}

variable "demo_address" {
  description = "The demo container's IPv4 address in CIDR notation, or \"dhcp\". With DHCP the host answers on the container's behalf, since the image carries no DHCP client."
  type        = string
  default     = "dhcp"

  validation {
    condition     = var.demo_address == "dhcp" || can(cidrhost(var.demo_address, 0))
    error_message = "demo_address must be \"dhcp\" or an IPv4 address in CIDR notation, such as 192.168.1.50/24."
  }
}

variable "tunnel_address" {
  description = "The tunnel connector's IPv4 address in CIDR notation, or \"dhcp\"."
  type        = string
  default     = "dhcp"

  validation {
    condition     = var.tunnel_address == "dhcp" || can(cidrhost(var.tunnel_address, 0))
    error_message = "tunnel_address must be \"dhcp\" or an IPv4 address in CIDR notation."
  }
}

variable "gateway" {
  description = "The IPv4 gateway for a static address. Leave it unset with DHCP."
  type        = string
  default     = null
}

# -------------------------------------------------------------- the images --

variable "demo_version" {
  description = "The release of the published image to run, a tag of ghcr.io/roarge/sysml-federation without its v."
  type        = string
  default     = "0.3.0"

  validation {
    condition     = can(regex("^[0-9]+\\.[0-9]+\\.[0-9]+(-[0-9A-Za-z.]+)?$", var.demo_version))
    error_message = "demo_version must be a release version such as 0.3.0."
  }
}

variable "cloudflared_version" {
  description = "The tunnel connector's release, the tag checkly/compose.yml pins, so that the host and the compose session run the same connector."
  type        = string
  default     = "2026.9.1"
}

# ------------------------------------------------------------- the tunnel --

variable "cloudflare_account_id" {
  description = "The account that owns the named tunnel."
  type        = string
}

variable "cloudflare_zone_id" {
  description = "The zone that holds the public hostname's DNS record."
  type        = string
}

variable "tunnel_id" {
  description = "The UUID of a named tunnel that already exists, as the tunnel provider's dashboard shows it, to adopt it and its routes. Leave it unset to create a new tunnel."
  type        = string
  default     = null
}

variable "tunnel_name" {
  description = "The named tunnel's name. When adopting a tunnel, give the name it was created with."
  type        = string
  default     = "sysml-federation"
}

variable "dns_record_id" {
  description = "The ID of the public hostname's DNS record, if one already exists, to adopt it. Leave it unset to create the record."
  type        = string
  default     = null
}

variable "hostname" {
  description = "The public hostname the tunnel serves, the DEMO_HOSTNAME of checkly/.env."
  type        = string
  default     = "demo.sysml-federation.org"
}

variable "collector_hostname" {
  description = "The optional second public hostname, routed to the trace collector's inbound receiver in the compose network. Unset, there is no such route."
  type        = string
  default     = null
}
