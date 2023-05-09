package metrics

import (
	"net/url"

	"github.com/prometheus/client_golang/prometheus"
)

// Cosmos records metrics for Cosmos chains
type Cosmos struct {
	heightGauge    *prometheus.GaugeVec
	valStatusGauge *prometheus.GaugeVec
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
		valStatusGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName(Namespace, CosmosSubsystem, "val_jailed_status"),
				Help: "Whether a validator is active, jailed, or tombstoned.",
			},
			// labels
			[]string{"chain_id", "source", "address", "status"},
		),
	}
}

// SetNodeHeight records the block height on the public_rpc_node_height gauge.
func (c *Cosmos) SetNodeHeight(chain string, restURL url.URL, height float64) {
	source := restURL.Hostname() + restURL.Path
	c.heightGauge.WithLabelValues(chain, source).Set(height)
}

// ValStatus is the status of a validator.
type ValStatus string

const (
	ValStatusActive     ValStatus = "active"
	ValStatusJailed     ValStatus = "jailed"
	ValStatusTombstoned ValStatus = "tombstoned"
)

// SetValJailStatus records the jailed status of a validator. Gauge set to 1 if the validator is jailed or tombstoned.
// In this context, "active" does not mean part of the validator active set.
func (c *Cosmos) SetValJailStatus(chain, consaddress string, restURL url.URL, status ValStatus) {
	source := restURL.Hostname() + restURL.Path
	var value float64
	if status != ValStatusActive {
		value = 1
	}
	c.valStatusGauge.WithLabelValues(chain, source, consaddress, string(status)).Set(value)
}

// Metrics returns all metrics for Cosmos chains to be added to a Prometheus registry.
func (c *Cosmos) Metrics() []prometheus.Collector {
	return []prometheus.Collector{
		c.heightGauge,
		c.valStatusGauge,
	}
}
