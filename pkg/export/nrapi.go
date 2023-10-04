package export

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"newrelic/multienv/pkg/connect"
	"newrelic/multienv/pkg/env"
	"newrelic/multienv/pkg/model"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func exportNrEvent(
	ctx context.Context,
	env *env.Environment,
	data []model.MeltModel,
) error {
	log.Print("------> NR Event Exporter = ", data)

	jsonModel, err := intoNrEvent(data)
	if err != nil {
		log.Error("Error generating NR Event API data = ", err)
		return err
	}

	log.Print("NR Event JSON = ", string(jsonModel))

	return nrApiRequest(jsonModel, getEventEndpoint())
}

func exportNrMetric(
	ctx context.Context,
	env *env.Environment,
	data []model.MeltModel,
) error {
	log.Print("------> NR Metric Exporter = ", data)

	jsonModel, err := intoNrMetric(data)
	if err != nil {
		log.Error("Error generating NR Metric API data = ", err)
		return err
	}

	log.Print("NR Metric JSON = ", string(jsonModel))

	return nrApiRequest(jsonModel, getMetricEndpoint())
}

func exportNrLog(
	ctx context.Context,
	env *env.Environment,
	data []model.MeltModel,
) error {
	log.Print("------> NR Log Exporter = ", data)

	jsonModel, err := intoNrLog(data)
	if err != nil {
		log.Error("Error generating NR Log API data = ", err)
		return err
	}

	log.Print("NR Log JSON = ", string(jsonModel))

	return nrApiRequest(jsonModel, getLogEndpoint())
}

func exportNrTrace(
	ctx context.Context,
	env *env.Environment,
	data []model.MeltModel,
) error {
	log.Print("------> TODO: NR Trace Exporter = ", data)
	return nil
}

func getEndpoint() string {
	endpoint := viper.GetString("nr_endpoint")
	if endpoint == "" {
		endpoint = "US"
	} else {
		if endpoint != "US" && endpoint != "EU" {
			endpoint = "US"
		}
	}
	return endpoint
}

func getEventEndpoint() string {
	accountId, err := requireAccountID()
	if err != nil {
		log.Error(err)
		return ""
	}

	if getEndpoint() == "US" {
		return "https://insights-collector.newrelic.com/v1/accounts/" + accountId + "/events"
	} else {
		return "https://insights-collector.eu01.nr-data.net/v1/accounts/" + accountId + "/events"
	}
}

func getMetricEndpoint() string {
	if getEndpoint() == "US" {
		return "https://metric-api.newrelic.com/metric/v1"
	} else {
		return "https://metric-api.eu.newrelic.com/metric/v1"
	}
}

func getLogEndpoint() string {
	if getEndpoint() == "US" {
		return "https://log-api.newrelic.com/log/v1"
	} else {
		return "https://log-api.eu.newrelic.com/log/v1"
	}
}

func intoNrEvent(meltData []model.MeltModel) ([]byte, error) {
	events := make([]map[string]any, 0)
	for _, element := range meltData {
		event, ok := element.Event()
		if ok {
			nrevent := map[string]any{
				"eventType": event.Type,
				"timestamp": element.Timestamp,
			}
			if element.Attributes != nil {
				for k, v := range element.Attributes {
					if k != "eventType" && k != "timestamp" {
						nrevent[k] = v
					}
				}
			}
			events = append(events, nrevent)
		}
	}

	return json.Marshal(events)
}

func intoNrMetric(meltData []model.MeltModel) ([]byte, error) {
	metrics := make([]map[string]any, 0)
	for _, element := range meltData {
		metric, ok := element.Metric()
		if ok {
			var nrmetric map[string]any

			switch metric.Type {
			case model.Gauge:
				nrmetric = map[string]any{
					"name":      metric.Name,
					"type":      "gauge",
					"value":     metric.Value.Value(),
					"timestamp": element.Timestamp,
				}
				if element.Attributes != nil {
					nrmetric["attributes"] = element.Attributes
				}
			case model.Count:
				nrmetric = map[string]any{
					"name":        metric.Name,
					"type":        "count",
					"value":       metric.Value.Value(),
					"interval.ms": metric.Interval.Milliseconds(),
					"timestamp":   element.Timestamp,
				}
				if element.Attributes != nil {
					nrmetric["attributes"] = element.Attributes
				}
			case model.Summary:
				//TODO: implement summary metrics
			default:
				// Skip this metric
				continue
			}

			if element.Attributes != nil {
				nrmetric["attributes"] = element.Attributes
			}
			metrics = append(metrics, nrmetric)
		}
	}

	metricModel := []any{
		map[string]any{
			"metrics": metrics,
		},
	}

	return json.Marshal(metricModel)
}

func intoNrLog(meltData []model.MeltModel) ([]byte, error) {
	logs := make([]map[string]any, 0)
	for _, element := range meltData {
		log, ok := element.Log()
		if ok {
			nrlog := map[string]any{
				"message":   log.Message,
				"timestamp": element.Timestamp,
				"logtype":   log.Type,
			}
			if element.Attributes != nil {
				nrlog["attributes"] = element.Attributes
			}
			logs = append(logs, nrlog)
		}
	}

	logModel := []any{
		map[string]any{
			"logs": logs,
		},
	}

	return json.Marshal(logModel)
}

func gzipString(inputData string) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write([]byte(inputData)); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func nrApiRequest(jsonModel []byte, endpoint string) error {
	apiKey, err := requireLicenseKey()
	if err != nil {
		return err
	}

	headers := map[string]string{
		"Api-Key":          apiKey,
		"Content-Type":     "application/json",
		"Content-Encoding": "gzip",
	}

	gzipBody, errGzip := gzipString(string(jsonModel))
	if errGzip != nil {
		log.Error("Error compressing body = ", errGzip)
		return errGzip
	}

	connector := connect.MakeHttpPostConnector(endpoint, gzipBody, headers)
	response, errReq := connector.Request()
	if errReq != nil {
		log.Error("Error sending request to NR API = ", errReq)
		return errReq
	}

	log.Print("Response from NR API = ", string(response))

	return nil
}

func requireAccountID() (string, error) {
	accountId := os.Getenv("NEW_RELIC_ACCOUNT_ID")
	if accountId == "" {
		accountId = viper.GetString("nr_account_id")
		if accountId == "" {
			return "", fmt.Errorf("'nr_account_id' not found in the pipeline config")
		}
	}

	return accountId, nil
}

func requireLicenseKey() (string, error) {
	licenseKey := os.Getenv("NEW_RELIC_LICENSE_KEY")

	if licenseKey == "" {
		licenseKey = viper.GetString("nr_license_key")
		if licenseKey == "" {
			return "", fmt.Errorf("'nr_license_key' not found in the pipeline config")
		}
	}

	return licenseKey, nil
}
