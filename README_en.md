# Gust

> Fork of [go-gost/gost](https://github.com/go-gost/gost) at `c7d793c`.
> Implementation module: [lovitus/gust-x](https://github.com/lovitus/gust-x), based on [go-gost/x](https://github.com/go-gost/x) v0.15.2.

## Additional Features (vs upstream)

- **Porty Stealth Port Forwarding** - Disguises TCP port forwarding as normal HTTPS/WebSocket traffic with AEAD authentication, smux session caching, multi-port mixed-role sessions, optional P2P direct mode, dynamic `bindpath` WS bridge for Mihomo/iOS, and the standalone `portyd` server.
- **sings + Mihomo smux** - `sings` integrates sing-shadowsocks TCP/UoT/AEAD-2022 and now accepts sing-box/Mihomo common `smux` TCP multiplexing, reducing WSS connection counts when used behind Porty `bindpath`.
- **SSH Relay Fallback** - When SSH server disables TCP forwarding (`AllowTcpForwarding=no`), automatically falls back through multiplexed relay, embedded relay binary, or exec-based tools. Original direct-tcpip is always prioritized.
- **Escape-Based Passwords** - Supports backslash escapes and quotes in inline passwords, while remaining compatible with URL encoding.
- **SOCKS5 UDP First-Packet Pinning** - Supports split TCP/UDP source paths with opt-in first-packet source locking.
- **Cross-Platform Builds** - Release tags publish `gost` and `portyd` binaries for the supported platform matrix.

See [FORK_CHANGES.md](FORK_CHANGES.md) for technical details and upstream merge notes. Full Porty and sings documentation lives in [gust-x docs/porty.md](https://github.com/lovitus/gust-x/blob/master/docs/porty.md) and [gust-x docs/sings-protocol.md](https://github.com/lovitus/gust-x/blob/master/docs/sings-protocol.md).

Embedded sing-box users can run `gost -singboxmanual` for the offline manual.
The repository copies are the [sing-box user manual](cmd/gost/SINGBOX_MANUAL.md)
and the detailed [sing-box acceptance record](SINGBOX-VALIDATION.md).

Development and releases are permanently separated: standard builds use
`master` and `v*` tags, while embedded builds use `singbox-backend` and
`singbox-v*` tags. See [RELEASE.md](RELEASE.md) for the isolated workflows.

---

# GO Simple Tunnel

### A simple security tunnel written in golang

[![en](https://img.shields.io/badge/English%20README-green)](README_en.md) [![zh](https://img.shields.io/badge/Chinese%20README-gray)](README.md)

## Features

- [x] [Listening on multiple ports](https://gost.run/en/getting-started/quick-start/)
- [x] [Multi-level forwarding chain](https://gost.run/en/concepts/chain/)
- [x] Rich protocol
- [x] [TCP/UDP port forwarding](https://gost.run/en/tutorials/port-forwarding/)
- [x] [Reverse Proxy](https://gost.run/en/tutorials/reverse-proxy/) and [Tunnel](https://gost.run/en/tutorials/reverse-proxy-tunnel/)
- [x] [TCP/UDP transparent proxy](https://gost.run/en/tutorials/redirect/)
- [x] DNS [resolver](https://gost.run/en/concepts/resolver/) and [proxy](https://gost.run/en/tutorials/dns/)
- [x] [TUN/TAP device](https://gost.run/en/tutorials/tuntap/) and [TUN2SOCKS](https://gost.run/en/tutorials/tungo/)
- [x] [Load balancing](https://gost.run/en/concepts/selector/)
- [x] [Routing control](https://gost.run/en/concepts/bypass/)
- [x] [Admission control](https://gost.run/en/concepts/limiter/)
- [x] [Bandwidth/Rate Limiter](https://gost.run/en/concepts/limiter/)
- [x] [Plugin System](https://gost.run/en/concepts/plugin/)
- [x] [Prometheus metrics](https://gost.run/en/tutorials/metrics/)
- [x] [Dynamic configuration](https://gost.run/en/tutorials/api/config/)
- [x] [Web API](https://gost.run/en/tutorials/api/overview/)
- [x] [GUI](https://github.com/go-gost/gostctl)/[WebUI](https://github.com/go-gost/gost-ui)

## Overview

![Overview](https://gost.run/images/overview.png)

There are three main ways to use GOST as a tunnel.

### Proxy

As a proxy service to access the network, multiple protocols can be used in combination to form a forwarding chain for traffic forwarding.

![Proxy](https://gost.run/images/proxy.png)

### Port Forwarding

Mapping the port of one service to the port of another service, you can also use a combination of multiple protocols to form a forwarding chain for traffic forwarding.

![Forward](https://gost.run/images/forward.png)

### Reverse Proxy

Use tunnel and intranet penetration to expose local services behind NAT or firewall to public network for access.

![Reverse Proxy](https://gost.run/images/reverse-proxy.png)

## Installation

### Package managers

Package-manager entries are generated and updated by the next stable tag release. Maintainer release notes are in [RELEASE.md](RELEASE.md).

#### Homebrew (macOS/Linux)

```bash
brew tap lovitus/gust https://github.com/lovitus/gust
brew install gust
gost -V
```

#### Scoop (Windows)

```powershell
scoop bucket add gust https://github.com/lovitus/gust
scoop install gust
gost -V
```

#### APT (Debian/Ubuntu amd64/arm64)

```bash
arch="$(dpkg --print-architecture)"
case "$arch" in amd64|arm64) ;; *) echo "unsupported APT arch: $arch" >&2; exit 1;; esac
sudo install -d -m 0755 /etc/apt/keyrings
curl -fsSL https://lovitus.github.io/gust/apt/gust-archive-keyring.gpg \
  | sudo tee /etc/apt/keyrings/gust-archive-keyring.gpg >/dev/null
echo "deb [arch=$arch signed-by=/etc/apt/keyrings/gust-archive-keyring.gpg] https://lovitus.github.io/gust/apt stable main" \
  | sudo tee /etc/apt/sources.list.d/gust.list >/dev/null
sudo apt update
sudo apt install gust
```

#### YUM/DNF (x86_64/aarch64)

```bash
sudo tee /etc/yum.repos.d/gust.repo >/dev/null <<'EOF'
[gust]
name=gust stable repository
baseurl=https://lovitus.github.io/gust/rpm/$basearch
enabled=1
repo_gpgcheck=1
gpgcheck=0
gpgkey=https://lovitus.github.io/gust/rpm/RPM-GPG-KEY-gust
EOF

sudo dnf install gust
# or: sudo yum install gust
```

The RPM repository currently signs repository metadata only. Individual RPM packages are not signed yet, so the repo file uses `repo_gpgcheck=1` and `gpgcheck=0`.

### Binary files

[https://github.com/lovitus/gust/releases](https://github.com/lovitus/gust/releases)

### install script

```bash
# install latest
bash <(curl -fsSL https://raw.githubusercontent.com/lovitus/gust/master/install.sh) --install
```
```bash
# select version for install 
bash <(curl -fsSL https://raw.githubusercontent.com/lovitus/gust/master/install.sh)
```

### From source

```
git clone https://github.com/lovitus/gust.git
cd gust/cmd/gost
go build
```

### Docker

```
docker run --rm gogost/gost -V
```

## Tools

### GUI

[go-gost/gostctl](https://github.com/go-gost/gostctl)

### WebUI

[go-gost/gost-ui](https://github.com/go-gost/gost-ui)

### Shadowsocks Android

[hamid-nazari/ShadowsocksGostPlugin](https://github.com/hamid-nazari/ShadowsocksGostPlugin)

## Support

Wiki: [https://gost.run](https://gost.run/en/)

YouTube: [https://www.youtube.com/@gost-tunnel](https://www.youtube.com/@gost-tunnel)

Telegram: [https://t.me/gogost](https://t.me/gogost)

Google group: [https://groups.google.com/d/forum/go-gost](https://groups.google.com/d/forum/go-gost)

Legacy version: [v2.gost.run](https://v2.gost.run/en/)
