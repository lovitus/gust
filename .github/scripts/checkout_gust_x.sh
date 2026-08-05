#!/usr/bin/env bash
set -euo pipefail

destination="${1:-../gust-x}"
requested_ref="${2:-}"
candidate_ref="${3:-}"
ref_type="${4:-}"

if [[ -n "${GH_TOKEN:-}" ]]; then
  if ! command -v gh >/dev/null 2>&1; then
    echo "GH_TOKEN is set but GitHub CLI is unavailable" >&2
    exit 1
  fi
  gh auth setup-git
fi

pinned_ref="$(tr -d '[:space:]' < .github/singbox-gust-x.ref)"
if [[ -z "${pinned_ref}" ]]; then
  echo "empty .github/singbox-gust-x.ref" >&2
  exit 1
fi

ref="${requested_ref}"
if [[ -z "${ref}" && "${ref_type}" == branch && -n "${candidate_ref}" ]]; then
  if git ls-remote --exit-code --heads https://github.com/lovitus/gust-x.git "${candidate_ref}" >/dev/null 2>&1; then
    ref="${candidate_ref}"
  fi
fi
if [[ -z "${ref}" ]]; then
  ref="${pinned_ref}"
fi

if [[ -e "${destination}" ]]; then
  echo "checkout destination already exists: ${destination}" >&2
  exit 1
fi
git clone --filter=blob:none --no-checkout https://github.com/lovitus/gust-x.git "${destination}"
git -C "${destination}" fetch --depth 1 origin "${ref}"
git -C "${destination}" checkout --detach FETCH_HEAD

resolved="$(git -C "${destination}" rev-parse HEAD)"
echo "gust-x ref: ${ref} (${resolved})"
