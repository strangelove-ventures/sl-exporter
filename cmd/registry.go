package cmd

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sync"
)

const (
	collector = "sl_exporter"
	// Todo (nour): should we add namespace and subsystem
	namespace = ""
	subsystem = ""
)

var (
	metricsRegistry *prometheus.Registry
	mu              sync.Mutex
)

func updateRegistry(config *Config) *prometheus.Registry {
	// Create sampleMetrics new registry for the updated metrics
	newRegistry := prometheus.NewRegistry()

	// Register build_info metric
	if err := newRegistry.Register(version.NewCollector(collector)); err != nil {
		log.Errorf("Error registering build_info : %v", err)
		return nil
	}

	registerMetrics(config, newRegistry)

	return newRegistry
}

func metricsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		handler := promhttp.HandlerFor(metricsRegistry, promhttp.HandlerOpts{})
		handler.ServeHTTP(w, r)
	})
}
