# The host configuration planned against stand-ins for both providers, so the
# runs need no credentials, no Proxmox host and no tunnel. Each run is one
# position of the switch (SR-49). What the stand-ins cannot show, that the host
# really runs the image and the edge really routes to it, waits for a run on the
# owner's host.

mock_provider "proxmox" {
  mock_resource "proxmox_virtual_environment_container" {
    defaults = {
      ipv4 = { eth0 = "192.168.1.77" }
    }
  }
}

mock_provider "cloudflare" {
  mock_data "cloudflare_zero_trust_tunnel_cloudflared_token" {
    defaults = {
      token = "stand-in-token"
    }
  }
}

variables {
  proxmox_node          = "pve"
  datastore             = "local-zfs"
  cloudflare_account_id = "0123456789abcdef0123456789abcdef"
  cloudflare_zone_id    = "fedcba9876543210fedcba9876543210"
  tunnel_id             = "4f1c2d3e-5a6b-4c7d-8e9f-0a1b2c3d4e5f"
  dns_record_id         = "abcdef0123456789abcdef0123456789"
}

run "switched_off_is_the_compose_session" {
  command = plan

  assert {
    condition     = length(proxmox_oci_image.demo) + length(proxmox_oci_image.cloudflared) + length(proxmox_virtual_environment_container.demo) + length(proxmox_virtual_environment_container.tunnel) == 0
    error_message = "Switched off, nothing may be created on the Proxmox host."
  }

  assert {
    condition     = cloudflare_zero_trust_tunnel_cloudflared_config.demo.config.ingress[0].hostname == "demo.sysml-federation.org" && cloudflare_zero_trust_tunnel_cloudflared_config.demo.config.ingress[0].service == "http://demo:8080"
    error_message = "Switched off, the public hostname must route to the compose session's demo."
  }

  assert {
    condition     = one([for rule in cloudflare_zero_trust_tunnel_cloudflared_config.demo.config.ingress : rule.service if rule.hostname == null]) == "http_status:404" && cloudflare_zero_trust_tunnel_cloudflared_config.demo.config.ingress[length(cloudflare_zero_trust_tunnel_cloudflared_config.demo.config.ingress) - 1].hostname == null
    error_message = "The last route must be the catch-all the tunnel provider requires, and the only one without a hostname."
  }

  assert {
    condition     = cloudflare_dns_record.demo.type == "CNAME" && cloudflare_dns_record.demo.content == "4f1c2d3e-5a6b-4c7d-8e9f-0a1b2c3d4e5f.cfargotunnel.com" && cloudflare_dns_record.demo.proxied
    error_message = "The public hostname must be a proxied CNAME to the named tunnel."
  }
}

run "switched_on_runs_both_on_the_host" {
  command = plan

  variables {
    run_on_proxmox = true
    demo_address   = "192.168.1.50/24"
    tunnel_address = "192.168.1.51/24"
    gateway        = "192.168.1.1"
  }

  assert {
    condition     = length(proxmox_oci_image.demo) + length(proxmox_oci_image.cloudflared) + length(proxmox_virtual_environment_container.demo) + length(proxmox_virtual_environment_container.tunnel) == 4
    error_message = "Switched on, the host must pull two images and run two containers."
  }

  assert {
    condition     = proxmox_oci_image.demo[0].reference == "ghcr.io/roarge/sysml-federation:0.3.0"
    error_message = "The demo must run from the published image at the pinned release."
  }

  assert {
    condition     = strcontains(file("${path.module}/../checkly/compose.yml"), "$${DEMO_IMAGE:-ghcr.io/roarge/sysml-federation}")
    error_message = "The published image must be the one the compose file runs."
  }

  assert {
    condition     = proxmox_oci_image.cloudflared[0].reference == "docker.io/cloudflare/cloudflared:2026.9.1" && strcontains(file("${path.module}/../checkly/compose.yml"), "image: cloudflare/cloudflared:2026.9.1@")
    error_message = "The host must run the tunnel connector at the release the compose file pins."
  }

  assert {
    condition     = alltrue([for c in concat(proxmox_virtual_environment_container.demo, proxmox_virtual_environment_container.tunnel) : c.unprivileged && !c.start_on_boot])
    error_message = "Both containers must be unprivileged and must not start with the host."
  }

  assert {
    condition     = proxmox_virtual_environment_container.tunnel[0].initialization[0].entrypoint == "cloudflared --no-autoupdate tunnel run"
    error_message = "The connector must run the named tunnel, as the compose file's named profile does."
  }

  assert {
    condition     = nonsensitive(proxmox_virtual_environment_container.tunnel[0].environment_variables["TUNNEL_TOKEN"] == "stand-in-token")
    error_message = "The connector must be given the named tunnel's token."
  }

  assert {
    condition     = cloudflare_zero_trust_tunnel_cloudflared_config.demo.config.ingress[0].service == "http://192.168.1.50:8080"
    error_message = "Switched on, the public hostname must route to the demo on the host."
  }
}

run "switched_on_with_dhcp_routes_to_the_address_the_host_reports" {
  command = apply

  variables {
    run_on_proxmox = true
  }

  assert {
    condition     = cloudflare_zero_trust_tunnel_cloudflared_config.demo.config.ingress[0].service == "http://192.168.1.77:8080"
    error_message = "With DHCP, the route must follow the address the host reports for the demo."
  }

  assert {
    condition     = alltrue([for c in concat(proxmox_virtual_environment_container.demo, proxmox_virtual_environment_container.tunnel) : c.network_interface[0].host_managed])
    error_message = "With DHCP, the host must answer on the containers' behalf, since neither image carries a DHCP client."
  }
}

run "the_collector_route_stays_with_the_compose_session" {
  command = plan

  variables {
    run_on_proxmox     = true
    demo_address       = "192.168.1.50/24"
    gateway            = "192.168.1.1"
    collector_hostname = "otel.sysml-federation.org"
  }

  assert {
    condition     = cloudflare_zero_trust_tunnel_cloudflared_config.demo.config.ingress[1].hostname == "otel.sysml-federation.org" && cloudflare_zero_trust_tunnel_cloudflared_config.demo.config.ingress[1].service == "http://otel-collector:4320"
    error_message = "The optional collector route must point at the compose session's collector whichever way the switch is set."
  }
}
