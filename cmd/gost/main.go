package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	_ "net/http/pprof"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/go-gost/core/logger"
	xlogger "github.com/go-gost/x/logger"
	"github.com/judwhite/go-svc"
)

type stringList []string

func (l *stringList) String() string {
	return fmt.Sprintf("%s", *l)
}
func (l *stringList) Set(value string) error {
	*l = append(*l, value)
	return nil
}

var (
	cfgFile      string
	outputFormat string
	services     stringList
	nodes        stringList
	debug        bool
	trace        bool
	apiAddr      string
	metricsAddr  string
	watchdog     bool
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)

	args := strings.Join(os.Args[1:], "  ")

	if strings.Contains(args, " -- ") {
		var (
			wg  sync.WaitGroup
			ret int
		)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		for wid, wargs := range strings.Split(" "+args+" ", " -- ") {
			wg.Add(1)
			go func(wid int, wargs string) {
				defer wg.Done()
				defer cancel()
				worker(wid, strings.Split(wargs, "  "), &ctx, &ret)
			}(wid, strings.TrimSpace(wargs))
		}

		wg.Wait()

		os.Exit(ret)
	}
}

func worker(id int, args []string, ctx *context.Context, ret *int) {
	cmd := exec.CommandContext(*ctx, os.Args[0], args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), fmt.Sprintf("_GOST_ID=%d", id))

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	if cmd.ProcessState.Exited() {
		*ret = cmd.ProcessState.ExitCode()
	}
}

func init() {
	var printVersion bool

	flag.Var(&services, "L", "service list")
	flag.Var(&nodes, "F", "chain node list")
	flag.StringVar(&cfgFile, "C", "", "configuration file")
	flag.BoolVar(&printVersion, "V", false, "print version")
	flag.StringVar(&outputFormat, "O", "", "output format, one of yaml|json format")
	flag.BoolVar(&debug, "D", false, "debug mode")
	flag.BoolVar(&trace, "DD", false, "trace mode")
	flag.StringVar(&apiAddr, "api", "", "api service address")
	flag.StringVar(&metricsAddr, "metrics", "", "metrics service address")
	flag.BoolVar(&watchdog, "watchdog", false, "enable watchdog (auto-restart on crash)")
	flag.Parse()

	if printVersion {
		fmt.Fprintf(os.Stdout, "gost %s (%s %s/%s)\n",
			version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}
}

func main() {
	if watchdog && os.Getenv("_GOST_WATCHDOG_CHILD") == "" {
		runWatchdog()
		return
	}

	log := xlogger.NewLogger()
	logger.SetDefault(log)

	p := &program{}

	if err := svc.Run(p); err != nil {
		logger.Default().Fatal(err)
	}
}

func runWatchdog() {
	log.Println("watchdog: started")

	backoff := time.Second
	maxBackoff := 60 * time.Second
	stableAfter := 5 * time.Minute

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	for {
		cmd := exec.Command(os.Args[0], os.Args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Env = append(os.Environ(), "_GOST_WATCHDOG_CHILD=1")

		start := time.Now()
		if err := cmd.Start(); err != nil {
			log.Printf("watchdog: failed to start: %v", err)
			time.Sleep(backoff)
			if backoff < maxBackoff {
				backoff *= 2
			}
			continue
		}

		childDone := make(chan error, 1)
		go func() { childDone <- cmd.Wait() }()

		select {
		case sig := <-sigCh:
			cmd.Process.Signal(sig)
			<-childDone
			os.Exit(0)
		case err := <-childDone:
			if err == nil {
				os.Exit(0)
			}

			elapsed := time.Since(start)
			if elapsed > stableAfter {
				backoff = time.Second
			}

			log.Printf("watchdog: process exited (%v), restarting in %v", err, backoff)
			time.Sleep(backoff)
			if backoff < maxBackoff {
				backoff *= 2
			}
		}
	}
}
