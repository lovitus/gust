package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

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
	if *readyFile != "" {
		_ = os.WriteFile(*readyFile, []byte("ready\n"), 0o600)
	}

	gui := app.NewWithID("us.lovis.gust.route-manager")
	controller := newController(gui, configAbs, *gostPath)
	controller.show()
	gui.Run()
	controller.processes.StopAll()

	// Give gost a brief opportunity to tear down its TUN device and routes.
	time.Sleep(100 * time.Millisecond)
}
