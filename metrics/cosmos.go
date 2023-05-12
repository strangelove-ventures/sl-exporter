package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/strangelove-ventures/sl-exporter/cosmos"
)

// Cosmos records metrics for Cosmos chains
type Cosmos struct {
	heightGauge         *prometheus.GaugeVec
	valJailGauge        *prometheus.GaugeVec
	valBlockSignCounter *prometheus.CounterVec
}

func NewCosmos() *Cosmos {
	return &Cosmos{
		heightGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName(namespace, cosmosSubsystem, "latest_block_height"),
				Help: "Latest block height of a cosmos node.",
			},
			[]string{"chain_id"},
		),
		valJailGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName(namespace, cosmosValSubsystem, "latest_jailed_status"),
				Help: "0 if the cosmos validator is not jailed. 1 if the validator is jailed. 2 if the validator is tombstoned.",
			},
			[]string{"chain_id", "address"},
		),
		valBlockSignCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: prometheus.BuildFQName(namespace, cosmosValSubsystem, "signed_blocks_count"),
				Help: "Count of observed blocks signed by a cosmos validator.",
			},
			[]string{"chain_id", "address"},
		),
	}
}

// SetNodeHeight records the block height on the public_rpc_node_height gauge.
func (c *Cosmos) SetNodeHeight(chain string, height float64) {
	c.heightGauge.WithLabelValues(chain).Set(height)
}

// SetValJailStatus records the jailed status of a validator.
// In this context, "active" does not mean part of the validator active set, only that the validator is not jailed.
func (c *Cosmos) SetValJailStatus(chain, consaddress string, status cosmos.JailStatus) {
	c.valJailGauge.WithLabelValues(chain, consaddress).Set(float64(status))
}

// IncValSignedBlocks increments the number of blocks signed by validator at consaddress.
func (c *Cosmos) IncValSignedBlocks(chain, consaddress string) {
	c.valBlockSignCounter.WithLabelValues(chain, consaddress).Inc()
}

// Metrics returns all metrics for Cosmos chains to be added to a Prometheus registry.
func (c *Cosmos) Metrics() []prometheus.Collector {
	return []prometheus.Collector{
		c.heightGauge,
		c.valJailGauge,
		c.valBlockSignCounter,
	}
}
