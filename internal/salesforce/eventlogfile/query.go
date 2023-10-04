package eventlogfile

import (
	"fmt"
	"newrelic/multienv/pkg/config"
	"newrelic/multienv/pkg/connect"
	"newrelic/multienv/pkg/receive"
	"strings"
	"time"
)

const (
	kQueryTemplate  = `SELECT Id,EventType,CreatedDate,LogDate,Interval,LogFile,Sequence FROM EventLogFile WHERE CreatedDate>=%s AND Interval='Hourly'`
	kDateTimeFormat = "2006-01-02T15:04:05.000Z"
)

type EventLogs struct {
	Record *EventLogFileRecord
	Logs   []map[string]string
}

type EventLogFileQueryConfig struct {
	InstanceUrl string
	ApiVersion  string
	interval    uint
}

type EventLogFileDataSource struct {
	eventLogFileApi Api
	queryConfig     *EventLogFileQueryConfig
	from            time.Time
}

func NewEventLogFileQueryConfigFromConfig(
	pipeConfig *config.PipelineConfig,
) (*EventLogFileQueryConfig, error) {
	instanceUrl, ok := pipeConfig.GetString("instanceUrl")
	if !ok {
		return nil, fmt.Errorf("missing instance URL in config")
	}

	apiVersion, ok := pipeConfig.GetString("apiVersion")
	if !ok {
		return nil, fmt.Errorf("missing api version in config")
	}

	return &EventLogFileQueryConfig{
		instanceUrl,
		apiVersion,
		pipeConfig.Interval,
	}, nil
}

func NewEventLogFileDataSource(
	config *EventLogFileQueryConfig,
	connector connect.HttpConnector,
) (receive.DataSource, error) {
	from := time.Now().
		UTC().
		Add(-config.Offset).
		Add(-(time.Duration(config.interval) * time.Second))

	return &EventLogFileDataSource{
		NewEventLogFileApi(
			&ApiConfig{
				config.InstanceUrl,
				config.ApiVersion,
			},
			connector,
		),
		config,
		from,
	}, nil
}

func (ds *EventLogFileDataSource) Fetch() (interface{}, error) {
	_, err := ds.eventLogFileApi.Query(makeQuery(
		ds.from,
		ds.queryConfig.Offset,
		ds.queryConfig.DateField,
		ds.queryConfig.GenerationInterval,
	))
	if err != nil {
		return nil, err
	}

	// This is only really necessary in the standalone case since lambda and
	// infra are one fetch per execution and no state is maintained.

	ds.from = time.Now().UTC().Add(-ds.queryConfig.Offset)

	return nil, nil
}

/*
func (ds *eventLogFileDataSource) fetchAllRecords() (*elfQueryResults, error) {
	elfResults := &elfQueryResults{}

	// First grab all the log file entries
	url := ds.makeQueryUrl()

	for done := false; !done; {
		model, err := ds.fetchRecords(url)
		if err != nil {
			return nil, err
		}

		elfResults.TotalSize = model.TotalSize
		elfResults.Records = append(elfResults.Records, model.Records...)
		done = model.Done

		if !done {
			url = model.NextRecordsUrl
		}
	}

	return elfResults, nil
}

func (ds *eventLogFileDataSource) fetchRecords(url string) (*eventLogFile, error) {
	ds.connector.SetUrl(url)

	data, err := ds.connector.Request()
	if err != nil {
		return nil, err
	}

	model := &eventLogFile{}

	err = ds.deser(data, model)
	if err != nil {
		return nil, err
	}

	return model, nil
}

func (ds *eventLogFileDataSource) fetchLogs(
	records *elfQueryResults,
) (*elfLogEntries, error) {
	for _, record := range records.Records {
		url := fmt.Sprintf("%s/%s", ds.instanceUrl, record.LogFile)

		ds.connector.SetUrl(url)

		data, err := ds.connector.Request()
		if err != nil {
			return nil, err
		}

		model := &deser.CsvRecords{}

		err = deser.DeserCsv(data, model)
		if err != nil {
			return nil, err
		}
	}
}
*/

func makeQuery(
	from time.Time,
	offset time.Duration,
	dateField dateField,
	generationInterval generationInterval,
) string {
	to := time.Now().UTC().Add(offset)

	return fmt.Sprintf(
		kQueryTemplate,
		dateField,
		from.Format(kDateTimeFormat),
		to.Format(kDateTimeFormat),
		generationInterval,
	)
}

func parseDateField(pipeConfig *config.PipelineConfig) (dateField, error) {
	dateField, ok := pipeConfig.GetString("dateField")
	if !ok {
		return "", fmt.Errorf("missing query date field in config")
	}

	if strings.EqualFold(dateField, string(kCreatedDate)) {
		return kCreatedDate, nil
	}

	if strings.EqualFold(dateField, string(kLogDate)) {
		return kLogDate, nil
	}

	return "", fmt.Errorf("invalid query date field in config")
}

func parseGenerationInterval(pipeConfig *config.PipelineConfig) (generationInterval, error) {
	generationInterval, ok := pipeConfig.GetString("generationInterval")
	if !ok {
		return "", fmt.Errorf("missing generation interval in config")
	}

	if strings.EqualFold(generationInterval, string(kDaily)) {
		return kDaily, nil
	}

	if strings.EqualFold(generationInterval, string(kHourly)) {
		return kHourly, nil
	}

	return "", fmt.Errorf("invalid generation interval in config")
}
