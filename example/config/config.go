package config

import (
	"time"

	"github.com/base-org/framework/example/flags"
	"github.com/base-org/framework/service"
	"github.com/urfave/cli/v2"

	"github.com/ethereum-optimism/optimism/op-service/txmgr"
)

type Config struct {
	service.Config

	TxMgrConfig txmgr.CLIConfig

	L1EthRpc     string
	PollInterval time.Duration
}

func (c *Config) Check() error {
	if err := c.Config.Check(); err != nil {
		return err
	}
	if err := c.TxMgrConfig.Check(); err != nil {
		return err
	}
	return nil
}

// NewConfig parses the Config from the provided flags or environment variables.
func NewConfig(ctx *cli.Context) *Config {
	return &Config{
		Config: service.NewConfig(ctx),

		TxMgrConfig: txmgr.ReadCLIConfig(ctx),

		L1EthRpc:     ctx.String(flags.L1EthRpcFlag.Name),
		PollInterval: ctx.Duration(flags.PollIntervalFlag.Name),
	}
}
