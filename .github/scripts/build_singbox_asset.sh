#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "usage: $0 --version VERSION --tag TAG --goos GOOS --goarch GOARCH --out-dir DIR" >&2
}

version=""
tag=""
target_goos=""
target_goarch=""
out_dir=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --version) version="${2:-}"; shift 2 ;;
    --tag) tag="${2:-}"; shift 2 ;;
    --goos) target_goos="${2:-}"; shift 2 ;;
    --goarch) target_goarch="${2:-}"; shift 2 ;;
    --out-dir) out_dir="${2:-}"; shift 2 ;;
    *) usage; exit 2 ;;
  esac
done
if [[ -z "${version}" || -z "${tag}" || -z "${target_goos}" || -z "${target_goarch}" || -z "${out_dir}" ]]; then
  usage
  exit 2
fi
if [[ ! -d ../gust-x ]]; then
  echo "gust-x must be checked out next to gust" >&2
  exit 1
fi

case "${target_goos}/${target_goarch}" in
  linux/amd64|linux/arm64|windows/amd64|windows/arm64|darwin/amd64|darwin/arm64) ;;
  *) echo "unsupported sing-box release target: ${target_goos}/${target_goarch}" >&2; exit 1 ;;
esac

tags="$(bash .github/scripts/singbox-tags.sh --target "${target_goos}" "${target_goarch}")"
archive_root="gust-with-singbox-${target_goos}-${target_goarch}"
stage="${out_dir}/${archive_root}"
if [[ -e "${stage}" ]]; then
  echo "asset staging path already exists: ${stage}" >&2
  exit 1
fi
mkdir -p "${stage}"

extension=""
if [[ "${target_goos}" == windows ]]; then
  extension=.exe
fi
binary_name="gust-with-singbox${extension}"
binary_path="${stage}/${binary_name}"

CGO_ENABLED=0 GOOS="${target_goos}" GOARCH="${target_goarch}" \
  go build -trimpath -tags "${tags}" \
  -ldflags="-s -w -X main.version=${tag}" \
  -o "${binary_path}" ./cmd/gost

go version -m "${binary_path}" | grep -F $'github.com/sagernet/sing-box\t'

singbox_dir="$(cd ../gust-x && go list -m -f '{{.Dir}}' github.com/sagernet/sing-box)"
cp "${singbox_dir}/LICENSE" "${stage}/sing-box-LICENSE"
cp SINGBOX-NOTICE.md "${stage}/SINGBOX-NOTICE.md"
curl --fail --silent --show-error --location \
  https://www.gnu.org/licenses/gpl-3.0.txt \
  -o "${stage}/GPL-3.0.txt"
gpl_digest="$(openssl dgst -sha256 "${stage}/GPL-3.0.txt" | awk '{print $NF}')"
if [[ "${gpl_digest}" != 3972dc9744f6499f0f9b2dbf76696f2ae7ad8af9b23dde66d6af86c9dfb36986 ]]; then
  echo "GPL-3.0.txt checksum mismatch: ${gpl_digest}" >&2
  exit 1
fi

runtime_files=("${binary_name}")
unavailable_features=(badlinkname tfogo_checklinkname0)
case "${target_goos}" in
  linux)
    cronet_module="github.com/sagernet/cronet-go/lib/linux_${target_goarch}"
    cronet_name=libcronet.so
    ;;
  windows)
    cronet_module="github.com/sagernet/cronet-go/lib/windows_${target_goarch}"
    cronet_name=libcronet.dll
    ;;
  darwin)
    cronet_module=""
    cronet_name=""
    unavailable_features+=(naive ccm)
    ;;
esac
if [[ -n "${cronet_module}" ]]; then
  cronet_dir="$(cd ../gust-x && go list -m -f '{{.Dir}}' "${cronet_module}")"
  cp "${cronet_dir}/LICENSE" "${stage}/cronet-go-LICENSE"
  cp "${cronet_dir}/${cronet_name}" "${stage}/${cronet_name}"
  runtime_files+=("${cronet_name}")
fi

host_goos="$(go env GOOS)"
host_goarch="$(go env GOARCH)"
if [[ "${host_goos}/${host_goarch}" == "${target_goos}/${target_goarch}" ]]; then
  if [[ "${target_goos}" == linux ]]; then
    LD_LIBRARY_PATH="${stage}" "${binary_path}" -V | tee "${stage}/version.txt"
  else
    "${binary_path}" -V | tee "${stage}/version.txt"
  fi
  grep -F 'flavor=singbox' "${stage}/version.txt"
fi

gust_commit="$(git rev-parse HEAD)"
gust_x_commit="$(git -C ../gust-x rev-parse HEAD)"
go_build_version="$(go env GOVERSION)"
runtime_json="$(printf '%s\n' "${runtime_files[@]}" | python3 -c 'import json,sys; print(json.dumps([line.rstrip("\n") for line in sys.stdin if line.rstrip("\n")]))')"
unavailable_json="$(printf '%s\n' "${unavailable_features[@]}" | python3 -c 'import json,sys; print(json.dumps([line.rstrip("\n") for line in sys.stdin if line.rstrip("\n")]))')"
export GUST_COMMIT="${gust_commit}" GUST_X_COMMIT="${gust_x_commit}" GO_BUILD_VERSION="${go_build_version}"
export SINGBOX_TAGS="${tags}" TARGET_GOOS="${target_goos}" TARGET_GOARCH="${target_goarch}"
export RUNTIME_FILES_JSON="${runtime_json}" UNAVAILABLE_FEATURES_JSON="${unavailable_json}"
python3 - <<'PY' > "${stage}/feature-manifest.json"
import json
import os

print(json.dumps({
    "flavor": "singbox",
    "singBoxVersion": "v1.13.16",
    "goVersion": os.environ["GO_BUILD_VERSION"],
    "goos": os.environ["TARGET_GOOS"],
    "goarch": os.environ["TARGET_GOARCH"],
    "buildTags": os.environ["SINGBOX_TAGS"].split(","),
    "gustCommit": os.environ["GUST_COMMIT"],
    "gustXCommit": os.environ["GUST_X_COMMIT"],
    "runtimeFiles": json.loads(os.environ["RUNTIME_FILES_JSON"]),
    "unavailableFeatures": json.loads(os.environ["UNAVAILABLE_FEATURES_JSON"]),
}, indent=2, sort_keys=True))
PY

python3 -m json.tool "${stage}/feature-manifest.json" >/dev/null
if [[ "${target_goos}" == windows ]]; then
  (cd "${out_dir}" && zip -qr "${archive_root}-${version}.zip" "${archive_root}")
else
  tar -czf "${out_dir}/${archive_root}-${version}.tar.gz" -C "${out_dir}" "${archive_root}"
fi
find "${stage}" -depth -delete
