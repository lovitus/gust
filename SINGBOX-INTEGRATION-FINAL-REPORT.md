# Gust 与 sing-box 融合目标最终技术总账

更新：2026-08-13

本文是本轮 Gust / gust-x / sing-box 融合目标的仓库内总账。它回答四个问题：

1. 最终真正合并了什么；
2. 代码的数据路径、配置路径、生命周期和安全边界如何工作；
3. 哪些大规模探索被保留为证据但没有进入正式分支；
4. 未来如何继续追平 GOST、gust-x 和 sing-box，而不把三个项目缠成不可维护的 fork。

用户语法以 [离线手册](cmd/gost/SINGBOX_MANUAL.md) 为准；架构约束以
[架构文档](SINGBOX-ARCHITECTURE.md) 为准；逐协议与性能证据以
[验证记录](SINGBOX-VALIDATION.md) 为准。本文侧重最终代码逻辑、选择理由和维护地图。

## 1. 完成状态与最终边界

本轮功能目标已完成，已认证功能基线无已知 blocker；报告、能力说明、CI 和离线恢复
材料也已完成同等级收口。没有自动创建 release tag；是否发布 `singbox-v*` 仍是独立
维护者决定。会随维护推进变化的 branch HEAD 和最新 workflow run 不在本文递归声明为
“永远最终”，应以远端 `singbox-backend`、GitHub Actions 和维护者恢复状态入口为准。

可复算的已认证基线：

```text
go-gost/core logger fix  6474e707cba7ccc968123795132c3b17dda42a17
Gust master              f0ae76d2d3be998d7990ef1a1a9a89cadf58f013
gust-x master            076218ba6f006b1d5714c57cf3912d3d96465f95
Gust functional code     aacea402ab96d3d414c20a0f9baa9fa7b7780abb
Gust reporting/policy    00df0af8c8e3ef428a02096a3a594ca32140d132
gust-x reporting/policy  6388338da2de795f0eccd5d26e82b01308310899
```

`00df0af` 与 `6388338` 校正了报告、能力说明、CI/asset policy 断言和手册回归，但没有
修改 embedded runtime 数据面。此后的纯文档校正也不会改变表中的功能基线；每个实际
branch HEAD 仍必须独立满足 clean、pushed、green 后才可发布。

分支关系固定为单向：

```text
GOST / gust-x upstream
          |
          v
       master  ----通用修复---->
          |
          | 仅向前合并
          v
  singbox-backend  ----仅含嵌入式扩展---->

qtui-route-manage 是另一条下游产品线，不进入上述两条线。
```

`master` 不包含 `backend/singbox`、sing-box build tag、native option 或资源 manifest；
`singbox-backend` 永远不得反向合回 `master`。这个约束由
[BRANCH_POLICY.md](BRANCH_POLICY.md) 和 CI 强制。

### 最终保留与隔离的代码规模

| 范围 | 最终处理 | 规模 | 说明 |
|---|---|---:|---|
| Gust master 安全 reload | 合并 | 6 文件，`+726/-140` | 4 个线性提交 |
| gust-x master 安全 reload | 合并 | 19 文件，`+1828/-92` | 6 个线性提交 |
| sing-box 运行时与本轮回归范围 | 合并到扩展分支 | 89 文件，`+9408/-673` | audited `9d3370f..b8c281c` integration/carried-master range |
| broad master transaction 实验 | **未合并** | 约 161 文件，`+14946/-2396` | 拆解后只取已闭环小栈，其余保留证据 |

因此“一万多行”并非整体塞进 master。大部分 sing-box 专属代码和回归测试留在扩展
分支；广泛改写 listener、quota、probe、router、service retention 的实验没有成为
发布基线。

## 2. 最终进程内架构

这不是调用外部 `sing-box` 可执行文件，也不是建立 localhost SOCKS/HTTP 中转。

```text
出站 -F
application -> GOST listener/handler -> GOST chain/prefix nodes
            -> sing-box self-dialing Transport
            -> selected native outbound/endpoint -> destination

入站 -L
native client -> selected sing-box inbound (service-owned Box)
              -> native route.rules (如果显式命中)
              -> __gust_egress -> GOST Router/chain -> destination
```

标准 flavor 不链接 sing-box runtime。build-neutral API 位于 gust-x
`backend/singbox`；只有 `with_singbox` 文件导入 native runtime。标准二进制可解析并
导出结构化 `+singbox` 配置，但真正启动时返回可识别的 unavailable 错误。

### 2.1 `-F`：self-dialing native 出站

CLI 把 `singbox://` 或 `protocol+singbox://` 解析成 `TransportSpec`。节点 parser 在
普通 Connector+Dialer 之前识别它；chain 遇到 self-dialing transport 时，把已选择的
前置 GOST route 交给 transport，并从当前位置切开 route 段。实际拨号只走
`Transport.Connect`，不会先创建一个普通 socket 再让 sing-box handshake。

如果节点没有前置 GOST route，启动原始 native 配置。若有前置 route，则对配置的
安全副本注入内部 `__gust_prefix` outbound：

- 无自有语义的 native `direct` leaf 被替换为内部 prefix leaf；
- 其他空 detour 网络 leaf 只有在 native socket ownership 和 resolver 语义可保持时
  才连接到 `__gust_prefix`；
- 用户显式 detour 不覆盖；
- DNS、NTP、remote rule-set、selector/urltest members 和其他后台网络 leaf 也必须
  证明能到达同一 prefix；
- route context 缺失或任一 leaf 无法认证时直接报错，绝不系统直连兜底。

每条不同的前置 route 产生独立 scope。这样不同 GOST 路径不会误共享 sing-box
session、mux 或 endpoint 状态；相同 scope 则复用已启动 runtime。

### 2.2 `-L`：native 入站接回 GOST

`-L protocol+singbox://...` 创建一个顶层 `service.Service` 和 service-owned Box。
full config 默认只保留选中的 inbound；额外 inbound 必须由 `activate_inbounds` 明确给出
且恰好等于选中 inbound 的完整 detour 依赖闭包。

构建有效配置时：

- 用户非内部 `route.final` 被拒绝，避免默认流量绕开 Gust；
- 用户显式 `route.rules` 保留优先级；
- 未命中流量进入内部 `__gust_egress`；
- `__gust_egress` 直接调用普通 GOST Router 的 TCP/UDP 路径；
- Reality/ShadowTLS handshake、remote DNS、NTP、remote rule-set 等空 detour 后台
  leaf 也接到同一 egress；
- 构造后还校验 inbound detour 实际是否具备所声明的 TCP/UDP injection 能力。

一个 `-L` 对应一个 Box。这是 socket、系统资源、reload 和失败隔离的正确性基线；
多个 listener 可共用同一 GOST chain 配置，但不自动合并 native Boxes。

## 3. 配置与类型系统

`-L` 与 `-F` 共用 source loader、URI lexer 和 merge 模型，但使用方向独立的 native
schema：

```text
selected full-config object
  < json/json64 object overlay
  < URI authority and userinfo
  < ordinary query assignments
```

关键规则：

- 自定义 lexer 保留 JSON 中的 `&`、`#`、`+`，不依赖会破坏这些值的普通
  `url.Query()`；
- `path=value` 根据 pinned native option 类型转换 bool、number、duration、array；
- `path:=JSON` 用于 object、array、null 或要求精确类型的值；
- 嵌套对象和数组索引由 typed path engine 处理；未知字段、错误方向和错误类型立即失败；
- 本地 JSON source 上限 16 MiB，只接受一个完整 JSON 值；
- JSON number 以 `json.Number` 保留，合法 `uint64` 不经过 `float64`；
- adapter controls（`inbound`、`outbound`、`endpoint`、`activate_*`）在 native decode
  前移除；
- native typed decode 后重新生成 canonical JSON，相同有效语义得到稳定 pool key；
  原始空白、注释和字段顺序不是保真目标。

## 4. 静态 Prepare 与资源授权

完整 sing-box 配置可以隐含 listener、文件、证书 watcher、控制 API、系统接口、DNS、
进程检查或后台网络活动。直接 `box.New` 无法证明这些副作用属于用户选中的节点，所以
所有 embedded Box 在构造前必须经过静态 Prepare：

1. 用 pinned native schema decode；
2. 校验生成 registry manifest 与固定版本一致；
3. 检查 tag、detour、DNS、route、rule-set、service、endpoint 引用；
4. 抽取脱敏 `ActivationManifest`；
5. 建立 typed dependency graph；
6. 从选中对象、固有 runtime 资源和显式 `activate_resources` 求闭包；
7. 若有效配置还含闭包外资源，在 `box.New` 前拒绝。

manifest 只记录稳定 positional ID、kind、ownership、network、address scope、port 和
effect，不记录用户 tag、host、path、URL、密钥或密码。授权 child 会补描述性 parent，
但不会从 parent 横向扩展到 sibling。

### 有意 fail-closed 的配置

以下不是“sing-box 没有实现”，而是 pinned native 行为无法满足 embedded 隔离或 GOST
route 所有权，因此当前适配器明确拒绝：

- prefix 缺失、依赖图不能完整接回、native socket ownership 与 prefix 冲突；
- 用户占用内部 `__gust_*` tag/type；
- full-config `-F` 含 inbound 或 `-L` 拥有冲突的 foreign `route.final`；
- 进程级 debug、system NTP write、`protect_path` IPC、Tor subprocess；
- Naive **outbound** 的 IPC/loopback bridge（Naive inbound 仍可用）；
- Tailscale endpoint 的进程全局 hooks、DHCP DNS 的 UDP/68 broadcast、resolved D-Bus；
- system proxy、不可关闭/泄露凭据或没有认证 extractor 的 native service；
- 有前置 GOST route 时无法保持 packet-socket 模型的 WireGuard endpoint；
- 未映射的 GOST service/node controls、动态入站端口和 outbound `Bind`。

错误经过 adapter 与依赖层两级脱敏。password、token、UUID、私钥和完整认证 header
不会出现在普通错误中；`-O` 是用户主动导出完整配置，仍可能含秘密。

## 5. TCP、UDP、DNS 与生命周期

### TCP

- 后端返回 native `net.Conn`，融合层不为每个 payload 增加额外代理 copy；
- connection setup 继承请求 cancel/deadline，建连成功后解除短 dial deadline，防止
  HTTP/2、gRPC、Cronet stream 被误取消；
- wrapper 透传 half-close；
- nil native addresses 使用稳定占位地址，避免元数据记录 panic；
- selected handle 和每条连接分别持 runtime lease，旧配置 handle 关闭不会立刻切断
  已建立连接。

### UDP 与 DNS

- 代理 outbound 保留 sing `Socksaddr` 域名到远端，不在 Gust adapter 提前解析；
- native direct UDP 在确实需要 `net.UDPAddr` 时才本机解析；
- connected prefix association 必须有固定目标，后续 `WriteTo` 不能偷换 destination；
- 支持 connected 与 unconnected packet 模式和一个 association 的多目标语义；
- GOST SOCKS 的无固定目标 UDP association 遇到经典 VMess 时，按实际 `WriteTo`
  目标惰性创建并复用有界的固定目标 packet session；默认 VMess UDP 不需要 XUDP，
  也不会被适配器静默改写 `packet_encoding`；
- proxy read 预留 512-byte address header headroom，使用有界 buffer pool；最终还有一次
  payload copy，但不再每包按 payload 大小 heap allocate；
- internal egress 只有在底层是真实 UDP socket 时才把域名解析成 `UDPAddr`；代理型
  PacketConn 继续接收域名 `Socksaddr`；
- implicit DNS 会物化成安全的 `prefer_go=true` local transport；用户显式 local DNS
  只有在平台条件安全时可用于 unscoped 配置，不能接入 prefix。需要 D-Bus、DHCP 或
  无法接回 prefix 的组合直接拒绝。

### runtime pool 与 registry

pool key 包含 canonical config、source kind、pinned version、feature set 和 prefix scope。
创建使用 singleflight；最后一个 handle/connection 释放后才关闭 Box，Close 有上限。

native registry 只初始化一次，但 backend 与 inbound 使用不同 outbound registry view：
前者注册 `__gust_prefix`，后者注册 `__gust_egress`。每次 Box 构造显式覆盖 context 的
registry view，避免入站/出站或不同 Box 相互污染。

## 6. `master` 中真正保留的通用修复

sing-box 审计暴露了 reload 发布顺序和动态引用降级问题，但最终只把通用、小范围、
有独立价值的修复放入 `master`。

### gust-x master

- `Config.Clone/Snapshot` 深拷贝候选与发布配置，保留具体 interface 类型、`int64` 和
  `time.Time`；`Global/Set` 不再共享可变对象；
- routing matcher 的 admission/bypass 引用提取与运行时共用同一 parser；
- `loader.Prepare` 在不碰 live registry 的候选副本上检查空名、重名、factory、TLS 和
  `references.go` 当前枚举覆盖的显式动态 registry 引用；
- candidate 引用只以 candidate 自身声明为准，缺 auther/limiter/router 等不能悄悄
  降级成无认证、无限流或 direct；
- `Prepared.Commit` 先在锁内应用并绑定 candidate loader runtime，再调用 process
  finalizer；全部成功后才 `config.Set(candidate)`，否则从上一 published snapshot
  重建，并聚合原错、loader 恢复错和 process 恢复错；这不是旁路构造后原子换指针；
- API reload 委托同一 process transaction；standalone API 也走同一 loader 锁；
- 解析中途未注册的 closer、listener、service 会清理。

### Gust master

- Start、SIGHUP、定时 reload 和 HTTP API reload 共用 `activateConfig`；Stop 共用同一
  loader 锁，但走独立的 deactivate 路径；
- API、metrics、profiling 与 proxy services 一起构成发布边界；
- profiling 改为同步 `net.Listen`，端口冲突不再在“reload success”之后异步暴露；
- API 自己触发同端口 reload 时用 retire 保留当前请求，候选成功后返回 200；候选失败
  时恢复旧 API/config 并返回错误；
- 没变的 process-owned API/metrics/profiling 对象可保留；loader registry 对象可重建；
- core 默认 logger 使用固定 holder 原子类型，不同 logger 实现热切换不再因
  `atomic.Value` concrete type 不一致而 panic；
- CI 的 Gust/gust-x candidate pairing、Mihomo UoT 和目标分支 policy 检查一致。

这是一项 failure-recovery / publication-order 保证，不是零停机或对象保留保证。同地址
proxy service reload 可有短暂中断；端口被外部进程抢占等外部世界变化可能让 rollback
失败，但错误会完整返回。完整 publication/recovery 保证属于 `Prepare/Commit` 路径；
legacy `loader.Load` 只被串行化，仍是直接 apply。Prepare 也不能预知端口占用、插件连接
或其他 activation-time 外部资源结果。

## 7. 三个依赖 fork：窄补丁而非整体魔改

业务代码仍 import `github.com/sagernet/...`；Gust 与 gust-x 两个主模块各自用三条
独立 `replace` 指向公开维护者 fork。因此每个模块的 replace 可以成对、独立解除。

| 模块 | 固定 ref | 补丁 | 规模 |
|---|---|---|---:|
| `sing` | `6340931c5367a32f3d90d9fc0efb7b87bb88e673` | badjson 精确 number；HTTP/SOCKS 认证错误脱敏 | 11 文件，`+378/-16` |
| `sing-box` | `2926c94dd0730f99cced92c56cb2d80a5625c93e` | v2ray HTTP/gRPC late setup 与 Close 同步 | 4 文件，`+176/-4` |
| `sing-vmess` | `1b9f30f9d98e524b351e82f9c596b4129af23f45` | Vision TLS/uTLS 用 `unsafe.Pointer`/`unsafe.Add` 通过 checkptr | 2 文件，`+9/-9` |

`sing` 的 number 修复覆盖 Decode、JSONObject、JSONArray、TypedMap、merge、omitempty
和 excluded paths，并拒绝 trailing JSON。认证修复只改变错误文案，不改变成功后的用户
context。

`sing-box` 为 HTTP2Conn/GunConn 的 setup、Read、Close 共享状态加同步：Close 先发生
时，late reader 立即关闭并固定返回 `net.ErrClosed`，避免资源泄漏或 use-after-close。

`sing-vmess` 不改变 Vision 内存布局，只避免把指针跨表达式转成 `uintptr` 再计算 offset。
它仍依赖反射访问 TLS/uTLS 私有字段；Go、crypto/tls、uTLS 或 sing-vmess 升级时必须
重新运行标准 TLS/uTLS race/checkptr 等价测试。

## 8. 验证证据

最终 GitHub 与 runner 证据：

| Gate | Subject | Result |
|---|---|---|
| Gust master CI | `f0ae76d`, run 31643007072 | 6/6 success |
| gust-x master policy | `076218b`, run 31643004524 | success |
| singbox functional CI | `aacea402`, run 31646252480 | 19/19 success |
| 报告/策略基线 | `00df0af`, run 31661558459 | 19/19 success |
| pinned/latest compatibility | `00df0af`, run 31661558468 | both success |
| gust-x 报告/策略基线 | `6388338`, run 31661536717 | success |

19-job singbox workflow 包含 3 个 test suites、4 个 Docker E2E shards、Mihomo UoT、
5 个 cross-build 和 6 个 native platform smokes。embedded suite 还包含 full tags、Linux
TUN/REDIRECT/TProxy 和 fail-closed policy checks。

真实 data-plane 矩阵覆盖 native inbounds/outbounds 的 Shadowsocks、SOCKS、HTTP、
Mixed、VMess、VLESS、Trojan、AnyTLS、Hysteria/Hysteria2、TUIC、SSH、Reality/Vision、
ShadowTLS、WireGuard endpoint、Direct，以及 Linux TUN、REDIRECT、TProxy。PASS 指真实
握手和应用 payload/datagram/DNS query，不是仅能编译、解析或打开端口。

私有 runner 只在公开材料中称 A/B/C；高性能 runner C 优先承担 Linux/Docker/netns
和 dependency remote equivalence 等严格机器门禁，性能则记录为 generic fixed Linux
runner。原始地址和日志不进入 Git 仓库。最终证据不是把一次旧 runner 日志冒充后续
提交的全量重跑，而是：机器专属 gate、绑定 `318b6c6` 功能树的性能基线，以及绑定
报告/策略基线的 GitHub 19/19 组合。后续若 runtime、依赖或性能相关配置发生变化，
必须重建对应证据，不能沿用本表。

### 固定 runner 性能

固定性能基线绑定 gust-x 功能树 `318b6c6`、Go 1.26.5、完整 tags、CGO disabled：

- TCP：同一 pinned sing-box 实现的 native direct baseline 677.39 MB/s，Gust
  674.93 MB/s，99.64%；
- UDP PPS：Gust / 同一 pinned native direct baseline = 99.11%；
- TCP p95/p99：99.73% / 99.78%；UDP p95/p99：100.43% / 101.12%；
- retained handle：220.1 ns/op、80 B、1 allocation；
- scoped cache hit：261.3 ns/op、80 B、1 allocation；
- direct/proxy packet read：0 B/0 allocs 与 24 B/1 alloc；
- fixed-port reload：3.59 ms median、3.59 ms p95；
- 1/2/10/50 Boxes 的 40 个 fresh-process samples，Close 后 goroutine 和 FD 全部精确
  回到启动前；heap 不宣称精确归零。

## 9. 用户可见变化

与本轮之前相比：

- standard flavor 保持原 Gust 数据面和依赖边界；
- singbox flavor 可用 `-F protocol+singbox://...` 直接嵌入 native outbound/endpoint，
  可在 GOST chain 首、中、尾组合；
- 可用 `-L protocol+singbox://...` 启动 native inbound，并把默认出站交回 GOST chain；
- 完整配置必须显式选择 initial outbound/endpoint/inbound；额外资源或 inbound 需要
  `activate_resources` / `activate_inbounds`；
- Reality、QUIC、native DNS、detour、selector/urltest 等只有在其完整依赖图能静态
  证明安全时才启动；
- 配置错误、认证错误、路由失败、DNS 失败和 prefix 丢失都 fail closed，不会偷走系统
  direct；
- `-singboxcheck` 可在不启动 runtime 的情况下输出结构摘要；`-O` 仍是 secret-bearing；
- reload 不会先发布候选再发现 listener 失败；active outbound connection 由 lease 排空；
- 某些 native registry type 虽存在，embedded 策略仍会因 IPC、全局状态或无法关闭而
  拒绝，这属于明确能力边界。

## 10. 没有合并的探索及其价值

曾构建约 161 文件的 broad master transaction 实验，用于红队验证 listener activation、
quota persistence、deferred probe、router route ownership、service retention、CRUD 和
generation visibility。它揭示了大量真实问题，但作为整体高度耦合且仍含未闭环契约，
所以没有推入正式 master。

保留的离线 topic bundles 包括 listener activation、quota activation、deferred probes、
transaction probe activation、service retention、router transaction 和 service lifecycle。
它们的定位是：

- 保存反例、失败方向、测试和可恢复代码；
- 不自动成为当前产品功能；
- 将来只能从届时最新 master 选择单个 topic 重放、重新审 API/diff 和完整验证；
- 绝不能 merge/replay consolidated stack wholesale。

所以探索没有全浪费：成熟结论进入了小型 master 栈和 sing-box fail-closed 边界；尚未
成熟的代码被隔离为可查证研究材料，而不是把风险转嫁给产品分支。

## 11. 未来追平上游

### GOST / gust-x

1. 在隔离 candidate worktree 获取新 upstream；不要直接改写已发布分支。
2. 按永久 master promotion ledger 重建/重放前置 bounded fork topics，或确认新基线已经
   包含 `gust-x 97079c1` 与 `Gust 4d4930e` 的行为等价历史；本节的 6+4 不是从原始
   upstream 重建完整 fork 的全部提交。
3. 再在 gust-x 保留本轮 6 个 bounded commits：UDP test fixture、core pin、Config
   snapshots、matcher refs、loader transaction、API delegation。
4. 在 Gust 保留本轮 4 个 commits：core pin、process runtime transaction、paired
   Mihomo CI、policy target。
5. range-diff 并跑 ordinary/race/vet/branch policy，以及 matching Gust/gust-x 的完整
   master go-ci（含 Docker E2E 与 Mihomo UoT）；审核通过后才更新 permanent master。
6. 先 forward-merge gust-x master 到 gust-x singbox-backend，再 forward-merge Gust
   master 到 Gust singbox-backend；绝不反向合并。

### sing / sing-box / sing-vmess

三个模块的 `replace` 可独立升级或解除，但 Gust 与 gust-x 必须配对操作：

1. 在干净 detached worktree 选择 Go tool 可获取的官方 module version/ref；
2. 运行行为等价测试，而不是要求 commit SHA 相同；
3. 单个模块通过后，在 Gust 与 gust-x 两边删除该模块的 `replace`，分别
   `go mod tidy`，用 `go list -m` 确认 Replace 为空，并检查双方 `go.sum` 不再引用对应
   `github.com/lovitus/...`；其他两条可继续保留，但仍要复核完整模块图；
4. 删除对应 CI `GONOSUMDB` 例外。若升级 sing-box，同步 `PinnedVersion`、generated
   registry manifest、版本说明和 Gust 的精确 gust-x pin；
5. 重跑依赖 ordinary/race/vet、paired bridge full-tags/race/vet、19-job CI 和
   pinned/latest compatibility；
6. gust-x singbox revision 或核心依赖变化必须重建 fixed-runner baseline，不能只改
   revision 字符串或放宽门槛。

## 12. 恢复、证据和操作规范

维护者私有、checksum-covered 离线恢复包以 `CURRENT-STATE.md` 为最高优先级入口；
最终 Gust/gust-x master 与
singbox-backend bundles、依赖 seed/fix bundles、format patches、重放脚本和 checksum
manifest 均已验证。公开仓库不保存私有 runner 地址或原始日志。

重型 Linux/Docker/netns/E2E 验证顺序为 private runner C、A/B、GitHub Actions；大陆
runner 使用镜像/代理并有界 fetch，不把网络失败误判为代码失败。GitHub workflow 可并发
运行。macOS 长任务必须把 `TMPDIR`、`GOCACHE`、`GOMODCACHE`、`GOTMPDIR` 放在外置
项目盘，不再填充系统 `/tmp`。

最终判断：项目增加了明确的维护成本，但已经被压缩成“一个扩展分支、三个窄依赖
replace、一套 bounded master reload 栈和可执行验证证据”。它不是零 fork，也不是
不可解耦的 sing-box 魔改；只要遵守单向分支、逐模块等价替换和 fixed-runner 重建规则，
未来可以持续追平三个上游。
