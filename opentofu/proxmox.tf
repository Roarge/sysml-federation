# The demo and the tunnel's connector on the owner's Proxmox host, for as long
# as the switch is on. Each runs as an application container made from its
# published image, which Proxmox VE supports from 9.1 as a technology preview.
# Neither image carries a shell or a DHCP client, so with DHCP the host answers
# on the container's behalf.

locals {
  on = var.run_on_proxmox ? 1 : 0

  # The address the tunnel routes the public hostname to while the switch is
  # on: the static one given, or the one the host reports once the container
  # has an address from DHCP.
  demo_host = var.demo_address == "dhcp" ? one([for c in proxmox_virtual_environment_container.demo : c.ipv4["eth0"]]) : split("/", var.demo_address)[0]
}

resource "proxmox_oci_image" "demo" {
  count        = local.on
  node_name    = var.proxmox_node
  datastore_id = var.template_datastore
  reference    = "ghcr.io/roarge/sysml-federation:${var.demo_version}"
}

resource "proxmox_oci_image" "cloudflared" {
  count        = local.on
  node_name    = var.proxmox_node
  datastore_id = var.template_datastore
  reference    = "docker.io/cloudflare/cloudflared:${var.cloudflared_version}"
}

resource "proxmox_virtual_environment_container" "demo" {
  count         = local.on
  node_name     = var.proxmox_node
  description   = "The SysML federation demo, from ghcr.io/roarge/sysml-federation:${var.demo_version}, on port 8080."
  tags          = ["sysml-federation"]
  unprivileged  = true
  start_on_boot = false

  operating_system {
    template_file_id = proxmox_oci_image.demo[0].id
  }

  disk {
    datastore_id = var.datastore
    size         = 2
  }

  memory {
    dedicated = 512
  }

  network_interface {
    name         = "eth0"
    bridge       = var.bridge
    host_managed = var.demo_address == "dhcp"
  }

  initialization {
    hostname = "sysml-federation"

    ip_config {
      ipv4 {
        address = var.demo_address
        gateway = var.demo_address == "dhcp" ? null : var.gateway
      }
    }
  }

  wait_for_ip {
    ipv4 = var.demo_address == "dhcp"
  }

  lifecycle {
    precondition {
      condition     = var.proxmox_node != null && var.datastore != null
      error_message = "Switching the demo on needs proxmox_node and datastore."
    }

    precondition {
      condition     = var.demo_address == "dhcp" || var.gateway != null
      error_message = "A static demo_address needs a gateway."
    }
  }
}

resource "proxmox_virtual_environment_container" "tunnel" {
  count         = local.on
  node_name     = var.proxmox_node
  description   = "The named tunnel's connector for the SysML federation demo, from cloudflare/cloudflared:${var.cloudflared_version}."
  tags          = ["sysml-federation"]
  unprivileged  = true
  start_on_boot = false

  operating_system {
    template_file_id = proxmox_oci_image.cloudflared[0].id
  }

  disk {
    datastore_id = var.datastore
    size         = 1
  }

  memory {
    dedicated = 256
  }

  network_interface {
    name         = "eth0"
    bridge       = var.bridge
    host_managed = var.tunnel_address == "dhcp"
  }

  # The image's own entrypoint is cloudflared --no-autoupdate, and its command
  # prints the version. The compose file's named profile runs the tunnel the
  # same way, with the token in the environment.
  environment_variables = {
    TUNNEL_TOKEN = data.cloudflare_zero_trust_tunnel_cloudflared_token.demo.token
  }

  initialization {
    hostname   = "cloudflared"
    entrypoint = "cloudflared --no-autoupdate tunnel run"

    ip_config {
      ipv4 {
        address = var.tunnel_address
        gateway = var.tunnel_address == "dhcp" ? null : var.gateway
      }
    }
  }

  lifecycle {
    precondition {
      condition     = var.tunnel_address == "dhcp" || var.gateway != null
      error_message = "A static tunnel_address needs a gateway."
    }
  }
}
