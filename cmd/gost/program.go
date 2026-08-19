package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-gost/core/auth"
	"github.com/go-gost/core/logger"
	"github.com/go-gost/core/service"
	api_service "github.com/go-gost/x/api/service"
	xauth "github.com/go-gost/x/auth"
	"github.com/go-gost/x/config"
	"github.com/go-gost/x/config/loader"
	auth_parser "github.com/go-gost/x/config/parsing/auth"
	"github.com/go-gost/x/config/parsing/parser"
	metrics "github.com/go-gost/x/metrics/service"
	"github.com/go-gost/x/registry"
	"github.com/judwhite/go-svc"
)

type program struct {
	srvApi       service.Service
	srvMetrics   service.Service
	srvProfiling service.Service

	cancel context.CancelFunc
}

func (p *program) Init(env svc.Environment) error {
	parser.Init(parser.Args{
		CfgFiles:    cfgFiles,
		Services:    services,
		Nodes:       nodes,
		Debug:       debug,
		Trace:       trace,
		ApiAddr:     apiAddr,
		MetricsAddr: metricsAddr,
	})

	return nil
}

func (p *program) Start() error {
	cfg, err := parser.Parse()
	if err != nil {
		return err
	}

	if outputFormat != "" {
		if err := cfg.Write(os.Stdout, outputFormat); err != nil {
			return err
		}
		os.Exit(0)
	}

	if err := p.activateConfig(cfg); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel
	go p.reload(ctx)

	return nil
}

func (p *program) Stop() error {
	if p.cancel != nil {
		p.cancel()
	}

	return loader.WithLock(func() error {
		for name := range registry.ServiceRegistry().GetAll() {
			registry.ServiceRegistry().Unregister(name)
			logger.Default().Debugf("service %s shutdown", name)
		}
		return p.deactivateRuntime()
	})
}

func (p *program) reload(ctx context.Context) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)

	var ticker <-chan time.Time
	if reload > 0 {
		t := time.NewTicker(reload)
		defer t.Stop()
		ticker = t.C
	}

	for {
		select {
		case <-c:
			if err := p.reloadConfig(); err != nil {
				logger.Default().Error(err)
			} else {
				logger.Default().Info("config reloaded")
			}

		case <-ticker:
			if err := p.reloadConfig(); err != nil {
				logger.Default().Errorf("auto reload: %v", err)
			} else {
				logger.Default().Debug("config auto reloaded")
			}

		case <-ctx.Done():
			return
		}
	}
}

func (p *program) reloadConfig() error {
	return loader.WithLock(func() error {
		cfg, err := parser.Parse()
		if err != nil {
			return err
		}
		return p.activateConfigLocked(cfg)
	})
}

func buildApiService(cfg *config.APIConfig, reload func() error) (service.Service, error) {
	var authers []auth.Authenticator
	if auther := auth_parser.ParseAutherFromAuth(cfg.Auth); auther != nil {
		authers = append(authers, auther)
	}
	if cfg.Auther != "" {
		authers = append(authers, registry.AutherRegistry().Get(cfg.Auther))
	}

	var auther auth.Authenticator
	if len(authers) > 0 {
		auther = xauth.AuthenticatorGroup(authers...)
	}

	network := "tcp"
	addr := cfg.Addr
	if strings.HasPrefix(addr, "unix://") {
		network = "unix"
		addr = strings.TrimPrefix(addr, "unix://")
	}
	return api_service.NewService(
		network, addr,
		api_service.PathPrefixOption(cfg.PathPrefix),
		api_service.AccessLogOption(cfg.AccessLog),
		api_service.AutherOption(auther),
		api_service.ReloadOption(reload),
	)
}

func buildMetricsService(cfg *config.MetricsConfig) (service.Service, error) {
	auther := auth_parser.ParseAutherFromAuth(cfg.Auth)
	if cfg.Auther != "" {
		auther = registry.AutherRegistry().Get(cfg.Auther)
	}

	network := "tcp"
	addr := cfg.Addr
	if strings.HasPrefix(addr, "unix://") {
		network = "unix"
		addr = strings.TrimPrefix(addr, "unix://")
	}
	return metrics.NewService(
		network, addr,
		metrics.PathOption(cfg.Path),
		metrics.AutherOption(auther),
	)
}
