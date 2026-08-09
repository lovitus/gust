# 自定义路由管理工具

`gost-route-manager` 是一个小型桌面前端，用 Gust 的 `tungo` 将指定网段转发到 SOCKS 或自定义 `-F` 节点。界面和托盘由 Go/Fyne 实现，不需要安装 Qt 运行库。

## 功能

- Windows、macOS 和 Linux 桌面；支持直接以管理员/root 身份运行。
- 普通用户可以查看和编辑配置。点击“提权”或尝试运行隧道时，会调用 Windows UAC、macOS 管理员授权或 Linux polkit。
- 隧道支持新增、编辑保存、删除、运行和停止；关闭窗口后仍驻留托盘。
- 同一配置只允许一个管理器实例；重复启动不会再创建窗口。普通窗口提权时会把实例锁安全交接给高权限窗口。
- 新增隧道保持空白，输入框中的示例是 placeholder，不需要先手工删除。
- 配置持久化在用户配置目录下的 `gust/route-manager.json`。提权时沿用原用户的配置文件，不会切换到 root 的配置。
- 每个隧道使用独立、稳定的 TUN 地址和接口名，可同时运行多条隧道。

## 输入格式

路由条目使用逗号分隔，至少包含一个 CIDR；还可以追加 DNS 和 MTU：

```text
10.233.0.0/16,10.27.0.0/16,10.26.0.0/26,dns=1.1.1.1,8.8.8.8,mtu=1420
```

目标只写地址时默认使用 SOCKS5：

```text
192.168.1.37:5555
```

也可以填写完整的 Gust `-F` URL，开头的 `-F` 可省略：

```text
socks5://user:password@192.168.1.37:5555
-F http2://proxy.example.com:443
```

## 构建和运行

先构建 `gost`，再构建当前平台的管理器；两个程序放在同一目录即可自动发现：

```sh
go build -o bin/gost ./cmd/gost
make route-manager
bin/gost-route-manager
```

也可以显式指定路径：

```sh
bin/gost-route-manager --gost /path/to/gost
```

Windows 原生终端中使用 `make route-manager-windows`（需要可用的 C 编译器和 Fyne/GLFW 开发环境）。最终二进制使用 `-s -w` 去除调试信息，且不携带 Qt、Electron 或浏览器运行时。

Linux 的 UI 提权依赖 `pkexec`。如果系统没有 polkit，可以直接执行 `sudo bin/gost-route-manager`；传入 `--config` 可以继续使用普通用户的配置文件。

## 安全边界

普通权限状态不会启动任何隧道。只有高权限实例会创建 TUN 设备、修改路由/DNS 并启动 `gost` 子进程。退出程序或从托盘选择“停止所有隧道”时，管理器会通知所有子进程退出。
