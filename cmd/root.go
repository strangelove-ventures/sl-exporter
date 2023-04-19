package cmd

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func Execute() {
	// Initialize metricsRegistry
	metricsRegistry = prometheus.NewRegistry()

	// Start background goroutine to re-evaluate the config and update the registry at `interval`
	go func() {
		ticker := time.NewTicker(cmdConfig.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				config, err := readConfig(cmdConfig.ConfigFile)
				if err != nil {
					log.Errorf("Error reading config file %s: %v", cmdConfig.ConfigFile, err)
					continue
				}
				if updatedRegistry, err := updateRegistry(config); err != nil {
					log.Errorf("error updating registery: %v", err)
				} else {
					metricsRegistry = updatedRegistry
				}
			}
		}
	}()

	http.Handle("/metrics", metricsHandler())
	log.Infof("Starting Prometheus metrics server - %s", cmdConfig.Bind)
	if err := http.ListenAndServe(cmdConfig.Bind, nil); err != nil {
		log.Fatalf("Failed to start http server: %v", err)
	}
}
