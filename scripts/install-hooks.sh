#!/usr/bin/env bash
# install-hooks.sh — install the ShipCheck git pre-commit hook.
#
# The hook runs the secret scanner against staged files so no credential is
# ever committed. Run once after cloning:
#   bash scripts/install-hooks.sh
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
hook_path="$repo_root/.git/hooks/pre-commit"

cat > "$hook_path" <<'HOOK'
#!/usr/bin/env bash
# ShipCheck pre-commit hook: block commits that contain credentials.
set -euo pipefail
repo_root="$(git rev-parse --show-toplevel)"
bash "$repo_root/scripts/secret-scan.sh" staged
HOOK

chmod +x "$hook_path"
echo "Installed pre-commit secret-scan hook at $hook_path"
