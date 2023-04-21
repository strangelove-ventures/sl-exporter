package cmd

import (
	"flag"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Metrics map[string]Metric `yaml:"metrics"`
}

type Metric struct {
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

type CMDConfig struct {
	ConfigFile string
	Bind       string
	Interval   time.Duration
	LogLevel   string
}

var cmdConfig CMDConfig

func init() {
	flag.StringVar(&cmdConfig.ConfigFile, "config", "config.yaml", "configuration file")
	flag.StringVar(&cmdConfig.Bind, "bind", "localhost:9100", "bind")
	flag.DurationVar(&cmdConfig.Interval, "interval", 15*time.Second, "duration interval")
	flag.StringVar(&cmdConfig.LogLevel, "loglevel", "info", "Log level (debug, info, warn, error)")
	flag.Parse()

	level, err := log.ParseLevel(cmdConfig.LogLevel)
	if err != nil {
		log.Fatalf("Invalid log level: %v", err)
	}
	log.SetLevel(level)

	log.Debugf("Config File: %s\n", cmdConfig.ConfigFile)
	log.Debugf("Interval: %s\n", cmdConfig.Interval)
	log.Debugf("Bind: %s\n", cmdConfig.Bind)
}

// readConfig reads config.yaml from disk
func readConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
