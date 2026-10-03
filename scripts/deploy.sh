#!/usr/bin/env bash
# deploy.sh — build and deploy ShipCheck to AWS via CDK.
#
# Steps:
#   1. Build the Lambda bootstrap binary (infra/lambda-build/bootstrap).
#   2. Build the SPA static assets (web/dist).
#   3. cdk deploy (requires the AWS CDK CLI — Node.js — and AWS credentials).
#
# All resources sit within AWS always-free allowances at low traffic. The
# deployed public demo defaults to the free deterministic advisor, so no model
# key is required. Run `cdk bootstrap` once per account/region before the first
# deploy.
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"

echo "== 1/3 Build Lambda binary =="
bash "$repo_root/scripts/build-lambda.sh"

echo "== 2/3 Build SPA =="
( cd "$repo_root/web" && npm ci && npm run build )

echo "== 3/3 cdk deploy =="
( cd "$repo_root/infra" && cdk deploy --require-approval never )

echo "Done. See the SpaUrl and ApiUrl stack outputs above."
