package cmd

import (
	"github.com/spf13/viper"
	"github.com/strangelove-ventures/sl-exporter/metrics"
)

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

type Config struct {
	File     string
	BindAddr string

	Static struct {
		Gauges []metrics.StaticGauge
	}

	// Deprecated
	Metrics map[string]Metric
}

func parseConfig(cfg *Config) error {
	viper.SetConfigFile(cfg.File)
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	return viper.Unmarshal(cfg)
}
