package standalone

import (
	"context"
	"fmt"
	"newrelic/multienv/integration"
	"newrelic/multienv/pkg/env"
	"newrelic/multienv/pkg/export"
	"newrelic/multienv/pkg/model"
	"newrelic/multienv/pkg/process"
	"newrelic/multienv/pkg/receive"
	"sync"

	"github.com/spf13/viper"
)

// Init data pipeline.
func InitPipeline(
	ctx context.Context,
	env *env.Environment,
	receiver receive.Receiver,
	processor process.Processor,
) {
	bufferSize := viper.GetInt("buffer")

	if bufferSize < 100 {
		bufferSize = 100
	}

	if bufferSize > 1000 {
		bufferSize = 1000
	}

	batchSize := viper.GetInt("batch_size")

	if batchSize < 10 {
		batchSize = 10
	}

	if batchSize > 100 {
		batchSize = 100
	}

	harvestTime := viper.GetInt("harvest_time")

	if harvestTime < 60 {
		harvestTime = 60
	}

	recvToProcCh := make(chan []model.MeltModel)
	procToExpCh := make(chan model.MeltModel)
	wg := &sync.WaitGroup{}

	InitExporter(
		ctx,
		wg,
		ExpWorkerConfig{
			InChannel:   procToExpCh,
			HarvestTime: harvestTime,
			BatchSize:   batchSize,
			Environment: env,
			Exporter:    export.SelectExporter(env.Config().Exporter()),
		},
	)

	InitProcessor(
		ctx,
		wg,
		ProcWorkerConfig{
			Processor:  processor,
			InChannel:  recvToProcCh,
			OutChannel: procToExpCh,
		},
	)

	InitReceiver(
		ctx,
		wg,
		RecvWorkerConfig{
			IntervalSec: env.Config().Interval(),
			Receiver:    receiver,
			OutChannel:  recvToProcCh,
		},
	)

	wg.Wait()
}

// Start Integration
func Start(
	ctx context.Context,
	env *env.Environment,
) error {
	receiver, err := integration.InitRecv(env)
	if err != nil {
		return fmt.Errorf("error initializing receiver: %w", err)
	}

	processor, err := integration.InitProc(env)
	if err != nil {
		return fmt.Errorf("error initializing processor: %w", err)
	}

	InitPipeline(ctx, env, receiver, processor)

	return nil
}
