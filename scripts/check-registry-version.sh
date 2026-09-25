#!/usr/bin/env bash
# Fails if a consumer of the formtransform pin disagrees with .registry-version.
# Run in CI; bump every consumer together (see CONVERSION_API.md).
set -euo pipefail
cd "$(dirname "$0")/.."

tag="$(tr -d '[:space:]' < .registry-version)"
ver="${tag#v}"
tarball="https://github.com/CorrelAid/formtransform/releases/download/${tag}/correlaid-formtransform-${ver}.tgz"
status=0

check() {
	if ! grep -qF "$2" "$1"; then
		echo "::error file=$1::expected $2 (from .registry-version = $tag)"
		status=1
	fi
}

check docker-compose.yml "ghcr.io/correlaid/schematron-worker:${tag}"
check ddi-emitter/package.json "$tarball"
check ddi-emitter/package-lock.json "$tarball"

if [ "$status" -eq 0 ]; then
	echo "formtransform pin consistent: $tag"
fi
exit "$status"
