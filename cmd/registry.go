package cmd

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"net/http"
)

const (
	collector = "sl_exporter"
	// Todo (nour): should we add namespace and subsystem
	namespace = ""
	subsystem = ""
)

var (
	metricsRegistry *prometheus.Registry
)

func updateRegistry(config *Config) (*prometheus.Registry, error) {
	// Create sampleMetrics new registry for the updated metrics
	newRegistry := prometheus.NewRegistry()

	// Register build_info metric
	if err := newRegistry.Register(version.NewCollector(collector)); err != nil {
		return nil, err
	}

	if err := registerMetrics(config, newRegistry); err != nil {
		return nil, err
	}

	return newRegistry, nil
}

func metricsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := promhttp.HandlerFor(metricsRegistry, promhttp.HandlerOpts{})
		handler.ServeHTTP(w, r)
	})
}
