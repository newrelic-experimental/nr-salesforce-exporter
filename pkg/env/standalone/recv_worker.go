package standalone

import (
	"context"
	"sync"
	"time"

	"newrelic/multienv/pkg/model"
	"newrelic/multienv/pkg/receive"

	log "github.com/sirupsen/logrus"
)

type RecvWorkerConfig struct {
	IntervalSec uint
	Receiver    receive.Receiver
	OutChannel  chan<- []model.MeltModel
}

var recvWorkerConfigHoldr SharedConfig[RecvWorkerConfig]

func InitReceiver(
	ctx context.Context,
	wg *sync.WaitGroup,
	config RecvWorkerConfig) {
	recvWorkerConfigHoldr.SetConfig(config)
	if !recvWorkerConfigHoldr.SetIsRunning() {
		log.Println("starting receiver worker...")
		wg.Add(1)
		go receiverWorker(ctx, wg)
	} else {
		log.Println("receiver worker already running, config updated.")
	}
}

func receiverWorker(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	config := recvWorkerConfigHoldr.Config()
	meltWriter := model.NewMeltWriter(500, config.OutChannel)

	// Poll on startup
	poll(ctx, &config, meltWriter)

	// Setup ticker for subsequent polling
	ticker := time.NewTicker(time.Duration(config.IntervalSec) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return // returning not to leak the goroutine

		case <-ticker.C:
			poll(ctx, &config, meltWriter)
		}
	}
}

func poll(
	ctx context.Context,
	config *RecvWorkerConfig,
	meltWriter *model.MeltWriter,
) {
	err := config.Receiver.Poll(ctx, meltWriter)
	if err != nil {
		// Warning only as we this could be just one iteration and we want to
		// keep executing. May want to have a threshold of consecutive errors
		// at which we then bail?
		log.Warnf("receiver poll error: %v", err)
		return
	}

	meltWriter.Flush()
}
