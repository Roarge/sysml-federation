#!/usr/bin/env bash
# The check session, in two modes.
#
#   bash checkly/scripts/session.sh up
#
# on the host brings the stack up: the demo, one tunnel, the collector, the
# viewer and the runner, with the profiles the environment calls for. Inside
# the runner the script is started with no argument and walks the session's
# twelve steps, each logged under its number so that the log reads as the
# session's design. No credential is ever printed.
set -euo pipefail

dir="$(cd "$(dirname "$0")/.." && pwd)"

log() {
    printf '%s %s\n' "$(date -u +%FT%TZ)" "$*"
}

# ------------------------------------------------------------- on the host --

if [ "${1:-}" = "up" ]; then
    compose="$dir/compose.yml"
    if [ -f "$dir/.env" ]; then
        set -a
        # shellcheck disable=SC1091
        . "$dir/.env"
        set +a
    fi
    # Without an account the runner sees no credentials at all, so that the
    # flag means the same whether or not the file carries them: the stack and
    # the probe, and nothing deployed.
    if [ "${SESSION_WITHOUT_ACCOUNT:-}" = "1" ]; then
        export CHECKLY_API_KEY="" CHECKLY_ACCOUNT_ID=""
    elif [ -z "${CHECKLY_API_KEY:-}" ] || [ -z "${CHECKLY_ACCOUNT_ID:-}" ]; then
        log "CHECKLY_API_KEY and CHECKLY_ACCOUNT_ID are needed for a session, or SESSION_WITHOUT_ACCOUNT=1 for the stack and the probe alone"
        exit 2
    fi
    export ROUTER_CONFIG_PATH=/otel/router.yaml
    # The second collector configuration guards its inbound port with the
    # token and refuses to start without one, whether or not the route to
    # that port exists.
    if [ "${OTEL_COLLECTOR_CONFIG:-}" = collector-checkly.yaml ] && [ -z "${OTEL_INGEST_TOKEN:-}" ]; then
        log "OTEL_COLLECTOR_CONFIG=collector-checkly.yaml needs OTEL_INGEST_TOKEN set to any secret string: it guards the inbound port 4320, which nothing reaches unless the second tunnel route exists"
        exit 2
    fi
    if [ -n "${TUNNEL_TOKEN:-}" ]; then
        if [ -z "${DEMO_HOSTNAME:-}" ]; then
            log "TUNNEL_TOKEN is set and DEMO_HOSTNAME is not: the named tunnel needs the hostname it carries"
            exit 2
        fi
        export TUNNEL_MODE=named
        profiles=(--profile checkly --profile named)
    else
        export TUNNEL_MODE=quick
        profiles=(--profile checkly --profile quick)
    fi
    if [ -z "${CHECKLY_ALERT_EMAIL:-}" ] && [ -z "${CHECKLY_WEBHOOK_URL:-}" ]; then
        log "neither CHECKLY_ALERT_EMAIL nor CHECKLY_WEBHOOK_URL is set, so the project deploys with no alert channel"
    fi
    log "tunnel mode $TUNNEL_MODE"
    exec docker compose -f "$compose" "${profiles[@]}" up
fi

# --------------------------------------------------------- in the runner --

cd "$dir"

mode="${TUNNEL_MODE:-quick}"
with_account=1
if [ -z "${CHECKLY_API_KEY:-}" ] || [ -z "${CHECKLY_ACCOUNT_ID:-}" ]; then
    with_account=0
fi
test_status=0
deploy_status=0
published=0
deployed=0
ping_url=""

# The runner sleeps between steps and while it waits for a stop. A sleep in
# the background under wait is what lets a signal reach the trap at once, and
# the three long calls to the service, the test session, the suite session
# and the trigger, run the same way for the same reason: a stop during one
# of them reaches the trap at once rather than when the call returns, which
# a trigger can put past the grace period.
pause() {
    sleep "$1" &
    wait $! || true
}

graphql() {
    curl -s -m 20 -H 'Content-Type: application/json' -d "$1" "$DEMO_URL/graphql"
}

# Step 12, on a stop. The project is destroyed if it was deployed, the
# account variable removed if it was published, and the exit code is the test
# session's. A deploy that failed is logged with its own status and does not
# replace the test session's.
on_stop() {
    trap - TERM INT
    log "[12/12] destroyOnStop"
    if [ "$deployed" = 1 ]; then
        npx checkly destroy --force || log "destroy failed, the project may still be deployed"
    fi
    if [ "$published" = 1 ]; then
        npx checkly env rm DEMO_URL --force || log "the account variable DEMO_URL was not removed"
    fi
    if [ "$deploy_status" != 0 ]; then
        log "the deploy ended with status $deploy_status"
    fi
    log "session ends with the test session's status $test_status"
    exit "$test_status"
}
trap on_stop TERM INT

# Step 1. The project's dependencies from the lock file, into the named
# volume. The runner image carries no HTTP client and no certificate store,
# so the two are installed beside them for the probes and the pings.
log "[1/12] installDependencies"
export DEBIAN_FRONTEND=noninteractive
apt-get -qq update
apt-get -qq install -y --no-install-recommends curl ca-certificates >/dev/null
npm ci --no-audit --no-fund

# Step 2. The public hostname, given in named mode and read from the quick
# tunnel's metrics endpoint otherwise.
log "[2/12] resolveTheDemoUrl"
if [ "$mode" = named ]; then
    DEMO_URL="https://$DEMO_HOSTNAME"
else
    hostname=""
    SECONDS=0
    while [ "$SECONDS" -lt 120 ]; do
        hostname="$(curl -s -m 5 http://tunnel-quick:2000/quicktunnel | sed -n 's/.*"hostname":"\([^"]*\)".*/\1/p' || true)"
        if [ -n "$hostname" ]; then
            break
        fi
        pause 2
    done
    if [ -z "$hostname" ]; then
        log "the quick tunnel reported no hostname within two minutes"
        exit 1
    fi
    DEMO_URL="https://$hostname"
fi
export DEMO_URL
log "the demo is at $DEMO_URL"

# Step 3. The viewer through the tunnel.
log "[3/12] waitForTheDemo"
code=""
SECONDS=0
while [ "$SECONDS" -lt 120 ]; do
    code="$(curl -s -o /dev/null -w '%{http_code}' -m 10 "$DEMO_URL/viewer/" || true)"
    if [ "$code" = 200 ]; then
        break
    fi
    pause 5
done
if [ "$code" != 200 ]; then
    log "the viewer did not answer 200 through the tunnel within 120 seconds, last status ${code:-none}"
    exit 1
fi
log "the viewer answers 200 through the tunnel"

# Step 4. Whether a subscription's events cross the tunnel: a held
# subscription and a mutation beside it. An event frame within ten seconds
# means the streams arrive, and the checks that need one read the verdict.
# The named tunnel is known to carry them.
log "[4/12] probeSubscriptions"
SSE_STREAMS=1
if [ "$mode" = quick ]; then
    SSE_STREAMS=0
    frames="$(mktemp)"
    curl -sN -m 15 -H 'Accept: text/event-stream' -H 'Content-Type: application/json' \
        -d '{"query":"subscription { modelChanged }"}' "$DEMO_URL/graphql" >"$frames" 2>/dev/null &
    probe=$!
    pause 2
    graphql '{"query":"mutation { setAttribute(partId: \"PIPE-S2\", name: \"throughput\", value: 1201) { id } }"}' >/dev/null || true
    for _ in $(seq 1 10); do
        if grep -q '^event: next' "$frames"; then
            SSE_STREAMS=1
            break
        fi
        pause 1
    done
    kill "$probe" 2>/dev/null || true
    wait "$probe" 2>/dev/null || true
    rm -f "$frames"
    graphql '{"query":"mutation { resetModel { version } }"}' >/dev/null || true
fi
export SSE_STREAMS
if [ "$SSE_STREAMS" = 1 ]; then
    log "subscription events arrive through the tunnel, SSE_STREAMS=1"
else
    log "no subscription event arrived within ten seconds, SSE_STREAMS=0"
fi

if [ "$with_account" = 0 ]; then
    log "no account: the stack, the tunnel and the viewer stay up until stopped, and nothing is deployed"
    trap 'exit 0' TERM INT
    while true; do
        pause 3600
    done
fi

# Step 5. Who the session runs as, and what the plan allows.
log "[5/12] readTheAccountPlan"
npx checkly whoami
npx checkly account plan --output=json

# Step 6. The demo URL as the account variable the deployed checks read.
log "[6/12] publishTheDemoUrl"
npx checkly env add DEMO_URL "$DEMO_URL" || npx checkly env update DEMO_URL "$DEMO_URL"
published=1

# Step 7. Every check once, against this instance, recorded as a session.
session="sysml-federation $(date -u +%FT%TZ)"
log "[7/12] recordATestSession: $session"
npx checkly test --record --reporter list --env DEMO_URL="$DEMO_URL" --env SSE_STREAMS="$SSE_STREAMS" \
    --test-session-name "$session" &
wait $! || test_status=$?
log "the test session ended with status $test_status"

# Step 8. The browser suite as a second recorded session.
log "[8/12] recordASuiteSession"
npx checkly pw-test --record --env DEMO_URL="$DEMO_URL" --env SSE_STREAMS="$SSE_STREAMS" \
    --test-session-name "$session, the suite" -- --project chromium &
wait $! || true

# Step 9. The project deployed for the session, and the heartbeat's ping
# address read from the deploy's output, or from the account when the output
# lacks it.
log "[9/12] deployTheProject"
deploy_out="$(mktemp)"
deploy_status=0
npx checkly deploy --force 2>&1 | tee "$deploy_out" || deploy_status=$?
if [ "$deploy_status" != 0 ]; then
    log "deploy failed with status $deploy_status. A slug already taken on the dashboard or the status page is a configuration error, fixed by CHECKLY_DASHBOARD_SLUG or CHECKLY_STATUS_SLUG"
    rm -f "$deploy_out"
    deployed=1
    on_stop
fi
deployed=1
ping_url="$(sed 's/\x1b\[[0-9;]*m//g' "$deploy_out" | grep -oE 'https://ping\.checklyhq\.com/[A-Za-z0-9/_-]+' | head -n 1 || true)"
rm -f "$deploy_out"
if [ -z "$ping_url" ]; then
    listed="$(npx checkly checks list --type HEARTBEAT --output json 2>/dev/null || true)"
    ping_url="$(printf '%s' "$listed" | grep -oE 'https://ping\.checklyhq\.com/[A-Za-z0-9/_-]+' | head -n 1 || true)"
    if [ -z "$ping_url" ]; then
        for id in $(printf '%s' "$listed" | grep -oE '"id": *"[^"]+"' | sed 's/.*"\([^"]*\)"$/\1/'); do
            ping_url="$(npx checkly checks get "$id" --output json 2>/dev/null | grep -oE 'https://ping\.checklyhq\.com/[A-Za-z0-9/_-]+' | head -n 1 || true)"
            if [ -n "$ping_url" ]; then
                break
            fi
        done
    fi
fi
if [ -z "$ping_url" ]; then
    log "no heartbeat ping address was found in the deploy output or the account, so the pings are skipped"
fi
dashboard_slug="${CHECKLY_DASHBOARD_SLUG:-sysml-federation-${CHECKLY_ACCOUNT_ID:0:8}}"
status_slug="${CHECKLY_STATUS_SLUG:-sysml-federation-${CHECKLY_ACCOUNT_ID:0:8}-status}"
log "dashboard https://$dashboard_slug.checklyhq.com"
log "status page https://$status_slug.checkly-status-page.com"

# Step 10. Every deployed check once, so the dashboard and the page fill.
log "[10/12] triggerOnce"
npx checkly trigger --tags demo --record &
wait $! || log "the trigger ended with a failure, the session goes on"

# Step 11. The heartbeat, every five minutes, until stopped.
log "[11/12] keepTheHeartbeat"
while true; do
    if [ -n "$ping_url" ]; then
        if curl -s -o /dev/null -m 5 --retry 3 "$ping_url"; then
            log "heartbeat sent"
        else
            log "heartbeat failed"
        fi
    fi
    pause 300
done
