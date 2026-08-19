package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-gost/core/logger"
	"github.com/go-gost/core/service"
	"github.com/go-gost/x/config"
	"github.com/go-gost/x/config/loader"
	xmetrics "github.com/go-gost/x/metrics"
	"github.com/go-gost/x/registry"
)

type runtimeChange struct {
	keepAPI       bool
	keepMetrics   bool
	keepProfiling bool
}

func keepAllRuntime() runtimeChange {
	return runtimeChange{keepAPI: true, keepMetrics: true, keepProfiling: true}
}

// activateConfig serializes parse-independent activation with every loader
// commit. The published config changes only after loader-owned and
// process-owned listeners have all bound successfully.
func (p *program) activateConfig(cfg *config.Config) error {
	return loader.WithLock(func() error {
		return p.activateConfigLocked(cfg)
	})
}

func (p *program) activateConfigLocked(cfg *config.Config) error {
	if err := validateRuntimeConfig(cfg); err != nil {
		return err
	}
	prepared, err := loader.Prepare(cfg)
	if err != nil {
		return err
	}
	candidate := prepared.Config()
	previous := config.Snapshot()
	oldMetricsEnabled := xmetrics.IsEnabled()
	newMetricsEnabled := metricsWanted(candidate.Metrics)
	change := keepAllRuntime()

	err = prepared.CommitLockedWithHooks(loader.ActivationHooks{
		Apply:   func() { xmetrics.Enable(newMetricsEnabled) },
		Restore: func() { xmetrics.Enable(oldMetricsEnabled) },
	}, func() error {
		var err error
		change, err = p.commitRuntime(candidate, previous)
		if err != nil {
			return fmt.Errorf("activate process services: %w", err)
		}
		return nil
	}, func() error {
		if err := p.restoreRuntime(previous, change); err != nil {
			return err
		}
		startRegisteredServices()
		return nil
	})
	if err != nil {
		return err
	}

	startRegisteredServices()
	p.startRuntime(change)
	return nil
}

func (p *program) commitRuntime(candidate, previous *config.Config) (runtimeChange, error) {
	change := runtimeChange{
		keepAPI:       runtimeServiceUnchanged(previous.API, candidate.API, apiWanted(candidate.API), p.srvApi),
		keepMetrics:   runtimeServiceUnchanged(previous.Metrics, candidate.Metrics, metricsWanted(candidate.Metrics), p.srvMetrics),
		keepProfiling: runtimeServiceUnchanged(previous.Profiling, candidate.Profiling, profilingWanted(candidate.Profiling), p.srvProfiling),
	}

	var closeErrs []error
	if !change.keepAPI && p.srvApi != nil {
		closeErrs = append(closeErrs, wrapRuntimeClose("API", retireRuntimeService(p.srvApi)))
		p.srvApi = nil
	}
	if !change.keepMetrics && p.srvMetrics != nil {
		closeErrs = append(closeErrs, wrapRuntimeClose("metrics", p.srvMetrics.Close()))
		p.srvMetrics = nil
	}
	if !change.keepProfiling && p.srvProfiling != nil {
		closeErrs = append(closeErrs, wrapRuntimeClose("profiling", p.srvProfiling.Close()))
		p.srvProfiling = nil
	}
	if err := errors.Join(closeErrs...); err != nil {
		return change, err
	}

	api, metricsService, profiling, err := buildRuntimeServices(candidate, change, p.reloadConfig)
	if err != nil {
		closeRuntimeServices(api, metricsService, profiling)
		return change, err
	}
	if !change.keepAPI {
		p.srvApi = api
	}
	if !change.keepMetrics {
		p.srvMetrics = metricsService
	}
	if !change.keepProfiling {
		p.srvProfiling = profiling
	}
	return change, nil
}

func (p *program) startRuntime(change runtimeChange) {
	if !change.keepAPI {
		startRuntimeService("@api", p.srvApi)
	}
	if !change.keepMetrics {
		startRuntimeService("@metrics", p.srvMetrics)
	}
	if !change.keepProfiling {
		startRuntimeService("@profiling", p.srvProfiling)
	}
}

func (p *program) restoreRuntime(previous *config.Config, change runtimeChange) error {
	api, metricsService, profiling, err := buildRuntimeServices(previous, change, p.reloadConfig)
	if err != nil {
		closeRuntimeServices(api, metricsService, profiling)
		return err
	}
	if !change.keepAPI {
		p.srvApi = api
	}
	if !change.keepMetrics {
		p.srvMetrics = metricsService
	}
	if !change.keepProfiling {
		p.srvProfiling = profiling
	}
	startRuntimeService("@api", api)
	startRuntimeService("@metrics", metricsService)
	startRuntimeService("@profiling", profiling)
	return nil
}

func buildRuntimeServices(cfg *config.Config, change runtimeChange, reload func() error) (api, metricsService, profiling service.Service, err error) {
	if cfg == nil {
		cfg = &config.Config{}
	}
	if !change.keepAPI && apiWanted(cfg.API) {
		api, err = buildApiService(cfg.API, reload)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("build API: %w", err)
		}
	}
	if !change.keepMetrics && metricsWanted(cfg.Metrics) {
		metricsService, err = buildMetricsService(cfg.Metrics)
		if err != nil {
			return api, nil, nil, fmt.Errorf("build metrics: %w", err)
		}
	}
	if !change.keepProfiling && profilingWanted(cfg.Profiling) {
		profiling, err = buildProfilingService(cfg.Profiling)
		if err != nil {
			return api, metricsService, nil, fmt.Errorf("build profiling: %w", err)
		}
	}
	return api, metricsService, profiling, nil
}

func startRegisteredServices() {
	for name, svc := range registry.ServiceRegistry().GetAll() {
		name, svc := name, svc
		go func() {
			if err := svc.Serve(); err != nil && !errors.Is(err, net.ErrClosed) {
				logger.Default().Debugf("service %s stopped: %v", name, err)
			}
		}()
	}
}

func closeRuntimeServices(services ...service.Service) {
	for _, svc := range services {
		if svc != nil {
			_ = svc.Close()
		}
	}
}

func startRuntimeService(name string, svc service.Service) {
	if svc == nil {
		return
	}
	go func() {
		log := logger.Default().WithFields(map[string]any{"kind": "service", "service": name})
		log.Info("listening on ", svc.Addr())
		if err := svc.Serve(); !errors.Is(err, http.ErrServerClosed) && !errors.Is(err, net.ErrClosed) {
			log.Error(err)
		}
	}()
}

func retireRuntimeService(svc service.Service) error {
	if retiree, ok := svc.(interface{ Retire() error }); ok {
		return retiree.Retire()
	}
	return svc.Close()
}

func (p *program) deactivateRuntime() error {
	var errs []error
	if p.srvApi != nil {
		errs = append(errs, wrapRuntimeClose("API", p.srvApi.Close()))
		p.srvApi = nil
	}
	if p.srvMetrics != nil {
		errs = append(errs, wrapRuntimeClose("metrics", p.srvMetrics.Close()))
		p.srvMetrics = nil
	}
	if p.srvProfiling != nil {
		errs = append(errs, wrapRuntimeClose("profiling", p.srvProfiling.Close()))
		p.srvProfiling = nil
	}
	xmetrics.Enable(false)
	return errors.Join(errs...)
}

func wrapRuntimeClose(name string, err error) error {
	if err == nil || errors.Is(err, net.ErrClosed) || errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return fmt.Errorf("close %s: %w", name, err)
}

func runtimeServiceUnchanged[T any](oldCfg, newCfg *T, wanted bool, running service.Service) bool {
	if !reflect.DeepEqual(oldCfg, newCfg) {
		return false
	}
	if wanted {
		return running != nil
	}
	return running == nil
}

func apiWanted(cfg *config.APIConfig) bool { return cfg != nil }
func metricsWanted(cfg *config.MetricsConfig) bool {
	return cfg != nil && cfg.Addr != ""
}
func profilingWanted(cfg *config.ProfilingConfig) bool { return cfg != nil }

func validateRuntimeConfig(cfg *config.Config) error {
	if cfg == nil {
		return nil
	}
	if apiWanted(cfg.API) {
		if err := validateListenAddress(cfg.API.Addr); err != nil {
			return fmt.Errorf("API address: %w", err)
		}
	}
	if metricsWanted(cfg.Metrics) {
		if err := validateListenAddress(cfg.Metrics.Addr); err != nil {
			return fmt.Errorf("metrics address: %w", err)
		}
	}
	if profilingWanted(cfg.Profiling) {
		addr := cfg.Profiling.Addr
		if addr == "" {
			addr = ":6060"
		}
		if _, err := net.ResolveTCPAddr("tcp", addr); err != nil {
			return fmt.Errorf("profiling address: %w", err)
		}
	}
	return nil
}

func validateListenAddress(addr string) error {
	if strings.HasPrefix(addr, "unix://") {
		if strings.TrimPrefix(addr, "unix://") == "" {
			return errors.New("empty unix socket path")
		}
		return nil
	}
	_, err := net.ResolveTCPAddr("tcp", addr)
	return err
}

type profilingService struct {
	server    *http.Server
	listener  net.Listener
	closeOnce sync.Once
	closeErr  error
}

func buildProfilingService(cfg *config.ProfilingConfig) (service.Service, error) {
	addr := cfg.Addr
	if addr == "" {
		addr = ":6060"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &profilingService{server: &http.Server{}, listener: ln}, nil
}

func (s *profilingService) Serve() error   { return s.server.Serve(s.listener) }
func (s *profilingService) Addr() net.Addr { return s.listener.Addr() }
func (s *profilingService) Close() error {
	s.closeOnce.Do(func() {
		listenerErr := s.listener.Close()
		if errors.Is(listenerErr, net.ErrClosed) {
			listenerErr = nil
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		serverErr := s.server.Shutdown(ctx)
		cancel()
		if errors.Is(serverErr, context.DeadlineExceeded) {
			serverErr = errors.Join(serverErr, s.server.Close())
		}
		if errors.Is(serverErr, http.ErrServerClosed) || errors.Is(serverErr, net.ErrClosed) {
			serverErr = nil
		}
		s.closeErr = errors.Join(serverErr, listenerErr)
	})
	return s.closeErr
}
