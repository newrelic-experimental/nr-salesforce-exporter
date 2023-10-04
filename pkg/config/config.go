package config

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

// TODO: I don't love the exporter stuff being here as opposed to in the export
// package but it is hard to get rid of the circular import cycle in that case.
// For now, keep here.

type ExporterType string

const (
	NrMetrics  ExporterType = "nrmetrics"
	NrEvents   ExporterType = "nrevents"
	NrLogs     ExporterType = "nrlogs"
	NrTraces   ExporterType = "nrtraces"
	Otel       ExporterType = "otel"
	Prometheus ExporterType = "prom"
)

func (expor ExporterType) Check() bool {
	switch expor {
	case NrMetrics:
	case NrEvents:
	case NrLogs:
	case NrTraces:
	case Prometheus:
	case Otel:
	default:
		return false
	}
	return true
}

type Config struct {
	interval uint
	exporter ExporterType
}

func NewConfigWithFile(configFile string) (*Config, error) {
	viper.SetConfigFile(configFile)

	return newConfig()
}

func NewConfigWithPaths(paths []string) (*Config, error) {
	viper.SetConfigName("config")
	for _, path := range paths {
		viper.AddConfigPath(path)
	}
	viper.AddConfigPath(".")

	return newConfig()
}

func newConfig() (*Config, error) {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("NR_SFX")

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	interval := viper.GetUint("interval")
	if interval <= 0 {
		interval = 60
	}

	exp := viper.GetString("exporter")
	if exp == "" {
		return nil, errors.New("exporter must be specified")
	}

	exporter := ExporterType(exp)
	if !exporter.Check() {
		return nil, fmt.Errorf("invalid 'exporter' value %s in config", exporter)
	}

	return &Config{
		interval,
		exporter,
	}, nil
}

func (c *Config) Interval() uint {
	return c.interval
}

func (c *Config) Exporter() ExporterType {
	return c.exporter
}
