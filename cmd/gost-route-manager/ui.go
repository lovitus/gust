package main

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"
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
	app        fyne.App
	window     fyne.Window
	configPath string
	gostPath   string
	config     routemanager.Config
	processes  *routemanager.ProcessManager
	running    map[string]bool
	statuses   map[string]string
	content    *fyne.Container
	loadErr    error
	binErr     error
}

func newController(a fyne.App, configPath, explicitGost string) *controller {
	cfg, loadErr := routemanager.Load(configPath)
	bin, binErr := routemanager.FindGost(explicitGost)
	c := &controller{
		app: a, configPath: configPath, gostPath: bin, config: cfg,
		processes: routemanager.NewProcessManager(bin), running: map[string]bool{},
		statuses: map[string]string{}, loadErr: loadErr, binErr: binErr,
	}
	c.window = a.NewWindow("自定义路由管理工具（类似 tun2socks）")
	c.window.Resize(fyne.NewSize(1200, 520))
	c.window.SetCloseIntercept(c.window.Hide)
	return c
}

func (c *controller) show() {
	c.render()
	c.setupTray()
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
	stop := fyne.NewMenuItem("停止所有隧道", func() { go c.stopAll() })
	quit := fyne.NewMenuItem("退出", func() {
		c.processes.StopAll()
		c.app.Quit()
	})
	desktopApp.SetSystemTrayMenu(fyne.NewMenu("Gust 路由管理", show, stop, fyne.NewMenuItemSeparator(), quit))
	desktopApp.SetSystemTrayIcon(theme.ComputerIcon())
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
	add := widget.NewButtonWithIcon("新增隧道", theme.ContentAddIcon(), c.addTunnel)
	add.Importance = widget.HighImportance
	header := container.NewPadded(container.NewBorder(nil, nil, brand, container.NewHBox(container.NewCenter(privilegeBadge(elevated)), elevate, add)))

	rows := container.NewVBox(c.tableHeader())
	if len(c.config.Tunnels) == 0 {
		rows.Add(widget.NewCard("暂无隧道", "点击右上角“新增隧道”创建第一条配置", widget.NewLabel("新增记录不会预填内容，灰色文字仅作为输入示例。")))
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
		fixed(150, widget.NewLabelWithStyle("隧道名字", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})),
		fixed(96, widget.NewLabelWithStyle("状态", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})),
		fixed(390, widget.NewLabelWithStyle("路由条目（逗号分隔，可含 dns= / mtu=）", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})),
		fixed(260, widget.NewLabelWithStyle("目标 SOCKS / 自定义 -F", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})),
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
	routes := widget.NewEntry()
	routes.SetText(t.Routes)
	routes.SetPlaceHolder("10.0.0.0/8,dns=1.1.1.1,mtu=1420")
	routes.OnChanged = func(value string) { t.Routes = value }
	target := widget.NewEntry()
	target.SetText(t.Target)
	target.SetPlaceHolder("192.168.1.37:5555 或 -F socks5://...")
	target.OnChanged = func(value string) { t.Target = value }
	status := c.statuses[t.ID]
	if status == "" {
		status = "已停止"
	}
	runText := "运行"
	if c.running[t.ID] {
		runText = "停止"
	}
	run := widget.NewButton(runText, func() {
		if c.running[t.ID] {
			c.statuses[t.ID] = "停止中"
			c.render()
			go func(id string) {
				if err := c.processes.Stop(id); err != nil {
					fyne.Do(func() { dialog.ShowError(err, c.window) })
				}
			}(t.ID)
			return
		}
		c.runTunnel(*t)
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
		dialog.NewConfirm("删除隧道", fmt.Sprintf("确定删除 %q？", t.Name), func(ok bool) {
			if !ok {
				return
			}
			_ = c.processes.Stop(t.ID)
			c.deleteTunnel(t.ID)
		}, c.window).Show()
	})
	remove.Importance = widget.DangerImportance
	return container.NewHBox(
		fixed(150, name), fixed(96, tunnelStatusBadge(status)), fixed(390, routes), fixed(260, target),
		container.NewHBox(run, save, remove),
	)
}

func tunnelStatusBadge(status string) fyne.CanvasObject {
	foreground := color.NRGBA{R: 73, G: 80, B: 87, A: 255}
	background := color.NRGBA{R: 239, G: 241, B: 243, A: 255}
	switch status {
	case "运行中":
		foreground = color.NRGBA{R: 19, G: 119, B: 72, A: 255}
		background = color.NRGBA{R: 230, G: 248, B: 237, A: 255}
	case "错误":
		foreground = color.NRGBA{R: 180, G: 35, B: 24, A: 255}
		background = color.NRGBA{R: 254, G: 236, B: 234, A: 255}
	case "启动中", "停止中":
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
	delete(c.statuses, id)
	if err := c.save(); err != nil {
		dialog.ShowError(err, c.window)
	}
	c.render()
}

func (c *controller) save() error {
	return routemanager.Save(c.configPath, c.config)
}

func (c *controller) runTunnel(t routemanager.Tunnel) {
	if !isElevated() {
		dialog.NewConfirm("需要提权", "创建 TUN 设备和系统路由需要管理员/root 权限。现在提权吗？", func(ok bool) {
			if ok {
				c.requestElevation()
			}
		}, c.window).Show()
		return
	}
	if c.binErr != nil {
		dialog.ShowError(c.binErr, c.window)
		return
	}
	if _, err := routemanager.BuildArgs(t); err != nil {
		dialog.ShowError(err, c.window)
		return
	}
	if err := c.save(); err != nil {
		dialog.ShowError(err, c.window)
		return
	}
	c.statuses[t.ID] = "启动中"
	c.render()
	err := c.processes.Start(t, func(err error) {
		fyne.Do(func() {
			c.running[t.ID] = false
			c.statuses[t.ID] = "已停止"
			if err != nil && !strings.Contains(err.Error(), "signal") {
				c.statuses[t.ID] = "错误"
				message := strings.TrimSpace(c.processes.Output(t.ID))
				if message == "" {
					message = err.Error()
				}
				dialog.ShowError(fmt.Errorf("隧道 %s 已退出: %s", t.Name, message), c.window)
			}
			c.render()
		})
	})
	if err != nil {
		c.statuses[t.ID] = "错误"
		c.render()
		dialog.ShowError(err, c.window)
		return
	}
	c.running[t.ID] = true
	c.statuses[t.ID] = "运行中"
	c.render()
}

func (c *controller) stopAll() {
	c.processes.StopAll()
	fyne.Do(func() {
		for id := range c.running {
			c.running[id] = false
			c.statuses[id] = "已停止"
		}
		c.render()
	})
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
