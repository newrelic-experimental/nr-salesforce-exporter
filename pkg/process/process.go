package process

import (
	"context"
	"newrelic/multienv/pkg/model"
)

type Processor interface {
	Process(ctx context.Context, data []model.MeltModel) error
}

