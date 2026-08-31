# Gust

> Fork of [go-gost/gost](https://github.com/go-gost/gost) at `c7d793c` with SSH relay fallback enhancements.
> Implementation module: [lovitus/gust-x](https://github.com/lovitus/gust-x), based on [go-gost/x](https://github.com/go-gost/x) v0.15.2.

## Additional Features (vs upstream)

- **Porty Stealth Port Forwarding** - Disguises TCP port forwarding as normal HTTPS/WebSocket traffic with AEAD authentication, smux session caching, multi-port mixed-role sessions, optional P2P direct mode, dynamic `bindpath` WS bridge for Mihomo/iOS, and standalone `portyd`/`portyc` binaries. The SSH-exit mode keeps SSH credentials in the access-side Gust process while `portyc` only relays verified SSH transports.
- **sings + Mihomo smux** - `sings` integrates sing-shadowsocks TCP/UoT/AEAD-2022 and now accepts sing-box/Mihomo common `smux` TCP multiplexing, which reduces WSS connection counts when used behind Porty `bindpath`.
- **SSH Relay Fallback** - When SSH server disables TCP forwarding (`AllowTcpForwarding=no`), automatically falls back through multiplexed relay, embedded relay binary, or exec-based tools (nc/socat/perl/python/bash). Original direct-tcpip is always prioritized.
- **Multiplexed Relay** - Single SSH exec session handles unlimited TCP connections via mux protocol, bypassing `MaxSessions` limits.
- **Smart Relay Upload** - Embedded relay binary auto-uploaded to remote server, hash-cached (only transferred once per binary version).
- **Escape-Based Passwords** - Supports backslash escapes and quotes in inline passwords (backward compatible with URL encoding).
- **SOCKS5 UDP First-Packet Pinning** - Supports split TCP/UDP source paths with opt-in first-packet source locking.
- **Cross-Platform Builds** - Pre-built binaries for 23 platform/architecture combinations.

See [FORK_CHANGES.md](FORK_CHANGES.md) for detailed technical documentation and upstream merge notes. Full Porty and sings documentation lives in [gust-x docs/porty.md](https://github.com/lovitus/gust-x/blob/master/docs/porty.md) and [gust-x docs/sings-protocol.md](https://github.com/lovitus/gust-x/blob/master/docs/sings-protocol.md).

维护规则：所有通用能力必须先在 `master` 实现；`singbox-backend` 只是在
`master` 之上的嵌入式 sing-box 扩展。分支同步、PR 目标和 tag 规则见
[BRANCH_POLICY.md](BRANCH_POLICY.md)。

---

# GO Simple Tunnel (Upstream)

### GO语言实现的安全隧道

[![zh](https://img.shields.io/badge/Chinese%20README-green)](README.md) [![en](https://img.shields.io/badge/English%20README-gray)](README_en.md)

## 功能特性

- [x] [多端口监听](https://gost.run/getting-started/quick-start/)
- [x] [多级转发链](https://gost.run/concepts/chain/)
- [x] [多协议支持](https://gost.run/tutorials/protocols/overview/)
- [x] [TCP/UDP端口转发](https://gost.run/tutorials/port-forwarding/)
- [x] [反向代理](https://gost.run/tutorials/reverse-proxy/)和[隧道](https://gost.run/tutorials/reverse-proxy-tunnel/)
- [x] [TCP/UDP透明代理](https://gost.run/tutorials/redirect/)
- [x] DNS[解析](https://gost.run/concepts/resolver/)和[代理](https://gost.run/tutorials/dns/)
- [x] [TUN/TAP设备](https://gost.run/tutorials/tuntap/)与[TUN2SOCKS](https://gost.run/tutorials/tungo/)
- [x] [负载均衡](https://gost.run/concepts/selector/)
- [x] [路由控制](https://gost.run/concepts/bypass/)
- [x] [准入控制](https://gost.run/concepts/admission/)
- [x] [限速限流](https://gost.run/concepts/limiter/)
- [x] [插件系统](https://gost.run/concepts/plugin/)
- [x] [Prometheus监控指标](https://gost.run/tutorials/metrics/)
- [x] [动态配置](https://gost.run/tutorials/api/config/)
- [x] [Web API](https://gost.run/tutorials/api/overview/)
- [x] [GUI](https://github.com/go-gost/gostctl)/[WebUI](https://github.com/go-gost/gost-ui)

## 概览

![Overview](https://gost.run/images/overview.png)

GOST作为隧道有三种主要使用方式。

### 正向代理

作为代理服务访问网络，可以组合使用多种协议组成转发链进行转发。

![Proxy](https://gost.run/images/proxy.png)

### 端口转发

将一个服务的端口映射到另外一个服务的端口，同样可以组合使用多种协议组成转发链进行转发。

![Forward](https://gost.run/images/forward.png)

### 反向代理

利用隧道和内网穿透将内网服务暴露到公网访问。

![Reverse Proxy](https://gost.run/images/reverse-proxy.png)

## 下载安装

### 包管理器

包管理器入口会在下一个稳定 tag 发布后自动生成和更新；维护者发布流程见 [RELEASE.md](RELEASE.md)。

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
# 或: sudo yum install gust
```

RPM 源当前签名的是仓库 metadata，暂不对单个 RPM 包签名，因此 repo 配置中使用 `repo_gpgcheck=1` 和 `gpgcheck=0`。

### 二进制文件

[https://github.com/lovitus/gust/releases](https://github.com/lovitus/gust/releases)

### 安装脚本

```bash
# 安装最新版本
bash <(curl -fsSL https://raw.githubusercontent.com/lovitus/gust/master/install.sh) --install
```
```bash
# 选择要安装的版本
bash <(curl -fsSL https://raw.githubusercontent.com/lovitus/gust/master/install.sh)
```

### 源码编译

```
git clone https://github.com/lovitus/gust.git
git clone https://github.com/lovitus/gust-x.git
cd gust
go build -trimpath -o gust ./cmd/gost
```

`gust` 的 `go.mod` 通过 `../gust-x` 使用配套实现仓库，因此两个仓库必须保持上述
同级目录结构，并使用同名产品分支。构建 `singbox-backend` 时应同时切换两个仓库，
再按 `SINGBOX-NOTICE.md` 中的固定 tags 构建。

### Docker

```
docker run --rm gogost/gost -V
```

## 工具

### GUI

[go-gost/gostctl](https://github.com/go-gost/gostctl)

### WebUI

[go-gost/gost-ui](https://github.com/go-gost/gost-ui)

### Shadowsocks Android插件

[hamid-nazari/ShadowsocksGostPlugin](https://github.com/hamid-nazari/ShadowsocksGostPlugin)

## 帮助与支持

Wiki站点：[https://gost.run](https://gost.run)

YouTube: [https://www.youtube.com/@gost-tunnel](https://www.youtube.com/@gost-tunnel)

Telegram：[https://t.me/gogost](https://t.me/gogost)

Google讨论组：[https://groups.google.com/d/forum/go-gost](https://groups.google.com/d/forum/go-gost)

旧版入口：[v2.gost.run](https://v2.gost.run)
