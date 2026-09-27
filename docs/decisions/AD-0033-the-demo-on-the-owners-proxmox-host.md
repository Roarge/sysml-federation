# AD-0033 The demo on the owner's Proxmox host, declared in OpenTofu

Status: accepted. Date: 2026-09-27.

## Context

A check session needs the demo reachable from the monitoring service's runners.
AD-0031 met that with a tunnel beside the demo, started by the compose profile
on whatever machine the operator has to hand. The named tunnel it uses was
created by hand in the tunnel provider's dashboard, with its route to
`http://demo:8080` and the DNS record for `demo.sysml-federation.org`. The
session's README said as much, that nothing of them was in the repository.
Nothing recorded which hostname pointed where, and nothing would
have noticed the route changing.

The owner runs a Proxmox VE host on their own network. From version 9.1 the
host can pull an image from an OCI registry and run it as an application
container, a technology preview in that release. The published image suits
that kind of container, with one process on one port and no shell. The demo
could run there, public through the same named tunnel, on a machine that stays
up when a laptop does not.

The articles on automating traceability use OpenTofu as their example of
infrastructure as code. Until this record, every OpenTofu resource they showed
belonged to a hypothetical pipeline, because the demo had none.

## Decision

We will add an OpenTofu configuration under `opentofu/` with one switch,
`run_on_proxmox`.

With the switch on, the owner's Proxmox host pulls two images, the published
one at a pinned release and the tunnel connector's at the version the compose
file pins. It runs each as an unprivileged application container. The tunnel's
route points at the demo container's address. With the switch off, both
containers and both images are removed and the route points at
`http://demo:8080` again, the compose session's demo. A session on the host
starts with one `tofu apply` and ends with another, and nothing runs on the
host between sessions.

When their IDs are given, import blocks adopt the named tunnel, its
configuration and the DNS record, so the first apply takes over what exists
and creates nothing new at the tunnel provider. Without the IDs, the first
apply creates all three, which is how anyone else sets the demo up on their
own account. All three carry `prevent_destroy`, so `tofu destroy` is refused
before it can take the public hostname with it.

The configuration's names follow the two providers' conventions, and none
carries a key from the demo's systems model. No test holds the configuration
to the systems model the way the agreement test holds the compose services to the
session composite. Recovering those links is what the articles set out to do,
and a key in every resource would leave nothing to recover.

The configuration is tested with OpenTofu's own test framework against mocked
providers, by `make opentofu-check`, which needs no credentials. Like the
experiment's tests (AD-0032), it runs locally and not in continuous
integration, because SC-07 names what the workflows run.

## Alternatives considered

A virtual machine on the Proxmox host, running Docker and the compose file and
installed by cloud-init. It is the mature route, and it would run the whole
session, runner included. It loses because it needs a cloud image, a snippets
datastore and SSH access for the provider, which is a lot of machinery for one
container the host can run from its image directly.

The Docker provider, running the image on any Docker host. It loses because it
restates `docker run` and the compose file's `demo` service in a second
language, and a second definition of the same container can drift from the
first.

Managing only the host, and leaving the tunnel and its DNS record as they were.
It loses because the route was the one piece of the session's infrastructure
recorded nowhere, and because the switch has to move the route anyway.

A demo that stays up. AD-0031 rejected a hosted demo for its account and its
bill, and a host that runs the demo only during a session keeps that position.

## Consequences

SC-01 is amended. It places the host configuration outside the product, as it
places the check project, and nothing under `opentofu/` is built, run or read
to run the demo.

While the switch is on, the named tunnel routes to the host, so a compose
session started with the `named` profile reaches no demo. The two are
alternatives for one session at a time. A session on the `quick` profile is
unaffected.

The state file holds the tunnel's token, which the connector's container needs.
It stays on the owner's machine, and the allowlist keeps it out of the
repository as it keeps out every file it does not name. `.terraform.lock.hcl`
is tracked, so a checkout resolves the same provider builds.

SR-49 stays in progress until a run on the owner's host is recorded.
Application containers are a technology preview in Proxmox VE 9.1, and the
mocked plan shows only what the configuration asks for.

The DNS record gets no element of its own in the demo's systems model. The
context's tunnel edge stands for whatever terminates the public hostname, and
the record is part of that.

The check session's script still starts a demo and a tunnel of its own, and its
trace steps expect the router's spans in the collector beside it, so a check
session runs against the compose stack and not against the demo on the host.
Pointing a session at the host is a change to the session, left for when it is
wanted.

`make opentofu-check` downloads the two providers the first time it runs, so it
needs a network the unit tests do not.

## Requirements affected

SR-49, SC-01, SC-04

## Sources

[The host configuration's README](https://github.com/Roarge/sysml-federation/blob/main/opentofu/README.md) for how to adopt the tunnel and run a session, with the check session record (AD-0031) and the experiment record (AD-0032) beside it.

For the host, the Proxmox VE documentation on [containers from OCI images](https://pve.proxmox.com/wiki/Linux_Container), and the provider's resources for [pulling an OCI image](https://github.com/bpg/terraform-provider-proxmox/blob/v0.114.0/docs/resources/oci_image.md) and [running a container](https://github.com/bpg/terraform-provider-proxmox/blob/v0.114.0/docs/resources/virtual_environment_container.md).

For the tunnel, its provider's resources for [a tunnel](https://registry.terraform.io/providers/cloudflare/cloudflare/latest/docs/resources/zero_trust_tunnel_cloudflared), [a tunnel's configuration](https://registry.terraform.io/providers/cloudflare/cloudflare/latest/docs/resources/zero_trust_tunnel_cloudflared_config) and [a DNS record](https://registry.terraform.io/providers/cloudflare/cloudflare/latest/docs/resources/dns_record). OpenTofu's [test framework](https://opentofu.org/docs/cli/commands/test/) runs the mocked plan.
