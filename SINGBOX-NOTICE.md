# gust-with-singbox distribution notice

This notice applies only to release assets named `gust-with-singbox-*`.
The standard `gust-*` assets do not link the sing-box runtime.

`gust-with-singbox` incorporates sing-box v1.13.16. The combined binary is
distributed under the GNU General Public License, version 3 or (at your
option) any later version, subject to sing-box's additional notice that no
derivative work may use its name or imply association without prior consent.
This project and its assets are named Gust and do not claim endorsement by or
association with SagerNet or sing-box.

The Linux asset also bundles the matching `libcronet.so` used by the Naive
outbound through cronet-go's pure-Go loader. Its upstream license notice and a
complete copy of GPLv3 are included in the archive.

Corresponding source and reproducible build inputs:

- Gust: https://github.com/lovitus/gust
- Gust implementation packages: https://github.com/lovitus/gust-x
- sing-box v1.13.16: https://github.com/SagerNet/sing-box/tree/v1.13.16
- cronet-go: https://github.com/sagernet/cronet-go
- all exact module versions: the `go.mod` and `go.sum` files in those source
  repositories

Each release archive contains `feature-manifest.json` with the exact Gust and
gust-x commits, Go version, target, build tags and unavailable features used
for that binary. The build command is equivalent to:

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -trimpath -tags "$(bash .github/scripts/singbox-tags.sh)" \
  -o gust-with-singbox ./cmd/gost
```

Keep `libcronet.so` beside `gust-with-singbox` when using the Naive outbound.
