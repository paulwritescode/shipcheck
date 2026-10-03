#!/usr/bin/env bash
# build-lambda.sh — compile the ShipCheck backend as a Lambda bootstrap binary.
#
# Produces infra/lambda-build/bootstrap: a static Go binary for the
# provided.al2023 custom runtime on arm64. No Docker required.
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
out_dir="$repo_root/infra/lambda-build"
mkdir -p "$out_dir"

echo "Building Lambda bootstrap (linux/arm64, provided.al2023)…"
cd "$repo_root"
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
  go build -tags lambda.norpc -o "$out_dir/bootstrap" ./cmd/shipcheck

echo "Built $out_dir/bootstrap"
