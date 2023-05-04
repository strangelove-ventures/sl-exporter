package cmd

import (
	"flag"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	log "github.com/sirupsen/logrus"
	"github.com/strangelove-ventures/sl-exporter/metrics"
)

const collector = "sl_exporter"

func Execute() {
	var cfg Config

	flag.StringVar(&cfg.File, "config", "config.yaml", "Path to configuration file")
	flag.StringVar(&cfg.BindAddr, "bind", ":9100", "Address to bind")
	flag.Parse()

	if err := parseConfig(&cfg); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	// Initialize prometheus registry
	registry := prometheus.NewRegistry()
	registry.MustRegister(version.NewCollector(collector))

	// Register static metrics
	registry.MustRegister(metrics.BuildStatic(cfg.Static.Gauges)...)

	// Register cosmos chain metrics
	cosmos := metrics.NewCosmos()
	registry.MustRegister(cosmos.Metrics()...)

	// Start the server
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{Timeout: 60 * time.Second}))
	log.Infof("Starting Prometheus metrics server - %s", cfg.BindAddr)
	if err := http.ListenAndServe(cfg.BindAddr, nil); err != nil {
		log.Fatalf("Failed to start http server: %v", err)
	}
}
