package metrics

import (
	"net/url"

	"github.com/prometheus/client_golang/prometheus"
)

// Comet records CometBFT (formerly Tendermint) metrics.
type Comet struct {
	heightGauge *prometheus.GaugeVec
}

func NewComet() *Comet {
	return &Comet{
		heightGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName(Namespace, Subsystem, "public_rpc_node_height"),
				Help: "Node height of a public RPC node",
			},
			[]string{"chain", "source"}, // rpc height labels
		),
	}
}

// SetNodeHeight records the block height on the public_rpc_node_height gauge.
func (c *Comet) SetNodeHeight(chain string, rpcURL url.URL, height float64) {
	source := rpcURL.Hostname() + rpcURL.Path
	c.heightGauge.WithLabelValues(chain, source).Set(height)
}

// Metrics returns all metrics for Comet chains to be added to a Prometheus registry.
func (c *Comet) Metrics() []prometheus.Collector {
	return []prometheus.Collector{
		c.heightGauge,
	}
}
