#!/usr/bin/env bash
# Validate DDI codebooks with the pinned schematron worker. Starts a private
# NATS + worker pair (containers surveyimport-*) on port 4223 if not running.
#   tools/survey-import/validate.sh out/q01-cvts7/codebook.xml ...
set -euo pipefail
root="$(git -C "$(dirname "$0")" rev-parse --show-toplevel)"
port=4223 token=surveyimport
if ! docker ps --format '{{.Names}}' | grep -qx surveyimport-worker; then
  docker rm -f surveyimport-nats surveyimport-worker >/dev/null 2>&1 || true
  docker run -d --name surveyimport-nats -p "$port:4222" nats:2 --auth "$token" >/dev/null
  docker run -d --name surveyimport-worker --network host \
    -e NATS_URL="nats://127.0.0.1:$port" -e NATS_TOKEN="$token" \
    "ghcr.io/correlaid/schematron-worker:$(cat "$root/.registry-version")" >/dev/null
  until docker logs surveyimport-worker 2>&1 | grep -q "Subscribed to"; do
    docker ps --format '{{.Names}}' | grep -qx surveyimport-worker || { docker logs surveyimport-worker; exit 2; }
    read -rt 1 <> <(:) || true
  done
fi
args=()
for f in "$@"; do args+=("$(realpath "$f")"); done
cd "$root"
NATS_URL="nats://127.0.0.1:$port" NATS_TOKEN="$token" go run ./tools/survey-import/validate "${args[@]}"
