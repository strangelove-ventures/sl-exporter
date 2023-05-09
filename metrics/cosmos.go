package metrics

import (
	"net/url"

	"github.com/prometheus/client_golang/prometheus"
)

// Cosmos records metrics for Cosmos chains
type Cosmos struct {
	heightGauge  *prometheus.GaugeVec
	valJailGauge *prometheus.GaugeVec
}

func NewCosmos() *Cosmos {
	return &Cosmos{
		heightGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName(Namespace, CosmosSubsystem, "latest_block_height"),
				Help: "Latest block height of a cosmos node.",
			},
			// labels
			[]string{"chain_id", "source"},
		),
		valJailGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName(Namespace, CosmosValSubsystem, "latest_jailed_status"),
				Help: "0 if the validator is not jailed. 1 if the validator is jailed. 2 if the validator is tombstoned.",
			},
			// labels
			[]string{"chain_id", "source", "address"},
		),
	}
}

// SetNodeHeight records the block height on the public_rpc_node_height gauge.
func (c *Cosmos) SetNodeHeight(chain string, restURL url.URL, height float64) {
	source := restURL.Hostname() + restURL.Path
	c.heightGauge.WithLabelValues(chain, source).Set(height)
}

// JailStatus is the status of a validator.
type JailStatus int

const (
	JailStatusActive JailStatus = iota
	JailStatusJailed
	JailStatusTombstoned
)

// SetValJailStatus records the jailed status of a validator. Gauge set to 1 if the validator is jailed or tombstoned.
// In this context, "active" does not mean part of the validator active set.
func (c *Cosmos) SetValJailStatus(chain, consaddress string, restURL url.URL, status JailStatus) {
	source := restURL.Hostname() + restURL.Path
	c.valJailGauge.WithLabelValues(chain, source, consaddress).Set(float64(status))
}

// Metrics returns all metrics for Cosmos chains to be added to a Prometheus registry.
func (c *Cosmos) Metrics() []prometheus.Collector {
	return []prometheus.Collector{
		c.heightGauge,
		c.valJailGauge,
	}
}
