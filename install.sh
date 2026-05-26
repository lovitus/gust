#!/usr/bin/env bash
set -euo pipefail

if [[ "${EUID}" -ne 0 ]]; then
    echo "Error: you must run this script as root."
    exit 1
fi

repo="lovitus/gust"
base_url="https://api.github.com/repos/${repo}/releases"

usage() {
    echo "usage: $0 [--install|VERSION]" >&2
}

detect_os() {
    case "$(uname -s)" in
        Linux)
            echo "linux"
            ;;
        Darwin)
            echo "darwin"
            ;;
        FreeBSD)
            echo "freebsd"
            ;;
        MINGW*|MSYS*|CYGWIN*)
            echo "windows"
            ;;
        *)
            echo "Unsupported operating system." >&2
            exit 1
            ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)
            echo "amd64"
            ;;
        armv5*)
            echo "armv5"
            ;;
        armv6*)
            echo "armv6"
            ;;
        armv7*)
            echo "armv7"
            ;;
        aarch64|arm64)
            echo "arm64"
            ;;
        i386|i686)
            echo "386"
            ;;
        mips64le*)
            echo "mips64le"
            ;;
        mips64*)
            echo "mips64"
            ;;
        mipsle*)
            echo "mipsle-softfloat"
            ;;
        mips*)
            echo "mips-softfloat"
            ;;
        riscv64)
            echo "riscv64"
            ;;
        s390x)
            echo "s390x"
            ;;
        *)
            echo "Unsupported CPU architecture." >&2
            exit 1
            ;;
    esac
}

download_asset() {
    local tag="$1"
    local os="$2"
    local arch="$3"
    local version="${tag#v}"
    local ext="tar.gz"

    if [[ "${os}" == "windows" ]]; then
        ext="zip"
    fi

    local asset="gost-${os}-${arch}-${version}.${ext}"
    local url="https://github.com/${repo}/releases/download/${tag}/${asset}"
    local out="$4/${asset}"

    echo "Downloading ${asset}..." >&2
    curl -fL --retry 3 -o "${out}" "${url}"
    printf '%s\n' "${out}"
}

install_gost() {
    local tag="$1"
    if [[ "${tag}" != v* ]]; then
        tag="v${tag}"
    fi

    local os
    local arch
    os="$(detect_os)"
    arch="$(detect_arch)"

    local tmp_dir
    tmp_dir="$(mktemp -d)"
    trap "rm -rf '${tmp_dir}'" EXIT

    local archive
    archive="$(download_asset "${tag}" "${os}" "${arch}" "${tmp_dir}")"

    if [[ "${archive}" == *.zip ]]; then
        if ! command -v unzip >/dev/null 2>&1; then
            echo "Missing required command: unzip" >&2
            exit 1
        fi
        unzip -q "${archive}" -d "${tmp_dir}"
    else
        tar -xzf "${archive}" -C "${tmp_dir}"
    fi

    local binary
    binary="$(find "${tmp_dir}" -maxdepth 1 -type f -name 'gost-*' ! -name '*.tar.gz' ! -name '*.zip' | head -n 1)"
    if [[ -z "${binary}" ]]; then
        echo "Downloaded archive did not contain a gost binary." >&2
        exit 1
    fi

    local target="/usr/local/bin/gost"
    if [[ "${os}" == "windows" ]]; then
        target="/usr/local/bin/gost.exe"
    fi

    echo "Installing ${target}..."
    install -m 0755 "${binary}" "${target}"
    echo "gost ${tag} installed successfully."
}

versions="$(curl -fsSL "${base_url}" | sed -n 's/.*"tag_name": "\([^"]*\)".*/\1/p')"

if [[ "${1:-}" == "--install" ]]; then
    latest_version="$(printf '%s\n' "${versions}" | head -n 1)"
    if [[ -z "${latest_version}" ]]; then
        echo "No releases found for ${repo}." >&2
        exit 1
    fi
    install_gost "${latest_version}"
elif [[ $# -eq 1 ]]; then
    install_gost "$1"
elif [[ $# -eq 0 ]]; then
    echo "Available gust versions:"
    select version in ${versions}; do
        if [[ -n "${version:-}" ]]; then
            install_gost "${version}"
            break
        fi
        echo "Invalid choice. Please select a valid version."
    done
else
    usage
    exit 2
fi
