package main

import (
	"context"

	aws_lambda "github.com/aws/aws-lambda-go/lambda"

	"newrelic/multienv/integration"
	"newrelic/multienv/pkg/config"
	nri_lambda "newrelic/multienv/pkg/env/lambda"
	"newrelic/multienv/pkg/model"
	"newrelic/multienv/pkg/receive"

	log "github.com/sirupsen/logrus"
)

var pipeConf config.PipelineConfig
var receiver receive.Receiver
var initErr error

func init() {
	pipeConf = nri_lambda.LoadConfig()
	receiver, initErr = integration.InitRecv(pipeConf)
}

// TODO: maybe we can get the scheduler rule from the context
// https://docs.aws.amazon.com/lambda/latest/dg/golang-context.html

func HandleRequest(ctx context.Context, event any) ([]model.MeltModel, error) {
	if initErr != nil {
		log.Errorf("error initializing: %v", initErr)
		return nil, initErr
	}

	log.Print("Event received: ", event)

	meltList := model.NewMeltList(500)

	err := receiver.Poll(ctx, meltList)
	if err != nil {
		log.Errorf("receiver poll failed: %v", err)
		return nil, err
	}

	return meltList.Set, nil
}

func main() {
	aws_lambda.Start(HandleRequest)
}
