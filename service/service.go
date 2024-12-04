package service

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/ethereum-optimism/optimism/op-service/cliapp"
	"github.com/ethereum-optimism/optimism/op-service/httputil"
	opmetrics "github.com/ethereum-optimism/optimism/op-service/metrics"
	"github.com/ethereum-optimism/optimism/op-service/oppprof"
	oprpc "github.com/ethereum-optimism/optimism/op-service/rpc"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
)

var ErrAlreadyStopped = errors.New("already stopped")

type Service interface {
	cliapp.Lifecycle
	Kill() error
	WithDriver(driver Driver) Service
	WithMetrics(metrics Metricer) Service
	WithRPC(rpc rpc.API) Service
	WithRPCs(rpcs []rpc.API) Service
}

type service struct {
	config  Config
	version string
	log     log.Logger

	driver  Driver
	metrics Metricer
	rpcs    []rpc.API

	pprofService *oppprof.Service
	metricsSrv   *httputil.HTTPServer
	rpcServer    *oprpc.Server

	stopped atomic.Bool
}

func NewService(version string, cfg Config, log log.Logger) Service {
	return &service{
		config:  cfg,
		version: version,
		log:     log,
	}
}

func (s *service) WithDriver(driver Driver) Service {
	s.driver = driver
	return s
}

func (s *service) WithMetrics(metrics Metricer) Service {
	s.metrics = metrics
	return s
}

func (s *service) WithRPC(rpc rpc.API) Service {
	s.rpcs = append(s.rpcs, rpc)
	return s
}

func (s *service) WithRPCs(rpcs []rpc.API) Service {
	s.rpcs = append(s.rpcs, rpcs...)
	return s
}

func (s *service) Start(ctx context.Context) error {
	s.log.Info("Starting")

	if s.metrics != nil {
		if err := s.initMetricsServer(); err != nil {
			return fmt.Errorf("failed to start metrics service: %w", err)
		}
	}
	if err := s.initPProf(); err != nil {
		return fmt.Errorf("failed to init profiling: %w", err)
	}
	if len(s.rpcs) > 0 {
		if err := s.initRPCServer(); err != nil {
			return fmt.Errorf("failed to start RPC service: %w", err)
		}
	}

	if s.driver != nil {
		if err := s.driver.Start(ctx); err != nil {
			return fmt.Errorf("failed to start driver: %w", err)
		}
	}

	s.metrics.RecordInfo(s.version)
	s.metrics.RecordUp()

	return nil
}

func (s *service) initPProf() error {
	c := s.config.PprofConfig()
	s.pprofService = oppprof.New(
		c.ListenEnabled,
		c.ListenAddr,
		c.ListenPort,
		c.ProfileType,
		c.ProfileDir,
		c.ProfileFilename,
	)

	if err := s.pprofService.Start(); err != nil {
		return fmt.Errorf("failed to start pprof service: %w", err)
	}

	return nil
}

func (s *service) initMetricsServer() error {
	c := s.config.MetricsConfig()
	if !c.Enabled {
		s.log.Info("Metricer disabled")
		return nil
	}
	s.log.Debug("Starting metrics service", "addr", c.ListenAddr, "port", c.ListenPort)
	metricsSrv, err := opmetrics.StartServer(s.metrics.Registry(), c.ListenAddr, c.ListenPort)
	if err != nil {
		return fmt.Errorf("failed to start metrics service: %w", err)
	}
	s.log.Info("Started metrics service", "addr", metricsSrv.Addr())
	s.metricsSrv = metricsSrv
	return nil
}

func (s *service) initRPCServer() error {
	c := s.config.RPCConfig()
	server := oprpc.NewServer(
		c.ListenAddr,
		c.ListenPort,
		s.version,
		oprpc.WithLogger(s.log),
	)
	for _, r := range s.rpcs {
		admin := r.Namespace == "admin"
		if !admin || c.EnableAdmin {
			if admin {
				s.log.Info("Admin RPC enabled")
			}
			server.AddAPI(r)
		}
	}
	s.log.Info("Starting JSON-RPC service")
	if err := server.Start(); err != nil {
		return fmt.Errorf("unable to start RPC service: %w", err)
	}
	s.rpcServer = server
	return nil
}

// Stopped returns if the service as a whole is stopped.
func (s *service) Stopped() bool {
	return s.stopped.Load()
}

// Kill is a convenience method to forcefully, non-gracefully, stop the Service.
func (s *service) Kill() error {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return s.Stop(ctx)
}

// Stop fully stops the batch-submitter and all its resources gracefully. After stopping, it cannot be restarted.
// See driver.StopBatchSubmitting to temporarily stop the batch submitter.
// If the provided ctx is cancelled, the stopping is forced, i.e. the batching work is killed non-gracefully.
func (s *service) Stop(ctx context.Context) error {
	if s.stopped.Load() {
		return ErrAlreadyStopped
	}
	s.log.Info("Service stopping")

	var result error
	if s.driver != nil {
		if err := s.driver.Stop(ctx); err != nil {
			result = errors.Join(result, fmt.Errorf("failed to stop driver: %w", err))
		}
	}

	if s.rpcServer != nil {
		if err := s.rpcServer.Stop(); err != nil {
			result = errors.Join(result, fmt.Errorf("failed to stop RPC service: %w", err))
		}
	}
	if s.pprofService != nil {
		if err := s.pprofService.Stop(ctx); err != nil {
			result = errors.Join(result, fmt.Errorf("failed to stop PProf service: %w", err))
		}
	}

	if s.metricsSrv != nil {
		if err := s.metricsSrv.Stop(ctx); err != nil {
			result = errors.Join(result, fmt.Errorf("failed to stop metrics service: %w", err))
		}
	}

	if result == nil {
		s.stopped.Store(true)
		s.log.Info("Service stopped")
	}
	return result
}
