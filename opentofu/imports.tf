# Adopts a tunnel and a DNS record that were created by hand, when their IDs
# are given, so the first apply takes them over and creates nothing new at the
# tunnel provider. Once they are in the state, these blocks do nothing. Without
# the IDs, the first apply creates both.

import {
  for_each = var.tunnel_id == null ? toset([]) : toset([var.tunnel_id])
  to       = cloudflare_zero_trust_tunnel_cloudflared.demo
  id       = "${var.cloudflare_account_id}/${each.value}"
}

import {
  for_each = var.tunnel_id == null ? toset([]) : toset([var.tunnel_id])
  to       = cloudflare_zero_trust_tunnel_cloudflared_config.demo
  id       = "${var.cloudflare_account_id}/${each.value}"
}

import {
  for_each = var.dns_record_id == null ? toset([]) : toset([var.dns_record_id])
  to       = cloudflare_dns_record.demo
  id       = "${var.cloudflare_zone_id}/${each.value}"
}
