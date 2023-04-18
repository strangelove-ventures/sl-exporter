package cmd

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"net/url"
)

// registerMetrics iterates config metrics and passes them to relevant handler
func registerMetrics(config *Config, registry *prometheus.Registry) {
	for metricName, metric := range config.Metrics {
		var collector prometheus.Collector

		switch metric.Type {
		case "samples":
			collector = sampleMetrics(metric, metricName)
		case "rpc":
			collector = rpcMetrics(metric, metricName)
		default:
			log.Fatalf("Unsupported metric type: %s", metric.Type)
		}

		if err := registry.Register(collector); err != nil {
			log.Fatalf("Error registering %s: %v", metricName, err)
		}
		log.Infof("Register %s collector - %s", metric.Type, metricName)
	}
}

// sampleMetrics handles static gauge samples
func sampleMetrics(metric Metric, metricName string) *prometheus.GaugeVec {
	gaugeVec := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: prometheus.BuildFQName(namespace, subsystem, metricName), Help: metric.Description},
		metric.Labels,
	)
	for _, sample := range metric.Samples {
		gaugeVec.WithLabelValues(sample.Labels...).Set(sample.Value)
	}
	return gaugeVec
}

// rpcMetrics handles dynamic gauge metrics for public_rpc_node_height
func rpcMetrics(metric Metric, metricName string) *prometheus.GaugeVec {
	gaugeVec := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: prometheus.BuildFQName(namespace, subsystem, metricName), Help: metric.Description},
		metric.Labels,
	)

	for _, chain := range metric.Chains {
		for _, rpc := range chain.RPCs {
			// Fetch and set the metric value for the rpc node
			value, err := fetchRPCNodeHeight(rpc)
			if err != nil {
				log.Errorf("Error fetching height for rpc %s: %v", rpc, err)
				continue
			}
			gaugeVec.WithLabelValues(chain.Name, urlHost(rpc)).Set(value)
		}
	}
	return gaugeVec
}

// urlHost extracts host from given rpc url
func urlHost(rpc string) string {
	parsedURL, err := url.Parse(rpc)
	if err != nil {
		log.Warnf("Error parsing URL: %v\n", err)
		return rpc
	}
	return parsedURL.Hostname()
}

func fetchRPCNodeHeight(rpcURL string) (float64, error) {
	// Implement the logic to fetch node height from the rpcURL
	// ...

	return 0, nil
}
