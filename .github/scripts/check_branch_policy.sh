#!/usr/bin/env bash
set -euo pipefail

target_branch="${1:?usage: check_branch_policy.sh <master|singbox-backend|qtui-route-manage> [gust-x-dir]}"
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
      fail "${path} must not exist on ${target_branch}"
    fi
  done
}

check_present() {
  local root="$1"
  shift
  local path
  for path in "$@"; do
    if ! git -C "${root}" ls-files --error-unmatch -- "${path}" >/dev/null 2>&1; then
      fail "required ${target_branch} file is missing: ${path}"
    fi
  done
}

check_no_singbox() {
  local root="$1"
  check_absent "${root}" \
    .github/singbox-gust-x.ref \
    .github/workflows/singbox-compat.yml \
    SINGBOX-NOTICE.md \
    SINGBOX-VALIDATION.md \
    cmd/gost/SINGBOX_MANUAL.md \
    licenses/GPL-3.0.txt
}

case "${target_branch}" in
  master)
    check_no_singbox .
    check_absent . \
      .github/qtui-route-manage-master.ref \
      .github/qtui-route-manage-gust-x.ref \
      .github/workflows/qtui-route-manage-release.yml \
      cmd/gost-route-manager \
      internal/routemanager \
      docs/route-manager.zh-CN.md
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
  qtui-route-manage)
    check_no_singbox .
    check_present . \
      .github/qtui-route-manage-master.ref \
      .github/qtui-route-manage-gust-x.ref \
      .github/workflows/qtui-route-manage-release.yml \
      cmd/gost-route-manager/main.go \
      internal/routemanager/config.go \
      docs/route-manager.zh-CN.md
    baseline="$(tr -d '[:space:]' < .github/qtui-route-manage-master.ref)"
    if [[ ! "${baseline}" =~ ^[0-9a-f]{40}$ ]]; then
      fail "invalid qtui-route-manage master baseline: ${baseline}"
    fi
    git fetch --no-tags --prune origin \
      '+refs/heads/master:refs/remotes/origin/master'
    if ! git merge-base --is-ancestor "${baseline}" origin/master; then
      fail "QtUI recorded baseline is not contained in origin/master"
    fi
    if ! git merge-base --is-ancestor "${baseline}" HEAD; then
      fail "qtui-route-manage must contain its recorded master baseline"
    fi
    if [[ -n "${gust_x_dir}" ]]; then
      check_absent "${gust_x_dir}" backend/singbox docs/singbox.md
      gust_x_pin="$(tr -d '[:space:]' < .github/qtui-route-manage-gust-x.ref)"
      if [[ ! "${gust_x_pin}" =~ ^[0-9a-f]{40}$ ]]; then
        fail "invalid qtui-route-manage gust-x pin: ${gust_x_pin}"
      fi
      if [[ "$(git -C "${gust_x_dir}" rev-parse HEAD)" != "${gust_x_pin}" ]]; then
        fail "gust-x checkout does not match the QtUI pin ${gust_x_pin}"
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
