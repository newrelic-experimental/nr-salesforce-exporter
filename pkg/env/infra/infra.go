package infra

import (
	"context"
	"fmt"
	multienv_integration "newrelic/multienv/integration"
	"newrelic/multienv/pkg/env"
	"newrelic/multienv/pkg/export"
	"newrelic/multienv/pkg/model"

	"github.com/newrelic/infra-integrations-sdk/v4/integration"
)

func Run(
	ctx context.Context,
	env *env.Environment,
	i *integration.Integration,
	entity *integration.Entity,
) error {
	receiver, err := multienv_integration.InitRecv(env)
	if err != nil {
		return fmt.Errorf("error initializing receiver: %w", err)
	}

	processor, err := multienv_integration.InitProc(env)
	if err != nil {
		return fmt.Errorf("error initializing processor: %w", err)
	}

	meltList := model.NewMeltList(500)

	err = receiver.Poll(ctx, meltList)
	if err != nil {
		return fmt.Errorf("error polling: %w", err)
	}

	err = processor.Process(ctx, meltList.Set)
	if err != nil {
		return fmt.Errorf("error processing: %w", err)
	}

	err = export.NrInfraExporter(i, entity)(ctx, env, meltList.Set)
	if err != nil {
		return fmt.Errorf("error exporting: %w", err)
	}

	return nil
}
