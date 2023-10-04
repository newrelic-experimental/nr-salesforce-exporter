package export

import (
	"context"
	"newrelic/multienv/pkg/config"
	"newrelic/multienv/pkg/env"
	"newrelic/multienv/pkg/model"

	log "github.com/sirupsen/logrus"
)

type ExportFunc = func(
	ctx context.Context,
	env *env.Environment,
	data []model.MeltModel,
) error

func SelectExporter(exporterType config.ExporterType) ExportFunc {
	switch exporterType {
	case config.NrEvents:
		return exportNrEvent
	case config.NrMetrics:
		return exportNrMetric
	case config.NrLogs:
		return exportNrLog
	case config.NrTraces:
		return exportNrTrace
	//case config.NrInfra:
	//	return exportNrInfra
	case config.Otel:
		return exportOtel
	case config.Prometheus:
		return exportProm
	default:
		return dummyExporter
	}
}

func dummyExporter(
	ctx context.Context,
	env *env.Environment,
	data []model.MeltModel,
) error {
	log.Warn("Dummy Exporter, do nothing")
	log.Warn("    Data = ", data)
	return nil
}
