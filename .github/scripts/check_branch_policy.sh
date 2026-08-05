#!/usr/bin/env bash
set -euo pipefail

target_branch="${1:?usage: check_branch_policy.sh <master|singbox-backend> [gust-x-dir]}"
gust_x_dir="${2:-}"

fail() {
  echo "branch policy: $*" >&2
  exit 1
}

check_absent() {
  local root="$1"
  shift
  local path
  for path in "$@"; do
    if git -C "${root}" ls-files --error-unmatch -- "${path}" >/dev/null 2>&1; then
      fail "${path} is sing-box-only and must not exist on master"
    fi
  done
}

check_present() {
  local root="$1"
  shift
  local path
  for path in "$@"; do
    if ! git -C "${root}" ls-files --error-unmatch -- "${path}" >/dev/null 2>&1; then
      fail "required sing-box extension file is missing: ${path}"
    fi
  done
}

case "${target_branch}" in
  master)
    check_absent . \
      .github/singbox-gust-x.ref \
      .github/workflows/singbox-compat.yml \
      SINGBOX-NOTICE.md \
      SINGBOX-VALIDATION.md \
      cmd/gost/SINGBOX_MANUAL.md \
      licenses/GPL-3.0.txt
    if grep -Fq 'github.com/sagernet/sing-box' go.mod; then
      fail "the embedded sing-box dependency must not enter master"
    fi
    if [[ -n "${gust_x_dir}" ]]; then
      check_absent "${gust_x_dir}" backend/singbox docs/singbox.md
      if grep -Fq 'github.com/sagernet/sing-box' "${gust_x_dir}/go.mod"; then
        fail "the gust-x embedded sing-box dependency must not enter master"
      fi
    fi
    ;;
  singbox-backend)
    check_present . \
      .github/singbox-gust-x.ref \
      .github/workflows/singbox-compat.yml \
      SINGBOX-NOTICE.md \
      SINGBOX-VALIDATION.md \
      cmd/gost/SINGBOX_MANUAL.md
    git fetch --no-tags --prune origin \
      '+refs/heads/master:refs/remotes/origin/master'
    if ! git merge-base --is-ancestor origin/master HEAD; then
      fail "singbox-backend must contain the current origin/master baseline"
    fi
    if [[ -n "${gust_x_dir}" ]]; then
      check_present "${gust_x_dir}" backend/singbox docs/singbox.md
    fi
    ;;
  *)
    fail "unsupported target branch: ${target_branch}"
    ;;
esac

echo "branch policy passed for ${target_branch}"
