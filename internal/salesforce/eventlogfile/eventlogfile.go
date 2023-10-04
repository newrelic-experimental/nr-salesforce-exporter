package eventlogfile

import (
	"fmt"

	"newrelic/multienv/pkg/config"
	"newrelic/multienv/pkg/connect"
	"newrelic/multienv/pkg/model"
	"newrelic/multienv/pkg/process"
	"newrelic/multienv/pkg/receive"
)

func InitRecv(pipeConfig *config.PipelineConfig) (receive.Receiver, error) {
	eventLogFileQueryConfig, err := NewEventLogFileQueryConfigFromConfig(pipeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create query config: %w", err)
	}

	connector, err := newHttpConnector(pipeConfig, eventLogFileQueryConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connector: %w", err)
	}

	ds, err := NewEventLogFileDataSource(eventLogFileQueryConfig, connector)
	if err != nil {
		return nil, fmt.Errorf("failed to create data source: %w", err)
	}

	return receive.NewGenericReceiver(ds, transformElfRecords), nil
}

func InitProc(pipeConfig *config.PipelineConfig) (process.Processor, error) {
	return nil, nil
}

func transformElfRecords(model interface{}, set model.MeltSink) error {
	/*
		elfResults, ok := model.(*elfResultSet)
		if !ok {
			return fmt.Errorf("unexpected model type looking for elf results")
		}

		fmt.Printf("%v", elfResults)
	*/
	return nil
}

func newHttpConnector(
	pipeConfig *config.PipelineConfig,
	queryConfig *EventLogFileQueryConfig,
) (connect.HttpConnector, error) {
	authenticator, err := NewAuthenticatorFromConfig(pipeConfig, queryConfig.InstanceUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create salesforce authenticator: %w", err)
	}

	connector := connect.MakeHttpGetConnector("", nil)

	connector.SetAuthenticator(authenticator)

	return connector, nil
}
