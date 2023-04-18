package cmd

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"time"
)

type Config struct {
	Metrics map[string]Metric `yaml:"metrics"`
}

type Metric struct {
	Type        string          `yaml:"type"`
	Description string          `yaml:"description"`
	Labels      []string        `yaml:"labels"`
	Samples     []Sample        `yaml:"samples,omitempty"`
	Chains      []ChainWithRPCs `yaml:"chains,omitempty"`
}

type ChainWithRPCs struct {
	Name string   `yaml:"name"`
	RPCs []string `yaml:"rpcs"`
}

type Sample struct {
	Labels []string `yaml:"labels"`
	Value  float64  `yaml:"value"`
}

// flags returns CLI flags
func flags() (string, time.Duration, string) {
	var configFile, bind, logLevel string
	var interval time.Duration
	flag.StringVar(&configFile, "config", "config.yaml", "configuration file")
	flag.StringVar(&bind, "bind", "localhost:9100", "bind")
	flag.DurationVar(&interval, "interval", 15*time.Second, "duration interval")
	flag.StringVar(&logLevel, "loglevel", "info", "Log level (debug, info, warn, error)")
	flag.Parse()

	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Fatalf("Invalid log level: %v", err)
	}
	log.SetLevel(level)

	log.Debugf("Config File: %s\n", configFile)
	log.Debugf("Interval: %s\n", interval)
	log.Debugf("Bind: %s\n", bind)

	return configFile, interval, bind
}

// readConfig reads config.yaml from disk
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
