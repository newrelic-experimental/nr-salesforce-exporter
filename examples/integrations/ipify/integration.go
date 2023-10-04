package ipify

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"newrelic/multienv/pkg/config"
	melt "newrelic/multienv/pkg/model"
	"newrelic/multienv/pkg/process"
	"newrelic/multienv/pkg/receive"
)

///// Ipify API example \\\\\

type ipify struct {
	IpAddress string `json:"ip"`
}

func InitRecv(pipeConfig *config.PipelineConfig) (receive.Receiver, error) {
	return receive.NewSimpleReceiver(
		"https://api.ipify.org/?format=json",
		nil,
		func() interface{} { return &ipify{} },
		transformIpifyResponse,
	),
	nil
}

/*

func InitRecvWithReqBuilder(pipeConfig config.PipelineConfig) (config.RecvConfig, error) {
	connector := connect.MakeHttpConnectorWithBuilder(requestBuilder)

	return config.RecvConfig{
		Connector: &connector,
		Deser:     deser.DeserJson,
	}, nil
}

// Custom request builder. Not necessary, only to show how it works.
func requestBuilder(conf *connect.HttpConfig) (*http.Request, error) {
	req, err := http.NewRequest("GET", "https://api.ipify.org/?format=json", nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}

*/

func InitProc(pipeConfig *config.PipelineConfig) (process.Processor, error) {
	return nil, nil
}

// Processor function
func transformIpifyResponse(model interface{}, set melt.MeltSink) error {
	ipify, ok := model.(*ipify)
	if !ok {
		return fmt.Errorf("unknown type for data")
	}

	log.Println("My IP is = " + ipify.IpAddress)
	
	mlog := melt.MakeLog(ipify.IpAddress, "IPAddress", time.Now())
	mlog.Attributes = map[string]any{"type": "ip"}
	
	set.Put(&mlog)

	return nil
}
