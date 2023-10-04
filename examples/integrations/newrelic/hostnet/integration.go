package hostnet

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"newrelic/multienv/pkg/config"
	"newrelic/multienv/pkg/connect"
	melt "newrelic/multienv/pkg/model"
	"newrelic/multienv/pkg/process"
	"newrelic/multienv/pkg/receive"
)

///// New Relic GraphQL example, avarage host.net.receiveBytesPerSecond metric \\\\\

type nerdGraphResp struct {
	Data struct {
		Actor struct {
			Account struct {
				Nrql struct {
					Results []struct {
						Value float64 `json:"average.host.net.receiveBytesPerSecond"`
					}
				}
			}
		}
	}
}

type nrCred struct {
	AccountID string
	UserKey   string
}

// No danger of data races because it's only set on Init and after that only read
var recv_interval = 0

func InitRecv(pipeConfig *config.PipelineConfig) (receive.Receiver, error) {
	recv_interval = int(pipeConfig.Interval)
	if recv_interval == 0 {
		log.Warn("NR Graph QL: Interval not set, using 5 seconds")
		recv_interval = 5
	}

	var nrCred nrCred

	accountId, err := requireAccountID(pipeConfig)
	if err != nil {
		return nil, err
	}

	apiKey, err := requireApiKey(pipeConfig)
	if err != nil {
		return nil, err
	}

	nrCred.AccountID = accountId
	nrCred.UserKey = apiKey
	
	url := "https://api.newrelic.com/graphql"
	query := "SELECT average(host.net.receiveBytesPerSecond) FROM Metric SINCE " + strconv.Itoa(recv_interval) + " seconds AGO"
	body := fmt.Sprintf(`{
		actor { account(id: %s) 
		{ nrql
		(query: "%s")
		{ results } } } 
	}`, nrCred.AccountID, query)

	headers := map[string]string{"API-Key": nrCred.UserKey}

	return receive.NewSimpleReceiver(
		url,
		headers,
		func() interface{} { return &nerdGraphResp{} },
		transformNerdGraphResponse,
		receive.WithMethod(connect.Post),
		receive.WithBody(body),
		receive.WithTimeout(10 * time.Second),
	),
	nil
}

func InitProc(pipeConfig *config.PipelineConfig) (process.Processor, error) {
	return nil, nil
}

// Processor function
func transformNerdGraphResponse(model interface{}, set melt.MeltSink) error {
	nrdata, ok := model.(*nerdGraphResp)
	if !ok {
		return fmt.Errorf("unexpected model type looking for nerdgraph response")
	}
	
	if len(nrdata.Data.Actor.Account.Nrql.Results) == 0 {
		return nil
	}

	log.Printf("NR value received = %v\n", nrdata.Data.Actor.Account.Nrql.Results[0].Value)

	val := melt.MakeNumeric(int(nrdata.Data.Actor.Account.Nrql.Results[0].Value))
	interval := time.Duration(recv_interval) * time.Second
	
	// Making a count out of an avarage rate doesn't make much sense. This is just to test count metrics
	countMetric := melt.MakeCountMetric("nr.test.AvrgBytesPerSec", val, interval, time.Now())
	gaugeMetric := melt.MakeGaugeMetric("nr.test.Random", melt.MakeNumeric(rand.Float64()), time.Now())
	gaugeMetric.Attributes = map[string]any{"test_key": "test_value"}

	set.Put(&countMetric)
	set.Put(&gaugeMetric)

	return nil
}

func requireAccountID(pipeConf *config.PipelineConfig) (string, error) {
	var ok bool
	
	accountId := os.Getenv("NEW_RELIC_ACCOUNT_ID")
	if accountId == "" {
		accountId, ok = pipeConf.GetString("nr_account_id")
		if !ok {
			return "", fmt.Errorf("'nr_account_id' not found in the pipeline config")
		}
	}

	return accountId, nil
}

func requireApiKey(pipeConf *config.PipelineConfig) (string, error) {
	var ok bool

	apiKey := os.Getenv("NEW_RELIC_API_KEY")

	if apiKey == "" {
		apiKey, ok = pipeConf.GetString("nr_api_key")
		if !ok {
			return "", fmt.Errorf("'nr_api_key' not found in the pipeline config")
		}
	}

	return apiKey, nil
}