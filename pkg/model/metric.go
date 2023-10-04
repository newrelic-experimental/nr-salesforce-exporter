package model

import "time"

type MetricType int

const (
	Gauge MetricType = iota
	Count
	Summary
)

// Metric model variant.
type MetricModel struct {
	Name     string
	Type     MetricType
	Value    Numeric
	Interval time.Duration
	//TODO: summary metric data model
}

// Make a gauge metric.
func MakeGaugeMetric(name string, value Numeric, timestamp time.Time) MeltModel {
	return MeltModel{
		Type:      Metric,
		Timestamp: timestamp.UnixMilli(),
		Data: MetricModel{
			Name:  name,
			Type:  Gauge,
			Value: value,
		},
	}
}

// Make a count metric. If interval is 0 the counter is cumulative. Otherwise is delta.
func MakeCountMetric(name string, value Numeric, interval time.Duration, timestamp time.Time) MeltModel {
	return MeltModel{
		Type:      Metric,
		Timestamp: timestamp.UnixMilli(),
		Data: MetricModel{
			Name:     name,
			Type:     Count,
			Value:    value,
			Interval: interval,
		},
	}
}

//TODO: make summary metric
