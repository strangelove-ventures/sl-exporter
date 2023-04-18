package cmd

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func Execute() {
	configFile, interval, bind := flags()

	// Initialize metricsRegistry
	metricsRegistry = prometheus.NewRegistry()

	// Start background goroutine to re-evaluate the config and update the registry at `interval`
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				config, err := readConfig(configFile)
				if err != nil {
					log.Errorf("Error reading config file %s: %v", configFile, err)
					continue
				}
				updatedRegistry := updateRegistry(config)
				if updatedRegistry != nil {
					mu.Lock()
					metricsRegistry = updatedRegistry
					mu.Unlock()
				}
			}
		}
	}()

	http.Handle("/metrics", metricsHandler())
	log.Infof("Starting Prometheus metrics server - %s", bind)
	if err := http.ListenAndServe(bind, nil); err != nil {
		log.Fatalf("Failed to start http server: %v", err)
	}
}
