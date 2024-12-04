package service

import opmetrics "github.com/ethereum-optimism/optimism/op-service/metrics"

type Metricer interface {
	opmetrics.RegistryMetricer
	RecordInfo(version string)
	RecordUp()
}
