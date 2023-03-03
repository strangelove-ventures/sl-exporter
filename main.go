package main

import (
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	log "github.com/sirupsen/logrus"
)

var config Config

const (
	collector = "sl_exporter"
)

func main() {
	var err error
	var configFile, bind string
	// =====================
	// Get OS parameter
	// =====================
	flag.StringVar(&configFile, "config", "config.yaml", "configuration file")
	flag.StringVar(&bind, "bind", "localhost:9100", "bind")
	flag.Parse()

	// =====================
	// Load config & yaml
	// =====================
	var b []byte
	if b, err = ioutil.ReadFile(configFile); err != nil {
		log.Errorf("Failed to read config file: %s", err)
		os.Exit(1)
	}

	// Load yaml
	if err := yaml.Unmarshal(b, &config); err != nil {
		log.Errorf("Failed to load config: %s", err)
		os.Exit(1)
	}

	// ========================
	// Register handler
	// ========================
	promReg := prometheus.NewRegistry()
	log.Infof("Register version collector - %s", collector)
	promReg.Register(version.NewCollector(collector))
	promReg.Register(&QueryCollector{})

	// Register http handler
	log.Infof("HTTP handler path - %s", "/metrics")

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		h := promhttp.HandlerFor(promReg, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	})

	// start server
	log.Infof("Starting http server - %s", bind)
	if err := http.ListenAndServe(bind, nil); err != nil {
		log.Errorf("Failed to start http server: %s", err)
	}
}

type Sample struct {
	Labels []string
	Value  float64
}

type Metric struct {
	Name        string
	Type        string
	Description string
	Labels      []string
	metricDesc  *prometheus.Desc
	Samples     []Sample
}

// Config config structure
type Config struct {
	Metrics map[string]Metric
}

// QueryCollector exporter
type QueryCollector struct{}

// Describe prometheus describe
func (e *QueryCollector) Describe(ch chan<- *prometheus.Desc) {
	for metricName, metric := range config.Metrics {
		// Todo (nour): do we need namespace
		metric.metricDesc = prometheus.NewDesc(
			prometheus.BuildFQName("", "", metricName),
			metric.Description,
			metric.Labels, nil,
		)
		config.Metrics[metricName] = metric
		log.Infof("metric description for \"%s\" registerd", metricName)
	}
}

// Collect prometheus collect
func (e *QueryCollector) Collect(ch chan<- prometheus.Metric) {

	// Execute each queries in metrics
	for name, metric := range config.Metrics {
		for _, sample := range metric.Samples {
			// Add metric
			switch strings.ToLower(metric.Type) {
			case "gauge":
				ch <- prometheus.MustNewConstMetric(metric.metricDesc, prometheus.GaugeValue, sample.Value, sample.Labels...)
			default:
				log.Errorf("Fail to add metric for %s: %s is not valid type", name, metric.Type)
				continue
			}
		}
	}
}
