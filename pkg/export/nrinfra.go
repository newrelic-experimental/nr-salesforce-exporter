package export

import (
	"context"
	"fmt"
	"newrelic/multienv/pkg/env"
	"newrelic/multienv/pkg/model"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/newrelic/infra-integrations-sdk/v4/data/event"
	"github.com/newrelic/infra-integrations-sdk/v4/data/metric"
	"github.com/newrelic/infra-integrations-sdk/v4/integration"
)

func NrInfraExporter(
	i *integration.Integration,
	entity *integration.Entity,
) ExportFunc {
	return func(
		ctx context.Context,
		env *env.Environment,
		data []model.MeltModel,
	) error {
		return exportNrInfra(data, i, entity)
	}
}

func exportNrInfra(
	data []model.MeltModel,
	i *integration.Integration,
	entity *integration.Entity,
) error {
	log.Print("------> NR Infra Exporter = ", data)

	for _, d := range data {
		if d.Type == model.Event || d.Type == model.Log {
			ev, ok := d.Event()
			if ok {
				nriEv, err := event.New(time.UnixMilli(d.Timestamp), "Event of type "+ev.Type, ev.Type)
				nriEv.Attributes = d.Attributes
				if err != nil {
					log.Error("Error creating event", err)
				} else {
					entity.AddEvent(nriEv)
				}
			}
		} else if d.Type == model.Metric {
			m, ok := d.Metric()
			if ok {
				switch m.Type {
				case model.Gauge:
					gauge, err := integration.Gauge(time.UnixMilli(d.Timestamp), m.Name, m.Value.Float())
					addAttributes(&d, &gauge)
					if err != nil {
						log.Error("Error creating gauge metric", err)
					} else {
						entity.AddMetric(gauge)
					}
				case model.Count:
					//TODO: NO TIME INTERVAL???
					count, err := integration.Count(time.UnixMilli(d.Timestamp), m.Name, m.Value.Float())
					addAttributes(&d, &count)
					if err != nil {
						log.Error("Error creating count metric", err)
					} else {
						entity.AddMetric(count)
					}
				case model.Summary:
					//TODO
				}
			}
		} else {
			log.Warn("Ignored data, not a metric, event or log: ", d)
		}
	}

	return nil
}

func addAttributes(model *model.MeltModel, metric *metric.Metric) {
	for k, v := range model.Attributes {
		switch val := v.(type) {
		case string:
			(*metric).AddDimension(k, val)
		case int:
			(*metric).AddDimension(k, strconv.Itoa(val))
		case float32:
			(*metric).AddDimension(k, strconv.FormatFloat(float64(val), 'f', 2, 32))
		case float64:
			(*metric).AddDimension(k, strconv.FormatFloat(val, 'f', 2, 32))
		case fmt.Stringer:
			(*metric).AddDimension(k, val.String())
		default:
			log.Warn("Attribute of unsupported type: ", k, v)
		}

	}
}
