package eventlogfile

import (
	"fmt"
	"net/url"
	"newrelic/multienv/pkg/connect"
	"newrelic/multienv/pkg/deser"
)

var (
	// Done for testing purposes only
	parseJson func([]byte, interface{}) error           = unmarshalJson
	parseCsv  func([]byte) ([]map[string]string, error) = unmarshalCsv
)

type EventLogFileRecord struct {
	Attributes struct {
		Type string
		Url  string
	}
	Id          string
	EventType   string
	CreatedDate string
	LogDate     string
	Interval    string
	LogFile     string
	Sequence    int
}

type eventLogFile struct {
	TotalSize      int
	Done           bool
	NextRecordsUrl string
	Records        []EventLogFileRecord
}

type ApiConfig struct {
	InstanceUrl string
	ApiVersion  string
}

type Api interface {
	Query(query string) ([]EventLogFileRecord, error)
	GetLogFile(record *EventLogFileRecord) ([]map[string]string, error)
}

type EventLogFileApi struct {
	config    *ApiConfig
	connector connect.HttpConnector
}

func NewEventLogFileApi(
	config *ApiConfig,
	connector connect.HttpConnector,
) Api {
	return &EventLogFileApi{
		config,
		connector,
	}
}

func (e *EventLogFileApi) Query(query string) ([]EventLogFileRecord, error) {
	url := makeQueryUrl(e.config, query)
	records := []EventLogFileRecord{}

	for done := false; !done; {
		model, err := e.query(url)
		if err != nil {
			return nil, err
		}

		records = append(records, model.Records...)
		done = model.Done

		if !done {
			url = model.NextRecordsUrl
		}
	}

	return records, nil
}

func (e *EventLogFileApi) GetLogFile(record *EventLogFileRecord) ([]map[string]string, error) {
	if record.LogFile == "" {
		return nil, nil
	}

	url := makeLogFileUrl(e.config, record)
	e.connector.SetUrl(url)

	data, err := e.connector.Request()
	if err != nil {
		return nil, err
	}

	rows, err := parseCsv(data)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (e *EventLogFileApi) query(url string) (*eventLogFile, error) {
	e.connector.SetUrl(url)

	data, err := e.connector.Request()
	if err != nil {
		return nil, err
	}

	model := &eventLogFile{}

	err = parseJson(data, model)
	if err != nil {
		return nil, err
	}

	return model, nil
}

func makeQueryUrl(config *ApiConfig, query string) string {
	return fmt.Sprintf(
		"%s/services/data/%s/query?q=%s",
		config.InstanceUrl,
		config.ApiVersion,
		url.QueryEscape(query),
	)
}

func makeLogFileUrl(config *ApiConfig, record *EventLogFileRecord) string {
	return fmt.Sprintf(
		"%s%s",
		config.InstanceUrl,
		record.LogFile,
	)
}

func unmarshalJson(data []byte, model interface{}) error {
	return deser.DeserJson(data, model)
}

func unmarshalCsv(data []byte) ([]map[string]string, error) {
	model := &deser.CsvRecords{}

	err := deser.DeserCsv(data, model)
	if err != nil {
		return nil, err
	}

	return model.AsMaps(), nil
}
