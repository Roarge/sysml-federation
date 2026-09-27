output "demo_url" {
  description = "Where the demo answers through the named tunnel."
  value       = "https://${var.hostname}"
}

output "origin" {
  description = "Where the tunnel hands each request: the demo on the host while the switch is on, the compose session's demo otherwise."
  value       = local.demo_origin
}

output "tunnel_id" {
  description = "The named tunnel's UUID."
  value       = cloudflare_zero_trust_tunnel_cloudflared.demo.id
}

output "tunnel_token" {
  description = "The connector's credential, the TUNNEL_TOKEN a compose session's named profile needs. Read it with tofu output -raw tunnel_token."
  value       = data.cloudflare_zero_trust_tunnel_cloudflared_token.demo.token
  sensitive   = true
}
