package export

import (
	"context"
	"newrelic/multienv/pkg/env"
	"newrelic/multienv/pkg/model"

	log "github.com/sirupsen/logrus"
)

func exportOtel(
	ctx context.Context,
	env *env.Environment,
	data []model.MeltModel,
) error {
	log.Print("------> TODO: OpenTelemetry Exporter = ", data)
	return nil
}
