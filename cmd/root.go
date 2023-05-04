package cmd

import (
	"flag"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/strangelove-ventures/sl-exporter/metrics"
)

func Execute() {
	var cfg Config

	flag.StringVar(&cfg.File, "config", "config.yaml", "Path to configuration file")
	flag.StringVar(&cfg.BindAddr, "bind", ":9100", "Address to bind")
	flag.StringVar(&cfg.LogLevel, "loglevel", "info", "Log level (debug, info, warn, error)")
	flag.Parse()

	if err := parseConfig(&cfg); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	registry, err := buildRegistry(cfg)
	if err != nil {
		log.Fatalln(err)
	}
	// Register static metrics
	registry.MustRegister(metrics.BuildStatic(cfg.Static.Gauges)...)

	http.Handle("/metrics", metricsHandler(registry))
	log.Infof("Starting Prometheus metrics server - %s", cfg.BindAddr)
	if err := http.ListenAndServe(cfg.BindAddr, nil); err != nil {
		log.Fatalf("Failed to start http server: %v", err)
	}
}
