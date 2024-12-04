package service

import (
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	opmetrics "github.com/ethereum-optimism/optimism/op-service/metrics"
	"github.com/ethereum-optimism/optimism/op-service/oppprof"
	oprpc "github.com/ethereum-optimism/optimism/op-service/rpc"
	"github.com/urfave/cli/v2"
)

type Config interface {
	Check() error
	LogConfig() oplog.CLIConfig
	MetricsConfig() opmetrics.CLIConfig
	PprofConfig() oppprof.CLIConfig
	RPCConfig() oprpc.CLIConfig
}

type config struct {
	logConfig     oplog.CLIConfig
	metricsConfig opmetrics.CLIConfig
	pprofConfig   oppprof.CLIConfig
	rpcConfig     oprpc.CLIConfig
}

func NewConfig(ctx *cli.Context) Config {
	return &config{
		logConfig:     oplog.ReadCLIConfig(ctx),
		metricsConfig: opmetrics.ReadCLIConfig(ctx),
		pprofConfig:   oppprof.ReadCLIConfig(ctx),
		rpcConfig:     oprpc.ReadCLIConfig(ctx),
	}
}

func (c *config) Check() error {
	if err := c.metricsConfig.Check(); err != nil {
		return err
	}
	if err := c.pprofConfig.Check(); err != nil {
		return err
	}
	if err := c.rpcConfig.Check(); err != nil {
		return err
	}
	return nil
}

func (c *config) LogConfig() oplog.CLIConfig {
	return c.logConfig
}

func (c *config) MetricsConfig() opmetrics.CLIConfig {
	return c.metricsConfig
}

func (c *config) PprofConfig() oppprof.CLIConfig {
	return c.pprofConfig
}

func (c *config) RPCConfig() oprpc.CLIConfig {
	return c.rpcConfig
}
