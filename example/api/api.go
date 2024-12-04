package api

import (
	"context"

	"github.com/base-org/framework/example/config"
	"github.com/base-org/framework/example/driver"
	"github.com/base-org/framework/example/metrics"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
)

type api struct {
	cfg *config.Config
	l   log.Logger
	m   metrics.Metricer
	d   driver.Driver
}

func NewAPI(cfg *config.Config, l log.Logger, m metrics.Metricer, d driver.Driver) rpc.API {
	a := &api{
		cfg: cfg,
		l:   l,
		m:   m,
		d:   d,
	}
	return rpc.API{
		Namespace: "example",
		Service:   a,
	}
}

func (a *api) LatestHeader() (*types.Header, error) {
	recordDur := a.m.RecordRPCServerRequest("example_latestHeader")
	defer recordDur()
	a.l.Info("example_latestHeader")
	return a.d.LatestHeader(), nil
}

func (a *api) SendTx(ctx context.Context, candidate txmgr.TxCandidate) (*types.Receipt, error) {
	recordDur := a.m.RecordRPCServerRequest("example_sendTx")
	defer recordDur()
	a.l.Info("example_sendTx", "candidate", candidate)
	return a.d.TxManager().Send(ctx, candidate)
}
