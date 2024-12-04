package flags

import (
	"time"

	"github.com/urfave/cli/v2"

	opservice "github.com/ethereum-optimism/optimism/op-service"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	opmetrics "github.com/ethereum-optimism/optimism/op-service/metrics"
	"github.com/ethereum-optimism/optimism/op-service/oppprof"
	oprpc "github.com/ethereum-optimism/optimism/op-service/rpc"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
)

const EnvVarPrefix = "CB_EXAMPLE"

func prefixEnvVars(name string) []string {
	return opservice.PrefixEnvVar(EnvVarPrefix, name)
}

var (
	L1EthRpcFlag = &cli.StringFlag{
		Name:     "l1-eth-rpc",
		Usage:    "HTTP provider URL for L1",
		EnvVars:  prefixEnvVars("L1_ETH_RPC"),
		Required: true,
	}
	PollIntervalFlag = &cli.DurationFlag{
		Name:    "poll-interval",
		Usage:   "How frequently to poll L2 for new blocks",
		Value:   6 * time.Second,
		EnvVars: prefixEnvVars("POLL_INTERVAL"),
	}
)

// Flags contains the list of configuration options available to the binary.
var Flags = []cli.Flag{
	L1EthRpcFlag,
	PollIntervalFlag,
}

func init() {
	Flags = append(Flags, txmgr.CLIFlags(EnvVarPrefix)...)
	Flags = append(Flags, oprpc.CLIFlags(EnvVarPrefix)...)
	Flags = append(Flags, oplog.CLIFlags(EnvVarPrefix)...)
	Flags = append(Flags, opmetrics.CLIFlags(EnvVarPrefix)...)
	Flags = append(Flags, oppprof.CLIFlags(EnvVarPrefix)...)
}
