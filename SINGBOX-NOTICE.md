# gust-with-singbox distribution notice

This notice applies only to release assets named `gust-with-singbox-*`.
The standard `gust-*` assets do not link the sing-box runtime.

`gust-with-singbox` incorporates maintainer-fork pseudo-version
`v1.13.17-0.20260811025254-2926c94dd073`, based on sing-box v1.13.16. The combined binary is
distributed under the GNU General Public License, version 3 or (at your
option) any later version, subject to sing-box's additional notice that no
derivative work may use its name or imply association without prior consent.
This project and its assets are named Gust and do not claim endorsement by or
association with SagerNet or sing-box.

Linux assets bundle the matching `libcronet.so` and Windows assets bundle the
matching `libcronet.dll` required by the compiled full-feature native graph.
The embedded resource policy rejects Naive outbound on every platform because
the pinned path requires an IPC/loopback bridge; Naive inbound remains
available and does not use Cronet. Darwin assets also omit the Naive outbound
tag and CCM because the
pinned cronet-go release supplies static Darwin libraries while Gust's
reproducible release build uses `CGO_ENABLED=0`. Every archive records this
platform feature boundary in `feature-manifest.json` and includes the relevant
upstream license notices and a complete copy of GPLv3.

Corresponding source and reproducible build inputs:

- Gust: https://github.com/lovitus/gust
- Gust implementation packages: https://github.com/lovitus/gust-x
- sing-box upstream base v1.13.16: https://github.com/SagerNet/sing-box/tree/v1.13.16
- cronet-go: https://github.com/sagernet/cronet-go
- all exact module versions: the `go.mod` and `go.sum` files in those source
  repositories

Each release archive contains `feature-manifest.json` with the exact Gust and
gust-x commits, Go version, target, build tags and unavailable features used
for that binary. The build command is equivalent to:

```sh
TAG=singbox-vX.Y.Z
CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" \
  go build -trimpath -tags "$(bash .github/scripts/singbox-tags.sh --target "$GOOS" "$GOARCH")" \
  -ldflags="-s -w -X main.version=${TAG}" \
  -o gust-with-singbox ./cmd/gost
```

Keep any bundled `libcronet.so` or `libcronet.dll` beside the matching
`gust-with-singbox` executable. Its presence is a runtime packaging
requirement for the compiled graph, not permission to activate Naive outbound.
