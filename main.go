package main

import (
	"flag"
	"github.com/prometheus/common/version"
	"io"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Metrics map[string]Metric `yaml:"metrics"`
}

type Metric struct {
	Type        string   `yaml:"type"`
	Description string   `yaml:"description"`
	Labels      []string `yaml:"labels"`
	Samples     []Sample `yaml:"samples"`
}

type Sample struct {
	Labels []string `yaml:"labels"`
	Value  float64  `yaml:"value"`
}

const (
	collector = "sl_exporter"
)

func main() {
	var configFile, bind string
	flag.StringVar(&configFile, "config", "config.yaml", "configuration file")
	flag.StringVar(&bind, "bind", "localhost:9100", "bind")
	flag.Parse()

	config, err := readConfig(configFile)
	if err != nil {
		log.Fatalf("Error reading config file %s: %v", configFile, err)
	}

	// Create registry
	registry := prometheus.NewRegistry()

	// Register build_info metric
	if err := registry.Register(version.NewCollector(collector)); err != nil {
		log.Fatalf("Error registering build_info : %v", err)
	}
	log.Infof("Register version collector - %s", collector)

	registerStaticMetrics(config, registry)

	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	log.Infof("Starting Prometheus metrics server - %s", bind)
	if err := http.ListenAndServe(bind, nil); err != nil {
		log.Fatalf("Failed to start http server: %v", err)
	}
}

func readConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	// Load yaml
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func registerStaticMetrics(config *Config, registry *prometheus.Registry) {
	// Iterate config metrics
	for metricName, metric := range config.Metrics {
		var collector prometheus.Collector

		switch metric.Type {
		case "gauge":
			gaugeVec := prometheus.NewGaugeVec(
				// Todo (nour): should we add namespace and subsystem
				prometheus.GaugeOpts{Name: prometheus.BuildFQName("", "", metricName), Help: metric.Description},
				metric.Labels,
			)
			for _, sample := range metric.Samples {
				gaugeVec.WithLabelValues(sample.Labels...).Set(sample.Value)
			}
			collector = gaugeVec
		default:
			log.Fatalf("Unsupported metric type: %s", metric.Type)
		}

		if err := registry.Register(collector); err != nil {
			log.Fatalf("Error registering %s: %v", metricName, err)
		}
		log.Infof("Register %s collector - %s", metric.Type, metricName)
	}
}
