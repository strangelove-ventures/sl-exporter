package metrics

import "github.com/prometheus/client_golang/prometheus"

type StaticGauge struct {
	Name        string
	Description string
	Labels      []string
	Samples     []StaticSample
}

type StaticSample struct {
	Labels []string
	Value  float64
}

// BuildStatic returns static metrics
func BuildStatic(gauges []StaticGauge) []prometheus.Collector {
	metrics := make([]prometheus.Collector, len(gauges))
	for i, g := range gauges {
		gaugeVec := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{Name: prometheus.BuildFQName(Namespace, CosmosSubsystem, g.Name), Help: g.Description},
			g.Labels,
		)
		for _, sample := range g.Samples {
			gaugeVec.WithLabelValues(sample.Labels...).Set(sample.Value)
		}
		metrics[i] = gaugeVec
	}
	return metrics
}
