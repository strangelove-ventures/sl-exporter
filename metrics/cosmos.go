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
	valSignedBlock      *prometheus.GaugeVec
	valMissedBlocks     *prometheus.GaugeVec
	valSlashingWindow   *prometheus.GaugeVec
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
				Name: prometheus.BuildFQName(namespace, cosmosValSubsystem, "signed_blocks_total"),
				Help: "Count of observed blocks signed by a cosmos validator.",
			},
			[]string{"chain_id", "address"},
		),
		valSignedBlock: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName(namespace, cosmosValSubsystem, "latest_signed_block_height"),
				Help: "The latest observed block signed by a cosmos validator.",
			},
			[]string{"chain_id", "address"},
		),
		valMissedBlocks: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName(namespace, cosmosValSubsystem, "latest_missed_blocks"),
				Help: "The number of missed blocks within the signing window by a cosmos validator.",
			},
			[]string{"chain_id", "address"},
		),
		valSlashingWindow: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName(namespace, cosmosValSubsystem, "slashing_window_blocks"),
				Help: "The slashing window (in number of blocks) for a cosmos validator.",
			},
			[]string{"chain_id"},
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

// SetValSignedBlock sets latest signed block height for a validator.
func (c *Cosmos) SetValSignedBlock(chain, consaddress string, height float64) {
	c.valSignedBlock.WithLabelValues(chain, consaddress).Set(height)
}

// SetValMissedBlocks sets the number of missed blocks within the slashing window for a validator.
func (c *Cosmos) SetValMissedBlocks(chain, consaddress string, missed float64) {
	c.valMissedBlocks.WithLabelValues(chain, consaddress).Set(missed)
}

// SetValSlashingParams sets the slashing window for all validators on the chain.
// Accepts a struct for future expansion.
func (c *Cosmos) SetValSlashingParams(chain string, window float64) {
	c.valSlashingWindow.WithLabelValues(chain).Set(window)
}

// Metrics returns all metrics for Cosmos chains to be added to a Prometheus registry.
func (c *Cosmos) Metrics() []prometheus.Collector {
	return []prometheus.Collector{
		c.heightGauge,
		c.valJailGauge,
		c.valBlockSignCounter,
		c.valSignedBlock,
		c.valMissedBlocks,
		c.valSlashingWindow,
	}
}
