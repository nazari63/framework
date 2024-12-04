package main

import (
	"context"
	"fmt"
	"os"

	"github.com/base-org/framework/example/api"
	"github.com/base-org/framework/example/config"
	"github.com/base-org/framework/example/driver"
	"github.com/base-org/framework/example/flags"
	"github.com/base-org/framework/example/metrics"
	"github.com/base-org/framework/service"
	"github.com/urfave/cli/v2"

	opservice "github.com/ethereum-optimism/optimism/op-service"
	"github.com/ethereum-optimism/optimism/op-service/cliapp"
	"github.com/ethereum-optimism/optimism/op-service/ctxinterrupt"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum/go-ethereum/log"
)

// autopopulated by the Makefile
var (
	Version   = ""
	GitCommit = ""
	GitDate   = ""
)

func main() {
	oplog.SetupDefaults()

	app := cli.NewApp()
	app.Flags = cliapp.ProtectFlags(flags.Flags)
	app.Version = opservice.FormatVersion(Version, GitCommit, GitDate, "")
	app.Name = "example"
	app.Usage = "Example Service"
	app.Description = "Example service that uses the Optimism Service Framework."
	app.Action = cliapp.LifecycleCmd(Main(Version))

	ctx := ctxinterrupt.WithSignalWaiterMain(context.Background())
	err := app.RunContext(ctx, os.Args)
	if err != nil {
		log.Crit("Application failed", "message", err)
	}
}

func Main(version string) cliapp.LifecycleAction {
	return func(cliCtx *cli.Context, closeApp context.CancelCauseFunc) (cliapp.Lifecycle, error) {
		cfg := config.NewConfig(cliCtx)
		if err := cfg.Check(); err != nil {
			return nil, fmt.Errorf("invalid CLI flags: %w", err)
		}

		l := oplog.NewLogger(oplog.AppOut(cliCtx), cfg.LogConfig())
		oplog.SetGlobalLogHandler(l.Handler())
		opservice.ValidateEnvVars(flags.EnvVarPrefix, flags.Flags, l)

		s := service.NewService(version, cfg, l)

		m := metrics.NewMetrics("")
		s.WithMetrics(m)

		d, err := driver.NewDriver(cfg, l, m)
		if err != nil {
			return nil, fmt.Errorf("failed to create driver: %w", err)
		}
		s.WithDriver(d)

		s.WithRPC(d.TxManager().API())
		s.WithRPC(api.NewAPI(cfg, l, m, d))

		return s, nil
	}
}
