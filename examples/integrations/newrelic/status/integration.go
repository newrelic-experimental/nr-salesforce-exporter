package status

import (
	"fmt"
	"newrelic/multienv/pkg/env"
	melt "newrelic/multienv/pkg/model"
	"newrelic/multienv/pkg/process"
	"newrelic/multienv/pkg/receive"
	"time"
)

///// New Relic Status API example \\\\\

type nrStatus struct {
	Page struct {
		Id         string
		Name       string
		Url        string
		Time_zone  string
		Updated_at string
	}
	Status struct {
		Indicator   string
		Description string
	}
}

func InitRecv(env *env.Environment) (receive.Receiver, error) {
	return receive.NewSimpleReceiver(
		"https://status.newrelic.com/api/v2/status.json",
		nil,
		func() interface{} { return &nrStatus{} },
		transformNrStatus,
	), nil
}

func InitProc(env *env.Environment) (process.Processor, error) {
	return nil, nil
}

func transformNrStatus(model interface{}, set melt.MeltSink) error {
	nrStatus, ok := model.(*nrStatus)
	if !ok {
		return fmt.Errorf("unexpected model type looking for NR status")
	}

	mlog := melt.MakeLog(nrStatus.Status.Description, "NRStatus", time.Now())
	mlog.Attributes = map[string]any{
		"updatedAt": nrStatus.Page.Updated_at,
		"indicator": nrStatus.Status.Indicator,
		"id":        nrStatus.Page.Id,
	}

	set.Put(&mlog)

	return nil
}
