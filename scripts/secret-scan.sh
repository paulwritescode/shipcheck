#!/usr/bin/env bash
# secret-scan.sh — dependency-free credential scanner for ShipCheck.
#
# Scans tracked/staged files for common credential patterns so no secret is
# ever committed to this public repository. Exit non-zero if a match is found.
#
# Usage:
#   scripts/secret-scan.sh            # scan files staged for commit
#   scripts/secret-scan.sh --all      # scan all tracked files
set -euo pipefail

mode="${1:-staged}"

# Portable file collection (works on bash 3.2, the macOS default — no mapfile).
files=()
while IFS= read -r line; do
  [[ -n "$line" ]] && files+=("$line")
done < <(
  if [[ "$mode" == "--all" ]]; then
    git ls-files
  else
    git diff --cached --name-only --diff-filter=ACM
  fi
)

if [[ ${#files[@]} -eq 0 ]]; then
  echo "secret-scan: no files to scan."
  exit 0
fi

# Patterns for high-signal credential formats. Kept deliberately specific to
# avoid false positives on ordinary prose/code.
patterns=(
  'AKIA[0-9A-Z]{16}'                                   # AWS access key id
  'aws_secret_access_key[[:space:]]*=[[:space:]]*[A-Za-z0-9/+=]{40}'
  'hf_[A-Za-z0-9]{30,}'                                # Hugging Face token
  'nvapi-[A-Za-z0-9_-]{20,}'                           # NVIDIA API key
  '-----BEGIN[[:space:]].*PRIVATE KEY-----'            # private keys
  'xox[baprs]-[0-9A-Za-z-]{10,}'                       # Slack tokens
  'ghp_[A-Za-z0-9]{36}'                                # GitHub PAT
)

found=0
for f in "${files[@]}"; do
  # Skip binary files and this scanner itself (it contains the patterns).
  [[ -f "$f" ]] || continue
  [[ "$f" == "scripts/secret-scan.sh" ]] && continue
  if grep -Iq . "$f" 2>/dev/null; then
    for p in "${patterns[@]}"; do
      if grep -EnH "$p" "$f" 2>/dev/null; then
        found=1
      fi
    done
  fi
done

if [[ "$found" -ne 0 ]]; then
  echo ""
  echo "secret-scan: potential credential(s) detected (see matches above)."
  echo "Remove the secret and store it in Lambda env / AWS Secrets Manager instead."
  exit 1
fi

echo "secret-scan: no credentials detected."
exit 0
