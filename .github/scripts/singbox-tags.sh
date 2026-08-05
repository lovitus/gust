#!/usr/bin/env bash

set -euo pipefail

# Gust deliberately excludes sing-box's badlinkname and
# tfogo_checklinkname0 tags. They bind to Go runtime internals and the pinned
# release does not link with Gust's newer Go toolchain when they are enabled.
full_tags=(
  with_singbox
  with_singbox_full
  with_gvisor
  with_quic
  with_dhcp
  with_wireguard
  with_utls
  with_naive_outbound
  with_purego
  with_acme
  with_clash_api
  with_tailscale
  with_ccm
  with_ocm
)

limited_tags=(
  with_singbox
  with_singbox_limited
  with_gvisor
  with_quic
  with_dhcp
  with_wireguard
  with_utls
  with_acme
  with_clash_api
  with_tailscale
  with_ocm
)

if [[ $# -eq 0 ]]; then
  tags=("${full_tags[@]}")
elif [[ $# -eq 3 && "$1" == "--target" ]]; then
  case "$2" in
    linux|windows)
      tags=("${full_tags[@]}")
      ;;
    darwin)
      # cronet-go publishes static Darwin libraries. The reproducible
      # CGO_ENABLED=0 release flavor therefore omits Naive and CCM instead of
      # advertising features that compile but cannot start at runtime.
      tags=("${limited_tags[@]}")
      ;;
    *)
      echo "unsupported sing-box release target: $2/$3" >&2
      exit 1
      ;;
  esac
else
  echo "usage: $0 [--target GOOS GOARCH]" >&2
  exit 2
fi

(IFS=,; echo "${tags[*]}")
