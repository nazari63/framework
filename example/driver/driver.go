package driver

import (
	"context"
	_ "net/http/pprof"
	"sync"
	"time"

	"github.com/base-org/framework/example/config"
	"github.com/base-org/framework/example/metrics"
	"github.com/base-org/framework/service"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
)

type Driver interface {
	service.Driver
	TxManager() txmgr.TxManager
	LatestHeader() *types.Header
}

func NewDriver(cfg *config.Config, l log.Logger, m metrics.Metricer) (Driver, error) {
	txManager, err := txmgr.NewSimpleTxManager("example", l, m, cfg.TxMgrConfig)
	if err != nil {
		return nil, err
	}
	client, err := ethclient.Dial(cfg.L1EthRpc)
	if err != nil {
		return nil, err
	}
	return &driver{
		cfg:       cfg,
		l:         l,
		m:         m,
		txManager: txManager,
		client:    client,
	}, nil
}

type driver struct {
	cfg *config.Config
	l   log.Logger
	m   metrics.Metricer

	txManager txmgr.TxManager
	client    eth.L1Client
	latest    *types.Header

	wg                sync.WaitGroup
	shutdownCtx       context.Context
	cancelShutdownCtx context.CancelFunc
}

func (d *driver) TxManager() txmgr.TxManager {
	return d.txManager
}

func (d *driver) LatestHeader() *types.Header {
	return d.latest
}

func (d *driver) Start(ctx context.Context) error {
	d.shutdownCtx, d.cancelShutdownCtx = context.WithCancel(context.Background())
	d.wg.Add(1)
	go d.loop()
	return nil
}

func (d *driver) Stop(ctx context.Context) error {
	d.l.Info("Stopping driver")

	d.txManager.Close()

	d.cancelShutdownCtx()
	d.wg.Wait()

	d.l.Info("Driver stopped")

	return nil
}

func (d *driver) loop() {
	defer d.wg.Done()

	ticker := time.NewTicker(d.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			d.l.Debug("Querying latest header")
			h, err := d.client.HeaderByNumber(context.Background(), nil)
			if err != nil {
				d.l.Error("failed to get latest header", "err", err)
				continue
			}
			d.latest = h
			d.m.RecordL1Ref("latest_block", eth.InfoToL1BlockRef(eth.HeaderBlockInfo(h)))

		case <-d.shutdownCtx.Done():
			d.l.Info("Main loop returning")
			return
		}
	}
}
