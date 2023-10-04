package standalone

import (
	"context"
	"newrelic/multienv/pkg/env"
	"newrelic/multienv/pkg/export"
	"newrelic/multienv/pkg/model"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type ExpWorkerConfig struct {
	InChannel   <-chan model.MeltModel
	BatchSize   int
	HarvestTime int
	Environment *env.Environment
	Exporter    export.ExportFunc
}

var expWorkerConfig SharedConfig[ExpWorkerConfig]

func InitExporter(
	ctx context.Context,
	wg *sync.WaitGroup,
	config ExpWorkerConfig,
) {
	expWorkerConfig.SetConfig(config)
	if !expWorkerConfig.SetIsRunning() {
		log.Println("starting exporter worker...")
		wg.Add(1)
		go exporterWorker(ctx, wg)
	} else {
		log.Println("exporter worker already running, config updated.")
	}
}

func exporterWorker(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	config := expWorkerConfig.Config()

	ticker := time.NewTicker(time.Duration(config.HarvestTime) * time.Second)
	defer ticker.Stop()

	buffer := MakeReservoirBuffer[model.MeltModel](500)

	for {
		select {
		case <-ctx.Done():
			return // returning not to leak the goroutine

		case <-ticker.C:
			err := flush(ctx, &config, buffer)
			if err != nil {
				log.Error("exporter failed = ", err)
			}

		case data := <-config.InChannel:
			switch data.Type {
			case model.Metric:
				metric, _ := data.Metric()
				log.Println("exporter received a Metric", metric.Name)
			case model.Event:
				event, _ := data.Event()
				log.Println("exporter received an Event", event.Type)
			case model.Log:
				dlog, _ := data.Log()
				log.Println("exporter received a Log", dlog.Message, dlog.Type)
			case model.Trace:
				//TODO
				log.Warn("TODO: Exporter received a Trace")
			}

			buffer.Put(data)

			if buffer.Size() >= config.BatchSize {
				err := flush(ctx, &config, buffer)
				if err != nil {
					log.Error("exporter failed = ", err)
				}
			}
		}
	}
}

func flush(
	ctx context.Context,
	config *ExpWorkerConfig,
	buffer Buffer[model.MeltModel],
) error {
	bufSize := buffer.Size()
	buf := *buffer.Clear()

	log.Println("harvest cycle, buffer size = ", bufSize)

	err := config.Exporter(ctx, config.Environment, buf[0:bufSize])

	if err != nil {
		//TODO: handle error condition, refill buffer? Discard data? Retry?
		return err
	}

	return nil
}
