package cmd

import (
	"fmt"
	"net/url"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// registerMetrics iterates config metrics and passes them to relevant handler
func registerMetrics(config Config, registry *prometheus.Registry) error {
	for metricName, metric := range config.Metrics {
		var collector prometheus.Collector

		if len(metric.Chains) > 0 {
			collector = rpcMetrics(metric, metricName)
		} else {
			return fmt.Errorf("unsupported metric: %s", metricName)
		}

		if err := registry.Register(collector); err != nil {
			return fmt.Errorf("error registering %s: %w", metricName, err)
		}
		log.Infof("Register collector - %s", metricName)
	}
	return nil
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
			gaugeVec.WithLabelValues(chain.Name, urlHost(rpc)).Set(float64(value))
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
