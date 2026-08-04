#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "usage: $0 --version VERSION --artifacts-dir DIR --out-dir DIR" >&2
}

version=""
artifacts_dir=""
out_dir=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      version="${2:-}"
      shift 2
      ;;
    --artifacts-dir)
      artifacts_dir="${2:-}"
      shift 2
      ;;
    --out-dir)
      out_dir="${2:-}"
      shift 2
      ;;
    *)
      usage
      exit 2
      ;;
  esac
done

if [[ -z "${version}" || -z "${artifacts_dir}" || -z "${out_dir}" ]]; then
  usage
  exit 2
fi
if [[ "${version}" == v* ]]; then
  echo "package version must not include leading v: ${version}" >&2
  exit 1
fi
if ! command -v nfpm >/dev/null 2>&1; then
  go_path="$(go env GOPATH)"
  export PATH="${go_path}/bin:${PATH}"
fi
if ! command -v nfpm >/dev/null 2>&1; then
  echo "missing required command: nfpm" >&2
  exit 1
fi

mkdir -p "${out_dir}"

work_dir="${RUNNER_TEMP:-/tmp}/gust-linux-packages-${version}"
rm -rf "${work_dir}"
mkdir -p "${work_dir}"

extract_binary() {
  local goarch="$1"
  local root="${work_dir}/linux-${goarch}"
  local archive="${artifacts_dir}/gost-linux-${goarch}-${version}.tar.gz"
  local binary="${root}/gost-linux-${goarch}"

  if [[ ! -f "${archive}" ]]; then
    echo "missing release archive: ${archive}" >&2
    exit 1
  fi

  mkdir -p "${root}"
  tar -xzf "${archive}" -C "${root}"
  if [[ ! -f "${binary}" ]]; then
    echo "archive ${archive} did not contain gost-linux-${goarch}" >&2
    exit 1
  fi
  chmod 0755 "${binary}"
  printf '%s\n' "${binary}"
}

build_package() {
  local goarch="$1"
  local deb_arch="$2"
  local rpm_arch="$3"
  local bin
  local config="${work_dir}/nfpm-${goarch}.yaml"

  bin="$(extract_binary "${goarch}")"

  cat > "${config}" <<EOF
name: gust
arch: ${goarch}
platform: linux
version: ${version}
release: "1"
section: utils
priority: optional
maintainer: gust maintainers <noreply@github.com>
vendor: gust
homepage: https://github.com/lovitus/gust
license: MIT
description: GOST fork with SSH relay fallback enhancements.
contents:
  - src: ${bin}
    dst: /usr/bin/gost
EOF

  nfpm pkg --config "${config}" --packager deb \
    --target "${out_dir}/gust_${version}_${deb_arch}.deb"
  nfpm pkg --config "${config}" --packager rpm \
    --target "${out_dir}/gust-${version}-1.${rpm_arch}.rpm"
}

build_package amd64 amd64 x86_64
build_package arm64 arm64 aarch64
