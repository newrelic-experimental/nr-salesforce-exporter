package eventlogfile

import (
	"fmt"
	"newrelic/multienv/pkg/connect"
	"newrelic/multienv/pkg/deser"
	"os"
	"regexp"
	"testing"
	"time"
)

var (
	queryUrlRE = *regexp.MustCompile(`https://newrelic\.com/services/data/v58\.0/query\?q=SELECT\+Id%2CEventType%2CCreatedDate%2CLogDate%2CInterval%2CLogFile%2CSequence\+FROM\+EventLogFile\+WHERE\+CreatedDate%3E%3D[\d]{4}-[\d]{2}-[\d]{2}T[\d]{2}%3A[\d]{2}%3A[\d]{2}\.[\d]{3}Z\+AND\+CreatedDate%3C[\d]{4}-[\d]{2}-[\d]{2}T[\d]{2}%3A[\d]{2}%3A[\d]{2}\.[\d]{3}Z\+AND\+Interval%3D%27Hourly%27$`)
)

type HttpConnectorStub struct {
	requestFunc func() ([]byte, error)
	setUrlFunc  func(string)
}

func (h *HttpConnectorStub) SetConfig(any)                                        {}
func (h *HttpConnectorStub) Request() ([]byte, error)                             { return h.requestFunc() }
func (h *HttpConnectorStub) ConnectorID() string                                  { return "HTTP" }
func (h *HttpConnectorStub) SetReqBuilder(builder connect.BuilderFunc)            {}
func (h *HttpConnectorStub) SetAuthenticator(authenticator connect.Authenticator) {}
func (h *HttpConnectorStub) SetMethod(method connect.HttpMethod)                  {}
func (h *HttpConnectorStub) SetUrl(url string)                                    { h.setUrlFunc(url) }
func (h *HttpConnectorStub) SetBody(body any)                                     {}
func (h *HttpConnectorStub) SetHeaders(headers map[string]string)                 {}
func (h *HttpConnectorStub) SetTimeout(timeout time.Duration)                     {}

func TestQuery(t *testing.T) {

	elfRecordData1, err := os.ReadFile("testdata/query_response_1.json")
	if err != nil {
		t.Errorf("missing query test data 1")
		return
	}

	elfRecordData2, err := os.ReadFile("testdata/query_response_2.json")
	if err != nil {
		t.Errorf("missing query test data 2")
		return
	}

	t.Run(
		"connector receives correct url",
		func(t *testing.T) {
			var actual string

			connector := &HttpConnectorStub{}
			connector.requestFunc = func() ([]byte, error) { return nil, fmt.Errorf("http_error") }
			connector.setUrlFunc = func(url string) { actual = url }

			api := NewEventLogFileApi(
				&ApiConfig{"https://newrelic.com", "v58.0"},
				connector,
			)

			query := "SELECT Id,EventType,CreatedDate,LogDate,Interval,LogFile,Sequence FROM EventLogFile WHERE CreatedDate>=2023-07-14T11:13:37.696Z AND CreatedDate<2023-07-14T15:18:37.696Z AND Interval='Hourly'"

			api.Query(query)

			if !queryUrlRE.Match([]byte(actual)) {
				t.Errorf("url does not match expected pattern: %s", actual)
				return
			}
		},
	)

	t.Run(
		"fails if connector.Request() does",
		func(t *testing.T) {
			connector := &HttpConnectorStub{}
			connector.requestFunc = func() ([]byte, error) { return nil, fmt.Errorf("http_error") }
			connector.setUrlFunc = func(url string) {}

			api := NewEventLogFileApi(
				&ApiConfig{"https://newrelic.com", "v58.0"},
				connector,
			)

			_, err := api.Query("SELECT Id FROM EventLogFile")
			if err == nil {
				t.Errorf("err was nil: expected http_error")
				return
			}

			if err.Error() != "http_error" {
				t.Errorf("unexpected error message: %s", err.Error())
				return
			}
		},
	)

	t.Run(
		"fails if deserialize does",
		func(t *testing.T) {
			connector := &HttpConnectorStub{}
			connector.requestFunc = func() ([]byte, error) { return elfRecordData2, nil }
			connector.setUrlFunc = func(url string) {}
			parseJsonOld := parseJson
			parseJson = func(data []byte, model interface{}) error { return fmt.Errorf("json_error") }

			api := NewEventLogFileApi(
				&ApiConfig{"https://newrelic.com", "v58.0"},
				connector,
			)

			_, err := api.Query("SELECT Id FROM EventLogFile")
			if err == nil {
				t.Errorf("err was nil: expected json_error")
				parseJson = parseJsonOld
				return
			}

			if err.Error() != "json_error" {
				t.Errorf("unexpected error message: %s", err.Error())
				parseJson = parseJsonOld
				return
			}

			parseJson = parseJsonOld
		},
	)

	t.Run(
		"connector.Request() returns correct data no paging",
		func(t *testing.T) {
			connector := &HttpConnectorStub{}
			connector.requestFunc = func() ([]byte, error) { return elfRecordData2, nil }
			connector.setUrlFunc = func(url string) {}

			api := NewEventLogFileApi(
				&ApiConfig{"https://newrelic.com", "v58.0"},
				connector,
			)

			data, err := api.Query("SELECT Id FROM EventLogFile")
			if err != nil {
				t.Errorf("unexpected error running query: %v", err)
				return
			}

			if len(data) != 3 {
				t.Errorf("unexpected length parsing records: %d != 3", len(data))
				return
			}

			record0 := data[0]
			record1 := data[1]

			if record0.Id != "0AT6t000004NdrcGAC" {
				t.Errorf("unexpected Id for record 0: %s != 0AT6t000004NdrcGAC", record0.Id)
				return
			}

			if record1.EventType != "FlowExecution" {
				t.Errorf("unexpected EventType for record 1: %s != FlowExecution", record1.EventType)
				return
			}
		},
	)

	t.Run(
		"connector.Request() returns correct data with paging",
		func(t *testing.T) {
			called := 0
			nextRecordUrl := ""
			connector := &HttpConnectorStub{}
			connector.requestFunc = func() ([]byte, error) {
				if called == 0 {
					called += 1
					return elfRecordData1, nil
				}
				return elfRecordData2, nil
			}

			connector.setUrlFunc = func(url string) {
				t.Log(url)
				if called == 1 {
					nextRecordUrl = url
				}
			}

			api := NewEventLogFileApi(
				&ApiConfig{"https://newrelic.com", "v58.0"},
				connector,
			)

			data, err := api.Query("SELECT Id FROM EventLogFile")
			if err != nil {
				t.Errorf("unexpected error running query: %v", err)
				return
			}

			if nextRecordUrl != "https://localdomain.test/more_records" {
				t.Errorf("unexpected value for nextRecordUrl: %s != https://localdomain.test/more_records", nextRecordUrl)
				return
			}

			if len(data) != 8 {
				t.Errorf("unexpected length parsing records: %d != 8", len(data))
				return
			}

			//record0 := data[0]
			record5 := data[5]

			if record5.Id != "0AT6t000004NdrcGAC" {
				t.Errorf("unexpected Id for record 0: %s != 0AT6t000004NdrcGAC", record5.Id)
				return
			}
		},
	)

}

func TestGetLogFile(t *testing.T) {

	elfRecordData1, err := os.ReadFile("testdata/query_response_1.json")
	if err != nil {
		t.Errorf("missing query test data 1")
		return
	}

	model := &eventLogFile{}

	err = deser.DeserJson(elfRecordData1, model)
	if err != nil {
		t.Errorf("error converting query test data 1: %v", err)
		return
	}

	logFileData1, err := os.ReadFile("testdata/get_log_file_response_1.csv")
	if err != nil {
		t.Errorf("missing get log file test data 1")
		return
	}

	t.Run(
		"returns nil if LogFile url is empty string",
		func(t *testing.T) {
			connector := &HttpConnectorStub{}
			connector.requestFunc = func() ([]byte, error) { return nil, fmt.Errorf("http_error") }
			connector.setUrlFunc = func(url string) {}

			api := NewEventLogFileApi(
				&ApiConfig{"https://newrelic.com", "v58.0"},
				connector,
			)

			record0 := model.Records[0]
			record0.LogFile = ""

			data, err := api.GetLogFile(&record0)

			if data != nil || err != nil {
				t.Errorf("expected nil for both data and err with empty LogFile: data = %s, err = %v", data, err)
				return
			}
		},
	)

	t.Run(
		"connector receives correct url",
		func(t *testing.T) {
			var actual string

			connector := &HttpConnectorStub{}
			connector.requestFunc = func() ([]byte, error) { return nil, fmt.Errorf("http_error") }
			connector.setUrlFunc = func(url string) { actual = url }

			api := NewEventLogFileApi(
				&ApiConfig{"https://newrelic.com", "v58.0"},
				connector,
			)

			record0 := &model.Records[0]
			api.GetLogFile(record0)

			if actual != "https://newrelic.com/services/data/v58.0/sobjects/EventLogFile/0AT6t000004Ndr3GAC/LogFile" {
				t.Errorf("unexpected url configuring collector: %s != https://newrelic.com/services/data/v58.0/sobjects/EventLogFile/0AT6t000004Ndr3GAC/LogFile", actual)
				return
			}
		},
	)

	t.Run(
		"fails if connector.Request() does",
		func(t *testing.T) {
			connector := &HttpConnectorStub{}
			connector.requestFunc = func() ([]byte, error) { return nil, fmt.Errorf("http_error") }
			connector.setUrlFunc = func(url string) {}

			api := NewEventLogFileApi(
				&ApiConfig{"https://newrelic.com", "v58.0"},
				connector,
			)

			record0 := &model.Records[0]

			_, err := api.GetLogFile(record0)
			if err == nil {
				t.Errorf("err was nil: expected http_error")
				return
			}

			if err.Error() != "http_error" {
				t.Errorf("unexpected error message: %s", err.Error())
				return
			}
		},
	)

	t.Run(
		"fails if deserialize does",
		func(t *testing.T) {
			connector := &HttpConnectorStub{}
			connector.requestFunc = func() ([]byte, error) { return logFileData1, nil }
			connector.setUrlFunc = func(url string) {}
			parseCsvOld := parseCsv
			parseCsv = func(b []byte) ([]map[string]string, error) { return nil, fmt.Errorf("csv_error") }

			api := NewEventLogFileApi(
				&ApiConfig{"https://newrelic.com", "v58.0"},
				connector,
			)

			record0 := &model.Records[0]

			_, err := api.GetLogFile(record0)
			if err == nil {
				t.Errorf("err was nil: expected csv_error")
				parseCsv = parseCsvOld
				return
			}

			if err.Error() != "csv_error" {
				t.Errorf("unexpected error message: %s", err.Error())
				parseCsv = parseCsvOld
				return
			}

			parseCsv = parseCsvOld
		},
	)

	t.Run(
		"connector.Request() returns correct data",
		func(t *testing.T) {
			connector := &HttpConnectorStub{}
			connector.requestFunc = func() ([]byte, error) { return logFileData1, nil }
			connector.setUrlFunc = func(url string) {}

			api := NewEventLogFileApi(
				&ApiConfig{"https://newrelic.com", "v58.0"},
				connector,
			)

			record0 := &model.Records[0]

			rows, err := api.GetLogFile(record0)
			if err != nil {
				t.Errorf("unexpected error retrieving log file: %v", err)
				return
			}

			if len(rows) != 45 {
				t.Errorf("unexpected length converting CSV to map: %d != 45", len(rows))
				return
			}

			eventType, ok := rows[0]["EVENT_TYPE"]
			if !ok {
				t.Errorf("missing EVENT_TYPE for row 1")
				return
			}

			if eventType != "ApexExecution" {
				t.Errorf("unexpected EVENT_TYPE for row 1: %s != ApexExecution", eventType)
				return
			}

			entryPoint, ok := rows[43]["ENTRY_POINT"]
			if !ok {
				t.Errorf("missing ENTRY_POINT for row 44")
				return
			}

			if entryPoint != "SupportHelperPublicController.getBannerStatus" {
				t.Errorf("unexpected ENTRY_POINT for row 44: %s != SupportHelperPublicController.getBannerStatus", entryPoint)
				return
			}
		},
	)
}
