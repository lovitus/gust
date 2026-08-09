package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/go-gost/gost/internal/routemanager"
)

type controller struct {
	app           fyne.App
	window        fyne.Window
	configPath    string
	gostPath      string
	config        routemanager.Config
	processes     *routemanager.ProcessManager
	running       map[string]bool
	desired       map[string]bool
	statuses      map[string]string
	restarts      map[string]int
	restartTimers map[string]*time.Timer
	orphans       []routemanager.OrphanProcess
	orphanErr     error
	orphanBusy    bool
	watchdogStop  chan struct{}
	content       *fyne.Container
	loadErr       error
	binErr        error
}

func newController(a fyne.App, configPath, explicitGost string) *controller {
	cfg, loadErr := routemanager.Load(configPath)
	bin, binErr := routemanager.FindGost(explicitGost)
	c := &controller{
		app: a, configPath: configPath, gostPath: bin, config: cfg,
		processes: routemanager.NewProcessManager(bin), running: map[string]bool{}, desired: map[string]bool{},
		statuses: map[string]string{}, restarts: map[string]int{}, restartTimers: map[string]*time.Timer{},
		watchdogStop: make(chan struct{}), loadErr: loadErr, binErr: binErr,
	}
	c.window = a.NewWindow("自定义路由管理工具（类似 tun2socks）")
	c.window.Resize(fyne.NewSize(1200, 520))
	c.window.SetCloseIntercept(c.window.Hide)
	c.orphans, c.orphanErr = routemanager.ScanOrphanProcesses(c.gostPath, c.processes.OwnedPIDs())
	return c
}

func (c *controller) show() {
	c.render()
	c.setupTray()
	c.startWatchdog()
	c.window.Show()
	if c.loadErr != nil {
		dialog.ShowError(c.loadErr, c.window)
	}
}

func (c *controller) setupTray() {
	desktopApp, ok := c.app.(desktop.App)
	if !ok {
		return
	}
	show := fyne.NewMenuItem("显示主窗口", func() { c.window.Show(); c.window.RequestFocus() })
	stop := fyne.NewMenuItem("停止所有隧道", func() {
		go func() {
			if err := c.stopAll(); err != nil {
				fyne.Do(func() { dialog.ShowError(err, c.window) })
			}
		}()
	})
	quit := fyne.NewMenuItem("退出", func() {
		go func() {
			if err := c.stopAll(); err != nil {
				fyne.Do(func() { dialog.ShowError(fmt.Errorf("无法安全退出: %w", err), c.window) })
				return
			}
			fyne.Do(c.app.Quit)
		}()
	})
	desktopApp.SetSystemTrayMenu(fyne.NewMenu("Gust 路由管理", show, stop, fyne.NewMenuItemSeparator(), quit))
	desktopApp.SetSystemTrayIcon(theme.ComputerIcon())
	desktopApp.SetSystemTrayWindow(c.window)
}

func (c *controller) restoreWindow() {
	c.window.Show()
	c.window.RequestFocus()
}

func (c *controller) render() {
	elevated := isElevated()
	title := canvas.NewText("Gust 自定义路由管理", theme.Color(theme.ColorNameForeground))
	title.TextSize = 22
	title.TextStyle = fyne.TextStyle{Bold: true}
	subtitle := widget.NewLabel("轻量隧道与系统路由控制 · 类似 tun2socks")
	brand := container.NewVBox(title, subtitle)
	elevate := widget.NewButtonWithIcon("提权", theme.LoginIcon(), c.requestElevation)
	elevate.Importance = widget.HighImportance
	if elevated {
		elevate.Disable()
	}
	add := widget.NewButtonWithIcon("新增记录", theme.ContentAddIcon(), c.addTunnel)
	add.Importance = widget.HighImportance
	logs := widget.NewButtonWithIcon("全部日志", theme.FileTextIcon(), c.showAllLogs)
	orphans := widget.NewButtonWithIcon(fmt.Sprintf("孤儿服务：%d 个", len(c.orphans)), theme.DeleteIcon(), c.cleanupOrphans)
	if len(c.orphans) > 0 {
		orphans.Importance = widget.DangerImportance
	} else {
		orphans.Importance = widget.LowImportance
	}
	if c.orphanBusy {
		orphans.SetText("正在清理…")
		orphans.Disable()
	}
	header := container.NewPadded(container.NewBorder(nil, nil, brand, container.NewHBox(container.NewCenter(privilegeBadge(elevated)), elevate, orphans, logs, add)))

	rows := container.NewVBox(c.tableHeader())
	if len(c.config.Tunnels) == 0 {
		rows.Add(widget.NewCard("暂无记录", "点击右上角“新增记录”创建第一条配置", widget.NewLabel("新增记录不会预填内容，灰色文字仅作为输入示例。")))
	}
	for i := range c.config.Tunnels {
		rows.Add(widget.NewCard("", "", c.tunnelRow(i)))
	}
	c.content = container.NewBorder(container.NewVBox(header, widget.NewSeparator()), c.footer(), nil, nil, container.NewVScroll(container.NewPadded(rows)))
	c.window.SetContent(c.content)
}

func privilegeBadge(elevated bool) fyne.CanvasObject {
	text := "● 高权状态：无权限"
	foreground := color.NRGBA{R: 180, G: 35, B: 24, A: 255}
	background := color.NRGBA{R: 254, G: 236, B: 234, A: 255}
	if elevated {
		text = "● 高权状态：高权限"
		foreground = color.NRGBA{R: 19, G: 119, B: 72, A: 255}
		background = color.NRGBA{R: 230, G: 248, B: 237, A: 255}
	}
	box := canvas.NewRectangle(background)
	box.StrokeColor = foreground
	box.StrokeWidth = 1
	label := canvas.NewText(text, foreground)
	label.TextSize = 13
	label.TextStyle = fyne.TextStyle{Bold: true}
	label.Alignment = fyne.TextAlignCenter
	return container.NewGridWrap(fyne.NewSize(156, 36), container.NewStack(box, container.NewCenter(label)))
}

func (c *controller) tableHeader() fyne.CanvasObject {
	return container.NewHBox(
		fixed(140, widget.NewLabelWithStyle("记录名字", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})),
		fixed(90, widget.NewLabelWithStyle("状态", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})),
		fixed(590, container.NewHBox(
			fixed(110, widget.NewLabelWithStyle("模式", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})),
			fixed(290, widget.NewLabelWithStyle("路由条目（可含 dns= / mtu=）", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})),
			widget.NewLabelWithStyle("SOCKS 或 -F 链 / 自由参数", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		)),
		widget.NewLabelWithStyle("操作", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)
}

func fixed(width float32, object fyne.CanvasObject) fyne.CanvasObject {
	// Fyne entries and buttons need more than 40 px on current desktop themes.
	// A shared safe height keeps labels, inputs and buttons on one visual baseline.
	return container.NewGridWrap(fyne.NewSize(width, 52), object)
}

func (c *controller) tunnelRow(index int) fyne.CanvasObject {
	t := &c.config.Tunnels[index]
	name := widget.NewEntry()
	name.SetText(t.Name)
	name.SetPlaceHolder("例如：zwy")
	name.OnChanged = func(value string) { t.Name = value }
	parameters := c.tunnelParameters(t)
	status := c.statuses[t.ID]
	if status == "" {
		status = "已停止"
	}
	runText := "运行"
	if c.desired[t.ID] {
		runText = "停止"
	}
	run := widget.NewButton(runText, func() {
		if c.desired[t.ID] {
			c.stopTunnel(t.ID)
			return
		}
		c.startTunnel(*t, true)
	})
	save := widget.NewButton("保存", func() {
		if _, err := routemanager.BuildArgs(*t); err != nil {
			dialog.ShowError(err, c.window)
			return
		}
		if err := c.save(); err != nil {
			dialog.ShowError(err, c.window)
			return
		}
		c.statuses[t.ID] = "已保存"
		c.render()
	})
	remove := widget.NewButton("删除", func() {
		dialog.NewConfirm("删除记录", fmt.Sprintf("确定删除 %q？", t.Name), func(ok bool) {
			if !ok {
				return
			}
			_ = c.processes.Stop(t.ID)
			c.deleteTunnel(t.ID)
		}, c.window).Show()
	})
	remove.Importance = widget.DangerImportance
	viewLogs := widget.NewButton("日志", func() { c.showTunnelLogs(t.ID, t.Name) })
	return container.NewHBox(
		fixed(140, name), fixed(90, tunnelStatusBadge(status)), fixed(590, parameters),
		container.NewHBox(run, save, viewLogs, remove),
	)
}

const (
	modeRouteLabel = "路由管理"
	modeFreeLabel  = "自由参数"
)

func (c *controller) tunnelParameters(t *routemanager.Tunnel) fyne.CanvasObject {
	mode := widget.NewSelect([]string{modeRouteLabel, modeFreeLabel}, nil)
	if t.Mode == routemanager.TunnelModeFree {
		mode.SetSelected(modeFreeLabel)
	} else {
		mode.SetSelected(modeRouteLabel)
	}
	mode.OnChanged = func(value string) {
		selected := ""
		if value == modeFreeLabel {
			selected = routemanager.TunnelModeFree
		}
		if t.Mode == selected {
			return
		}
		t.Mode = selected
		c.render()
	}

	var editor fyne.CanvasObject
	if t.Mode == routemanager.TunnelModeFree {
		args := widget.NewEntry()
		args.SetText(t.Args)
		args.SetPlaceHolder("-L ... -L ... -F ... -F ...（不要输入 gost）")
		args.OnChanged = func(value string) { t.Args = value }
		editor = args
	} else {
		routes := widget.NewEntry()
		routes.SetText(t.Routes)
		routes.SetPlaceHolder("10.0.0.0/8,dns=1.1.1.1,mtu=1420")
		routes.OnChanged = func(value string) { t.Routes = value }
		target := widget.NewEntry()
		target.SetText(t.Target)
		target.SetPlaceHolder("host:port 或 -F ... -F ...")
		target.OnChanged = func(value string) { t.Target = value }
		editor = container.NewBorder(nil, nil, fixed(290, routes), nil, target)
	}
	return container.NewBorder(nil, nil, fixed(110, mode), nil, editor)
}

func tunnelStatusBadge(status string) fyne.CanvasObject {
	foreground := color.NRGBA{R: 73, G: 80, B: 87, A: 255}
	background := color.NRGBA{R: 239, G: 241, B: 243, A: 255}
	switch {
	case status == "运行中":
		foreground = color.NRGBA{R: 19, G: 119, B: 72, A: 255}
		background = color.NRGBA{R: 230, G: 248, B: 237, A: 255}
	case status == "错误" || status == "配置错误" || status == "程序缺失":
		foreground = color.NRGBA{R: 180, G: 35, B: 24, A: 255}
		background = color.NRGBA{R: 254, G: 236, B: 234, A: 255}
	case status == "启动中" || status == "停止中" || strings.Contains(status, "重启"):
		foreground = color.NRGBA{R: 181, G: 71, B: 8, A: 255}
		background = color.NRGBA{R: 255, G: 244, B: 229, A: 255}
	}
	box := canvas.NewRectangle(background)
	label := canvas.NewText(status, foreground)
	label.TextSize = 13
	label.TextStyle = fyne.TextStyle{Bold: true}
	label.Alignment = fyne.TextAlignCenter
	return container.NewStack(box, container.NewCenter(label))
}

func (c *controller) footer() fyne.CanvasObject {
	path := widget.NewLabel("配置: " + c.configPath)
	path.TextStyle = fyne.TextStyle{Monospace: true}
	return container.NewVBox(widget.NewSeparator(), path)
}

func (c *controller) addTunnel() {
	id := fmt.Sprintf("tunnel-%d", time.Now().UnixNano())
	c.config.Tunnels = append(c.config.Tunnels, newTunnel(id))
	c.render()
}

func newTunnel(id string) routemanager.Tunnel {
	return routemanager.Tunnel{ID: id}
}

func (c *controller) deleteTunnel(id string) {
	for i := range c.config.Tunnels {
		if c.config.Tunnels[i].ID == id {
			c.config.Tunnels = append(c.config.Tunnels[:i], c.config.Tunnels[i+1:]...)
			break
		}
	}
	delete(c.running, id)
	delete(c.desired, id)
	delete(c.statuses, id)
	delete(c.restarts, id)
	c.cancelRestart(id)
	if err := c.save(); err != nil {
		dialog.ShowError(err, c.window)
	}
	c.render()
}

func (c *controller) save() error {
	return routemanager.Save(c.configPath, c.config)
}

func (c *controller) startTunnel(t routemanager.Tunnel, notify bool) {
	if !isElevated() {
		if notify {
			dialog.NewConfirm("需要提权", "创建 TUN 设备和系统路由需要管理员/root 权限。现在提权吗？", func(ok bool) {
				if ok {
					c.requestElevation()
				}
			}, c.window).Show()
		} else {
			c.rejectGuardedStart(t.ID, "无权限", nil)
		}
		return
	}
	if c.binErr != nil {
		if notify {
			dialog.ShowError(c.binErr, c.window)
		} else {
			c.rejectGuardedStart(t.ID, "程序缺失", c.binErr)
		}
		return
	}
	if _, err := routemanager.BuildArgs(t); err != nil {
		if notify {
			dialog.ShowError(err, c.window)
		} else {
			c.rejectGuardedStart(t.ID, "配置错误", err)
		}
		return
	}
	if err := c.save(); err != nil {
		if notify {
			dialog.ShowError(err, c.window)
		} else {
			c.rejectGuardedStart(t.ID, "配置错误", err)
		}
		return
	}
	if notify {
		// An explicit user action closes the previous circuit breaker and starts
		// a fresh retry sequence. Watchdog restarts pass notify=false.
		c.restarts[t.ID] = 0
	}
	if command, err := routemanager.CommandPreview(c.gostPath, t); err == nil {
		c.processes.AppendLog(t.ID, "即将执行: "+command)
	}
	c.desired[t.ID] = true
	c.statuses[t.ID] = "启动中"
	c.render()
	startedAt := time.Now()
	err := c.processes.Start(t, func(err error) {
		fyne.Do(func() {
			c.running[t.ID] = false
			if !c.desired[t.ID] {
				c.statuses[t.ID] = "已停止"
				c.render()
				return
			}
			if time.Since(startedAt) >= 30*time.Second {
				c.restarts[t.ID] = 0
			}
			summary := tunnelFailureSummary(c.processes.Logs(t.ID, 12), err)
			if err != nil && isPortBindingFailure(summary) {
				c.stopRestartingAfterFailure(t, "端口被占用", summary)
				return
			}
			c.restarts[t.ID]++
			if err != nil && c.restarts[t.ID] >= 5 {
				c.stopRestartingAfterFailure(t, "反复启动失败", summary)
				return
			}
			delay := restartBackoff(c.restarts[t.ID])
			c.statuses[t.ID] = fmt.Sprintf("%s 后重启", delay)
			if err != nil && !strings.Contains(err.Error(), "signal") && notify {
				dialog.ShowError(fmt.Errorf("隧道 %s 启动失败，将在 %s 后重试：%s", t.Name, delay, summary), c.window)
			}
			c.scheduleRestart(t.ID, delay)
			c.render()
		})
	})
	if err != nil {
		c.restarts[t.ID]++
		if c.restarts[t.ID] >= 5 {
			c.stopRestartingAfterFailure(t, "反复启动失败", err.Error())
			return
		}
		delay := restartBackoff(c.restarts[t.ID])
		c.statuses[t.ID] = fmt.Sprintf("%s 后重启", delay)
		c.scheduleRestart(t.ID, delay)
		c.render()
		if notify {
			dialog.ShowError(err, c.window)
		}
		return
	}
	c.running[t.ID] = true
	c.statuses[t.ID] = "运行中"
	c.render()
}

func (c *controller) stopRestartingAfterFailure(t routemanager.Tunnel, status, summary string) {
	c.desired[t.ID] = false
	c.cancelRestart(t.ID)
	c.statuses[t.ID] = status + "（守护已停止）"
	c.processes.AppendLog(t.ID, "守护已停止: "+summary)
	c.render()
	dialog.ShowError(fmt.Errorf("隧道 %s：%s。请修正后重新运行。", t.Name, summary), c.window)
}

func tunnelFailureSummary(lines []routemanager.LogLine, fallback error) string {
	for i := len(lines) - 1; i >= 0; i-- {
		text := strings.TrimSpace(lines[i].Text)
		if text == "" || strings.HasPrefix(text, "[管理器]") {
			continue
		}
		var payload struct {
			Message string `json:"msg"`
		}
		if json.Unmarshal([]byte(text), &payload) == nil && strings.TrimSpace(payload.Message) != "" {
			return truncateMessage(strings.TrimSpace(payload.Message), 600)
		}
		return truncateMessage(text, 600)
	}
	if fallback != nil {
		return truncateMessage(fallback.Error(), 600)
	}
	return "进程已退出"
}

func truncateMessage(text string, limit int) string {
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	return string(runes[:limit-1]) + "…"
}

func isPortBindingFailure(message string) bool {
	message = strings.ToLower(message)
	return strings.Contains(message, "address already in use") ||
		strings.Contains(message, "only one usage of each socket address") ||
		strings.Contains(message, "address/port is normally permitted")
}

func (c *controller) rejectGuardedStart(id, status string, err error) {
	c.desired[id] = false
	c.cancelRestart(id)
	c.statuses[id] = status
	message := "守护已停止: " + status
	if err != nil {
		message += ": " + err.Error()
	}
	c.processes.AppendLog(id, message)
	c.render()
}

func restartBackoff(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	if attempt > 5 {
		attempt = 5
	}
	return time.Duration(1<<uint(attempt-1)) * time.Second
}

func (c *controller) stopTunnel(id string) {
	c.desired[id] = false
	c.cancelRestart(id)
	c.statuses[id] = "停止中"
	c.render()
	go func() {
		if err := c.processes.Stop(id); err != nil {
			fyne.Do(func() { dialog.ShowError(err, c.window) })
			return
		}
		fyne.Do(func() {
			c.running[id] = false
			c.statuses[id] = "已停止"
			c.render()
		})
	}()
}

func (c *controller) scheduleRestart(id string, delay time.Duration) {
	c.cancelRestart(id)
	c.restartTimers[id] = time.AfterFunc(delay, func() {
		fyne.Do(func() {
			delete(c.restartTimers, id)
			if !c.desired[id] || c.processes.Running(id) {
				return
			}
			tunnel, ok := c.tunnelByID(id)
			if !ok {
				return
			}
			c.startTunnel(tunnel, false)
		})
	})
}

func (c *controller) cancelRestart(id string) {
	if timer := c.restartTimers[id]; timer != nil {
		timer.Stop()
		delete(c.restartTimers, id)
	}
}

func (c *controller) tunnelByID(id string) (routemanager.Tunnel, bool) {
	for _, tunnel := range c.config.Tunnels {
		if tunnel.ID == id {
			return tunnel, true
		}
	}
	return routemanager.Tunnel{}, false
}

func (c *controller) startWatchdog() {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				fyne.Do(func() {
					for id, desired := range c.desired {
						if desired && !c.processes.Running(id) && c.restartTimers[id] == nil {
							c.statuses[id] = "守护重启中"
							c.scheduleRestart(id, 0)
							c.render()
						}
					}
				})
			case <-c.watchdogStop:
				return
			}
		}
	}()
}

func (c *controller) stopAll() error {
	fyne.DoAndWait(func() {
		for id := range c.desired {
			c.desired[id] = false
			c.cancelRestart(id)
		}
	})
	err := c.processes.StopAll()
	fyne.Do(func() {
		for id := range c.running {
			if c.processes.Running(id) {
				c.statuses[id] = "停止失败"
				continue
			}
			c.running[id] = false
			c.statuses[id] = "已停止"
		}
		c.render()
	})
	return err
}

func (c *controller) cleanupOrphans() {
	if c.orphanBusy {
		return
	}
	if c.orphanErr != nil {
		dialog.ShowError(c.orphanErr, c.window)
		return
	}
	if len(c.orphans) == 0 {
		c.refreshOrphans(true)
		return
	}
	if !isElevated() {
		dialog.ShowInformation("需要提权", "清理孤儿服务需要高权限，即将请求系统授权。", c.window)
		c.requestElevation()
		return
	}
	targets := append([]routemanager.OrphanProcess(nil), c.orphans...)
	content := previewContent(
		fmt.Sprintf("将停止并清理 %d 个故障遗留的 gost-qt 服务。不会影响名为 gost 的其他进程。", len(targets)),
		formatOrphanPreview(targets),
		fyne.NewSize(780, 280),
	)
	confirm := dialog.NewCustomConfirm(
		"清理孤儿服务",
		"清理",
		"取消",
		content,
		func(ok bool) {
			if !ok {
				return
			}
			c.orphanBusy = true
			c.render()
			go func() {
				cleanupErr := routemanager.CleanupOrphanProcesses(targets)
				remaining, scanErr := routemanager.ScanOrphanProcesses(c.gostPath, c.processes.OwnedPIDs())
				fyne.Do(func() {
					c.orphanBusy = false
					c.orphans = remaining
					c.orphanErr = scanErr
					c.render()
					if cleanupErr != nil {
						dialog.ShowError(cleanupErr, c.window)
						return
					}
					if scanErr != nil {
						dialog.ShowError(scanErr, c.window)
						return
					}
					dialog.ShowInformation("清理完成", fmt.Sprintf("已处理 %d 个孤儿服务，剩余 %d 个。", len(targets), len(remaining)), c.window)
				})
			}()
		},
		c.window,
	)
	confirm.SetConfirmImportance(widget.DangerImportance)
	confirm.Show()
}

func previewContent(description, detail string, size fyne.Size) fyne.CanvasObject {
	header := widget.NewLabel(description)
	header.Wrapping = fyne.TextWrapWord
	preview := widget.NewLabel(detail)
	preview.Wrapping = fyne.TextWrapBreak
	preview.TextStyle = fyne.TextStyle{Monospace: true}
	scroll := container.NewVScroll(preview)
	return container.NewVBox(header, container.NewGridWrap(size, scroll))
}

func formatOrphanPreview(targets []routemanager.OrphanProcess) string {
	var out strings.Builder
	for i, target := range targets {
		if i > 0 {
			out.WriteString("\n\n")
		}
		executable := strings.TrimSpace(target.Executable)
		if executable == "" {
			executable = "gost-qt（路径无法读取）"
		}
		command := strings.TrimSpace(target.CommandLine)
		if command == "" {
			command = executable
		}
		cleanupAction := strings.TrimSpace(target.CleanupAction)
		if cleanupAction == "" {
			cleanupAction = "终止已验证的 gost-qt 进程"
		}
		fmt.Fprintf(&out, "PID: %d\n启动时间: %s\n执行文件: %s\n原启动命令: %s\n清理动作: %s",
			target.PID,
			time.UnixMilli(target.StartedAt).Format("2006-01-02 15:04:05"),
			executable,
			command,
			cleanupAction,
		)
	}
	return out.String()
}

func (c *controller) refreshOrphans(notify bool) {
	c.orphanBusy = true
	c.render()
	go func() {
		orphans, err := routemanager.ScanOrphanProcesses(c.gostPath, c.processes.OwnedPIDs())
		fyne.Do(func() {
			c.orphanBusy = false
			c.orphans = orphans
			c.orphanErr = err
			c.render()
			if err != nil {
				dialog.ShowError(err, c.window)
			} else if notify {
				dialog.ShowInformation("扫描完成", fmt.Sprintf("发现 %d 个孤儿服务。", len(orphans)), c.window)
			}
		})
	}()
}

func (c *controller) shutdown() error {
	select {
	case <-c.watchdogStop:
	default:
		close(c.watchdogStop)
	}
	for id := range c.restartTimers {
		c.cancelRestart(id)
	}
	for id := range c.desired {
		c.desired[id] = false
	}
	return c.processes.StopAll()
}

func (c *controller) showTunnelLogs(id, name string) {
	if strings.TrimSpace(name) == "" {
		name = id
	}
	c.showLogs(
		"任务日志 · "+name,
		func() uint64 { return c.processes.LogVersion(id) },
		func() (string, uint64) {
			lines, version := c.processes.LogsSnapshot(id, 100)
			return formatLogs(lines, nil), version
		},
	)
}

func (c *controller) showAllLogs() {
	names := make(map[string]string, len(c.config.Tunnels))
	for _, tunnel := range c.config.Tunnels {
		names[tunnel.ID] = tunnel.Name
	}
	c.showLogs(
		"全部任务日志 · 最近 1000 行",
		c.processes.AllLogVersion,
		func() (string, uint64) {
			lines, version := c.processes.AllLogsSnapshot(1000)
			return formatLogs(lines, names), version
		},
	)
}

func (c *controller) showLogs(title string, version func() uint64, load func() (string, uint64)) {
	view := widget.NewMultiLineEntry()
	view.Wrapping = fyne.TextWrapOff
	current := ""
	var currentVersion atomic.Uint64
	updating := false
	setText := func() {
		text, loadedVersion := load()
		updating = true
		current = text
		currentVersion.Store(loadedVersion)
		view.SetText(current)
		updating = false
	}
	view.OnChanged = func(value string) {
		if updating || value == current {
			return
		}
		updating = true
		view.SetText(current)
		updating = false
	}
	setText()
	content := container.NewGridWrap(fyne.NewSize(900, 460), view)
	d := dialog.NewCustomWithoutButtons(title, content, c.window)
	refresh := widget.NewButtonWithIcon("刷新", theme.ViewRefreshIcon(), setText)
	closeButton := widget.NewButton("关闭", d.Hide)
	d.SetButtons([]fyne.CanvasObject{refresh, closeButton})
	done := make(chan struct{})
	var stopOnce sync.Once
	d.SetOnClosed(func() { stopOnce.Do(func() { close(done) }) })
	d.Show()
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if version() == currentVersion.Load() {
					continue
				}
				fyne.Do(func() {
					select {
					case <-done:
						return
					default:
					}
					if version() != currentVersion.Load() {
						setText()
					}
				})
			case <-done:
				return
			}
		}
	}()
}

func formatLogs(lines []routemanager.LogLine, names map[string]string) string {
	if len(lines) == 0 {
		return "暂无日志"
	}
	var out strings.Builder
	for _, line := range lines {
		fmt.Fprintf(&out, "[%s]", line.Time.Format("15:04:05"))
		if names != nil {
			name := strings.TrimSpace(names[line.TunnelID])
			if name == "" {
				name = line.TunnelID
			}
			fmt.Fprintf(&out, " [%s]", name)
		}
		fmt.Fprintf(&out, " %s\n", line.Text)
	}
	return strings.TrimSuffix(out.String(), "\n")
}

func (c *controller) requestElevation() {
	if isElevated() {
		return
	}
	if err := c.save(); err != nil {
		dialog.ShowError(err, c.window)
		return
	}
	ready := filepath.Join(os.TempDir(), fmt.Sprintf("gust-route-manager-elevated-%d", time.Now().UnixNano()))
	args := []string{"--config", c.configPath, "--elevation-ready", ready}
	if c.gostPath != "" {
		args = append(args, "--gost", c.gostPath)
	}
	if err := elevateSelf(args); err != nil {
		dialog.ShowError(err, c.window)
		return
	}
	go func() {
		deadline := time.Now().Add(45 * time.Second)
		for time.Now().Before(deadline) {
			if _, err := os.Stat(ready); err == nil {
				_ = os.Remove(ready)
				fyne.Do(c.app.Quit)
				return
			}
			time.Sleep(250 * time.Millisecond)
		}
		fyne.Do(func() {
			dialog.ShowInformation("提权未完成", "未检测到高权限窗口，当前窗口将继续保留。", c.window)
		})
	}()
}
