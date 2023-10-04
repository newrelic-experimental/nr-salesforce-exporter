package main

import (
	"context"

	aws_lambda "github.com/aws/aws-lambda-go/lambda"
	log "github.com/sirupsen/logrus"

	"newrelic/multienv/integration"
	"newrelic/multienv/pkg/config"
	nri_lambda "newrelic/multienv/pkg/env/lambda"
	"newrelic/multienv/pkg/model"
	"newrelic/multienv/pkg/process"
)

var pipeConf config.PipelineConfig
var processor process.Processor
var initErr error

func init() {
	pipeConf = nri_lambda.LoadConfig()
	processor, initErr = integration.InitProc(pipeConf)
}

func HandleRequest(ctx context.Context, data []model.MeltModel) (any, error) {
	if initErr != nil {
		log.Errorf("error initializing: %v ", initErr)
		return nil, initErr
	}

	log.Print("Processor event received = ", data)

	processor.Process(ctx, data)

	log.Print("Sending to SQS processed data = ", data)
	return data, nil
}

func main() {
	aws_lambda.Start(HandleRequest)
}
