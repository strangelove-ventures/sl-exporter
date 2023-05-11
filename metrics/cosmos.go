package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/strangelove-ventures/sl-exporter/cosmos"
)

// Cosmos records metrics for Cosmos chains
type Cosmos struct {
	heightGauge  *prometheus.GaugeVec
	valJailGauge *prometheus.GaugeVec
}

func NewCosmos() *Cosmos {
	const (
		subsystem    = "cosmos"
		valSubsystem = subsystem + "_val"
	)

	return &Cosmos{
		heightGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName(namespace, subsystem, "latest_block_height"),
				Help: "Latest block height of a cosmos node.",
			},
			// labels
			[]string{"chain_id"},
		),
		valJailGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName(namespace, valSubsystem, "latest_jailed_status"),
				Help: "0 if the validator is not jailed. 1 if the validator is jailed. 2 if the validator is tombstoned.",
			},
			// labels
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

// Metrics returns all metrics for Cosmos chains to be added to a Prometheus registry.
func (c *Cosmos) Metrics() []prometheus.Collector {
	return []prometheus.Collector{
		c.heightGauge,
		c.valJailGauge,
	}
}
