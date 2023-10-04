package export

import (
	"context"
	"newrelic/multienv/pkg/env"
	"newrelic/multienv/pkg/model"

	log "github.com/sirupsen/logrus"
)

func exportProm(
	ctx context.Context,
	env *env.Environment,
	data []model.MeltModel,
) error {
	log.Print("------> TODO: Prometheus Exporter = ", data)
	return nil
}
