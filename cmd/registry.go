package cmd

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
)

const (
	collector = "sl_exporter"
	// Todo (nour): should we add namespace and subsystem
	namespace = ""
	subsystem = ""
)

func buildRegistry(config Config) (*prometheus.Registry, error) {
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

func metricsHandler(reg *prometheus.Registry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
		handler.ServeHTTP(w, r)
	})
}
