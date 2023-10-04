package standalone

import (
	"context"
	"newrelic/multienv/pkg/model"
	"newrelic/multienv/pkg/process"
	"sync"

	log "github.com/sirupsen/logrus"
)

type ProcWorkerConfig struct {
	Processor  process.Processor
	InChannel  <-chan []model.MeltModel
	OutChannel chan<- model.MeltModel
}

var procWorkerConfigHoldr SharedConfig[ProcWorkerConfig]

func InitProcessor(
	ctx context.Context,
	wg *sync.WaitGroup,
	config ProcWorkerConfig,
) {
	procWorkerConfigHoldr.SetConfig(config)

	if !procWorkerConfigHoldr.SetIsRunning() {
		if config.Processor == nil {
			log.Println("no processor defined; starting pipe worker...")
			wg.Add(1)
			go pipeWorker(ctx, wg)
			return
		}

		log.Println("starting processor worker...")
		wg.Add(1)
		go processorWorker(ctx, wg)

	} else {
		log.Println("processor worker already running, config updated.")
	}
}

func processorWorker(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	config := procWorkerConfigHoldr.Config()

	for {
		select {
		case <-ctx.Done():
			return // returning not to leak the goroutine

		case data := <-config.InChannel:
			err := config.Processor.Process(ctx, data)
			if err != nil {
				log.Errorf("processor error: %v", err)
				continue
			}

			writeData(&config, data)
		}
	}
}

func pipeWorker(ctx context.Context, wg *sync.WaitGroup) {
	config := procWorkerConfigHoldr.Config()

	for {
		select {
		case <-ctx.Done():
			return // returning not to leak the goroutine

		case data := <-config.InChannel:
			writeData(&config, data)
		}
	}
}

func writeData(config *ProcWorkerConfig, data []model.MeltModel) {
	for _, val := range data {
		config.OutChannel <- val
	}
}
