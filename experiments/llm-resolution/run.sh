#!/usr/bin/env bash
# The language model experiment, run against an Ollama server in your own
# network. From the root of a checkout:
#
#   OLLAMA_URL=http://192.168.1.20:11434 bash experiments/llm-resolution/run.sh -quick
#   OLLAMA_URL=http://192.168.1.20:11434 bash experiments/llm-resolution/run.sh
#
# The first is a quick run of twelve tests, a few minutes long, to show the
# chain works. The second is the full run, about an hour. Each writes one
# results file under experiments/llm-resolution/results/ and prints its path.
# That file is the one to hand back. Any other flag is passed on: -resume FILE
# picks up a stopped run, -model NAME asks another model, -out DIR writes
# elsewhere. Nothing is sent anywhere but OLLAMA_URL.
set -euo pipefail

dir="$(cd "$(dirname "$0")" && pwd)"

log() {
    printf '%s %s\n' "$(date -u +%FT%TZ)" "$*" >&2
}

if [ -z "${OLLAMA_URL:-}" ]; then
    log "set OLLAMA_URL to the server's address, for example OLLAMA_URL=http://192.168.1.20:11434"
    exit 2
fi
if ! command -v go >/dev/null 2>&1; then
    log "Go is not on the PATH; the experiment needs the Go version its go.mod names"
    exit 2
fi
# A quick look before anything is built, so a wrong address fails in seconds.
# The program checks again, and also that the server holds the model.
if command -v curl >/dev/null 2>&1; then
    if ! curl -fsS --max-time 10 "${OLLAMA_URL%/}/api/version" >/dev/null 2>&1; then
        log "nothing answers at ${OLLAMA_URL%/}/api/version: is Ollama running, and listening on 0.0.0.0?"
        exit 2
    fi
fi

bin="$(mktemp -d)"
trap 'rm -rf "$bin"' EXIT
(cd "$dir" && go build -o "$bin/llm-resolution" .)
status=0
(cd "$dir" && "$bin/llm-resolution" -url "$OLLAMA_URL" "$@") || status=$?
exit "$status"
