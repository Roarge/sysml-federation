# The demo on a Proxmox host

This OpenTofu configuration runs the demo on your own Proxmox VE host and
publishes it on your own hostname through a Cloudflare named tunnel. One
switch, `run_on_proxmox`, starts it and stops it. Switched on, the host pulls
the published image and runs it as a container, runs the tunnel's connector
beside it the same way, and the tunnel routes your hostname to the demo there.
Switched off, both containers and both images are gone, and the tunnel routes
the hostname to `http://demo:8080`, where the check session's compose stack
runs its own demo.

You don't need any of this to try the demo. `docker run --rm -p 8080:8080
ghcr.io/roarge/sysml-federation` runs it on one machine, as the
[repository README](../README.md) says. This is for keeping it on a public
hostname from a host you already run. The decision behind it is
[AD-0033](../docs/decisions/AD-0033-the-demo-on-the-owners-proxmox-host.md).

## What you need

- **A Proxmox VE host, version 9.1 or later.** Earlier versions can't run a
  container from an OCI image. Application containers are a
  [technology preview](https://pve.proxmox.com/wiki/Linux_Container) in 9.1.
- **Two kinds of storage on it.** The images are pulled into storage that
  holds container templates, `local` on a default installation. The
  containers' root filesystems go on storage that holds containers, which is
  `local-lvm` or `local-zfs` on a default installation.
- **A network bridge** the containers can join (`vmbr0` by default), with DHCP
  on it or two free addresses. The tunnel's connector needs to reach the
  internet from there.
- **A Proxmox API token** for a user allowed to create containers and to pull
  templates. The provider's documentation shows
  [how to create the user, a role and the token](https://github.com/bpg/terraform-provider-proxmox/blob/v0.114.0/docs/index.md)
  with `pveum`. Its example role is broader than this configuration needs.
- **A Cloudflare account with a domain on it**, and an API token with two
  permissions: **Account › Cloudflare Tunnel › Edit** and **Zone › DNS ›
  Edit**. You'll also need the account ID and the zone ID, which the domain's
  overview page shows ([where to find them](https://developers.cloudflare.com/fundamentals/account/find-account-and-zone-ids/)).
- **OpenTofu 1.8 or later** ([install](https://opentofu.org/docs/intro/install/)).

## Set it up once

**1. Give OpenTofu the credentials.** Both providers read them from the
environment, so nothing secret goes in a file:

```sh
export PROXMOX_VE_ENDPOINT='https://pve.example.lan:8006/'
export PROXMOX_VE_API_TOKEN='terraform@pve!provider=<secret>'
export CLOUDFLARE_API_TOKEN='<token>'
# only if the host still uses its self-signed certificate:
export PROXMOX_VE_INSECURE=true
```

**2. Say where things go.** Create `opentofu/terraform.tfvars`. Git ignores it,
since it holds your addresses:

```hcl
proxmox_node          = "pve"
datastore             = "local-lvm"
cloudflare_account_id = "<account id>"
cloudflare_zone_id    = "<zone id>"
hostname              = "demo.example.org"
```

`variables.tf` lists everything else you can set, each with its default:
static addresses in place of DHCP, another bridge, another release of the demo.

**3. If you already have a named tunnel, adopt it.** Say you set one up by hand
for the check session. Add its ID, its name and the ID of its DNS record, and
OpenTofu takes them over instead of creating new ones:

```hcl
tunnel_id     = "<tunnel uuid>"
tunnel_name   = "<the name it was created with>"
dns_record_id = "<record id>"
```

You'll find the tunnel's ID and name on its page in Cloudflare's Zero Trust
dashboard. A DNS record's ID isn't shown there, but the API gives it, as the
`id` of the record in this answer:

```sh
curl -s "https://api.cloudflare.com/client/v4/zones/<zone id>/dns_records?name.exact=demo.example.org" \
  -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN"
```

Leave all three out to start from nothing. The first apply then creates the
tunnel, its route and the DNS record.

**4. Initialise and look before you apply:**

```sh
cd opentofu
tofu init
tofu plan
```

When adopting, the plan should show three resources to import and little else.
If it shows changes to the route or the record, it is showing what OpenTofu
would change to match this configuration. Read them before you go on. When
starting from nothing, it shows three to add.

**5. Apply:**

```sh
tofu apply
```

The tunnel now exists and routes your hostname to `http://demo:8080`, but
nothing runs it yet. If you want the check session's named tunnel, its token
goes in `checkly/.env` as `TUNNEL_TOKEN`:

```sh
tofu output -raw tunnel_token
```

## Run the demo on the host

```sh
tofu apply -var run_on_proxmox=true
```

The host pulls both images, starts the two containers and moves the route to
the demo's address. Then open the address `tofu output demo_url` prints. The
first pull fetches both images from their registries, so it takes longer than
the ones after it.

To keep the switch on between commands, put `run_on_proxmox = true` in
`terraform.tfvars` instead.

## Stop it

```sh
tofu apply -var run_on_proxmox=false
```

Both containers and both images are removed from the host, and the route goes
back to `http://demo:8080`. The tunnel and its DNS record stay.

## Things to know

- **One demo per tunnel at a time.** While the switch is on, don't start a
  check session with the named tunnel. Its connector would serve the same
  tunnel as the one on the host, and the route points at the host. A check
  session on the quick tunnel is unaffected.
- **Check sessions still use the compose stack.** The session's script starts
  a demo and a tunnel of its own. Its trace steps need the collector beside
  that demo, so it doesn't run against the demo on the host.
- **`tofu destroy` is refused.** The tunnel, its route and the DNS record carry
  `prevent_destroy`, so a destroy can't take your public hostname with it. To
  remove them for good, delete them in the Cloudflare dashboard and then drop
  them from the state with `tofu state rm`.
- **The state file holds the tunnel's token.** `terraform.tfstate` stays on
  your machine, and git ignores it. Keep it out of anything you share.
- **The containers don't start with the host.** They run when you ask and stop
  when you ask.

## Testing the configuration

From the repository root:

```sh
make opentofu-check
```

It checks the formatting, validates the configuration and runs
`tests/on_demand.tftest.hcl`, which plans the configuration against stand-ins
for both providers. It needs no credentials and no host, only the network to
download the two providers the first time. Pass `TOFU=/path/to/tofu` if
`tofu` isn't on your path.
