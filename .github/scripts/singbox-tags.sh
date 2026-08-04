#!/usr/bin/env bash

set -euo pipefail

# Gust deliberately excludes sing-box's badlinkname and
# tfogo_checklinkname0 tags. They bind to Go runtime internals and the pinned
# release does not link with Gust's newer Go toolchain when they are enabled.
tags=(
  with_singbox
  with_singbox_full
  with_gvisor
  with_quic
  with_dhcp
  with_wireguard
  with_utls
  with_acme
  with_clash_api
  with_tailscale
  with_ccm
  with_ocm
)

(IFS=,; echo "${tags[*]}")
