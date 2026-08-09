package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/go-gost/gost/internal/routemanager"
)

func main() {
	defaultConfig, err := routemanager.DefaultConfigPath()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	configPath := flag.String("config", defaultConfig, "配置文件路径")
	gostPath := flag.String("gost", "", "gost 可执行文件路径")
	readyFile := flag.String("elevation-ready", "", "提权启动握手文件")
	flag.Parse()

	configAbs, err := filepath.Abs(*configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	handoff := *readyFile != ""
	if handoff {
		// Let the ordinary process exit and release its instance lock before the
		// elevated replacement tries to take ownership.
		_ = os.WriteFile(*readyFile, []byte("ready\n"), 0o600)
	}
	lockWait := time.Duration(0)
	if handoff {
		lockWait = 20 * time.Second
	}
	instance, err := acquireSingleInstance(configAbs, lockWait)
	if errors.Is(err, errInstanceRunning) {
		if activateErr := activateExistingInstance(configAbs, 2*time.Second); activateErr != nil {
			fmt.Fprintln(os.Stderr, "Gust 路由管理工具已经在运行；", activateErr)
		}
		return
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "单实例检查失败:", err)
		os.Exit(1)
	}
	defer instance.Close()

	gui := app.NewWithID("us.lovis.gust.route-manager")
	controller := newController(gui, configAbs, *gostPath)
	controller.show()
	go func() {
		for range instance.Activations() {
			fyne.Do(controller.restoreWindow)
		}
	}()
	gui.Run()
	controller.shutdown()

	// Give gost a brief opportunity to tear down its TUN device and routes.
	time.Sleep(100 * time.Millisecond)
}
