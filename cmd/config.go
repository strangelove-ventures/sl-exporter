package cmd

import (
	"github.com/spf13/viper"
	"github.com/strangelove-ventures/sl-exporter/metrics"
)

type Config struct {
	File       string
	BindAddr   string
	NumWorkers int

	Static struct {
		Gauges []metrics.StaticGauge
	}

	Cosmos []metrics.CosmosChain
}

func parseConfig(cfg *Config) error {
	viper.SetConfigFile(cfg.File)
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	return viper.Unmarshal(cfg)
}
