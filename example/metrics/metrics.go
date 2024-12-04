package metrics

import (
	"github.com/base-org/framework/service"
	"github.com/prometheus/client_golang/prometheus"

	opmetrics "github.com/ethereum-optimism/optimism/op-service/metrics"
	txmetrics "github.com/ethereum-optimism/optimism/op-service/txmgr/metrics"
)

const Namespace = "cb_example"

type Metricer interface {
	service.Metricer
	opmetrics.RefMetricer
	txmetrics.TxMetricer
	opmetrics.RPCMetricer

	Document() []opmetrics.DocumentedMetric
}

type metrics struct {
	ns       string
	registry *prometheus.Registry
	factory  opmetrics.Factory

	opmetrics.RefMetrics
	txmetrics.TxMetrics
	opmetrics.RPCMetrics

	info prometheus.GaugeVec
	up   prometheus.Gauge
}

func NewMetrics(procName string) Metricer {
	if procName == "" {
		procName = "default"
	}
	ns := Namespace + "_" + procName

	registry := opmetrics.NewRegistry()
	factory := opmetrics.With(registry)

	return &metrics{
		ns:       ns,
		registry: registry,
		factory:  factory,

		RefMetrics: opmetrics.MakeRefMetrics(ns, factory),
		TxMetrics:  txmetrics.MakeTxMetrics(ns, factory),
		RPCMetrics: opmetrics.MakeRPCMetrics(ns, factory),

		info: *factory.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "info",
			Help:      "Pseudo-metric tracking version and config info",
		}, []string{
			"version",
		}),
		up: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "up",
			Help:      "1 if the op-batcher has finished starting up",
		}),
	}
}

func (m *metrics) Registry() *prometheus.Registry {
	return m.registry
}

func (m *metrics) Document() []opmetrics.DocumentedMetric {
	return m.factory.Document()
}

// RecordInfo sets a pseudo-metric that contains versioning and
// config info for the op-batcher.
func (m *metrics) RecordInfo(version string) {
	m.info.WithLabelValues(version).Set(1)
}

// RecordUp sets the up metric to 1.
func (m *metrics) RecordUp() {
	prometheus.MustRegister()
	m.up.Set(1)
}
