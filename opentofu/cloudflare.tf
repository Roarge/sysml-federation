# The named tunnel, its routes and the public hostname's DNS record. When they
# already exist, imports.tf adopts them, and otherwise the first apply creates
# them. They outlive every session, so a destroy is refused.

locals {
  # Where the public hostname goes: the demo on the host while the switch is
  # on, and otherwise the demo service in the compose session's network.
  demo_origin = var.run_on_proxmox ? "http://${local.demo_host}:8080" : "http://demo:8080"
}

resource "cloudflare_zero_trust_tunnel_cloudflared" "demo" {
  account_id = var.cloudflare_account_id
  name       = var.tunnel_name
  config_src = "cloudflare"

  lifecycle {
    prevent_destroy = true
  }
}

resource "cloudflare_zero_trust_tunnel_cloudflared_config" "demo" {
  account_id = var.cloudflare_account_id
  tunnel_id  = cloudflare_zero_trust_tunnel_cloudflared.demo.id

  # The last rule is the catch-all the tunnel provider requires. The optional
  # collector route stays with the compose session whichever way the switch is
  # set, since no collector runs on the host.
  config = {
    ingress = concat(
      [{ hostname = var.hostname, service = local.demo_origin }],
      var.collector_hostname == null ? [] : [{ hostname = var.collector_hostname, service = "http://otel-collector:4320" }],
      [{ hostname = null, service = "http_status:404" }],
    )
  }

  lifecycle {
    prevent_destroy = true
  }
}

resource "cloudflare_dns_record" "demo" {
  zone_id = var.cloudflare_zone_id
  name    = var.hostname
  type    = "CNAME"
  content = "${cloudflare_zero_trust_tunnel_cloudflared.demo.id}.cfargotunnel.com"
  proxied = true
  ttl     = 1

  lifecycle {
    prevent_destroy = true
  }
}

# The connector's credential. The container on the host runs the tunnel with
# it, and the compose session's named profile takes the same value as
# TUNNEL_TOKEN in checkly/.env. It lands in the state file, which stays on the
# owner's machine.
data "cloudflare_zero_trust_tunnel_cloudflared_token" "demo" {
  account_id = var.cloudflare_account_id
  tunnel_id  = cloudflare_zero_trust_tunnel_cloudflared.demo.id
}
