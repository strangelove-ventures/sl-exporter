package metrics

import (
	"net/url"

	"github.com/prometheus/client_golang/prometheus"
)

// Cosmos records metrics for Cosmos chains
type Cosmos struct {
	rpcHeightGauge *prometheus.GaugeVec
}

func NewCosmos() *Cosmos {
	return &Cosmos{
		rpcHeightGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName(Namespace, CosmosSubsystem, "latest_block_height"),
				Help: "Latest block height of a cosmos node.",
			},
			[]string{"chain_id", "source"}, // rpc height labels
		),
	}
}

// SetNodeHeight records the block height on the public_rpc_node_height gauge.
func (c *Cosmos) SetNodeHeight(chain string, rpcURL url.URL, height float64) {
	source := rpcURL.Hostname() + rpcURL.Path
	c.rpcHeightGauge.WithLabelValues(chain, source).Set(height)
}

// Metrics returns all metrics for Cosmos chains to be added to a Prometheus registry.
func (c *Cosmos) Metrics() []prometheus.Collector {
	return []prometheus.Collector{
		c.rpcHeightGauge,
	}
}
