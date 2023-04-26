package cmd

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

func Execute() {
	config, err := readConfig(cmdConfig.ConfigFile)
	if err != nil {
		log.Fatalln(err)
	}
	registry, err := buildRegistry(config)
	if err != nil {
		log.Fatalln(err)
	}

	http.Handle("/metrics", metricsHandler(registry))
	log.Infof("Starting Prometheus metrics server - %s", cmdConfig.Bind)
	if err := http.ListenAndServe(cmdConfig.Bind, nil); err != nil {
		log.Fatalf("Failed to start http server: %v", err)
	}
}
