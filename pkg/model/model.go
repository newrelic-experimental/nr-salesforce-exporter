package model

import (
	"encoding/json"
	"errors"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/mitchellh/mapstructure"
)

type MeltType int

const (
	Metric MeltType = iota
	Event
	Log
	Trace
)

// Intermediate model.
type MeltModel struct {
	Type MeltType
	// Unix timestamp in millis.
	Timestamp  int64
	Attributes map[string]any
	// Either a MetricModel, EventModel, LogModel or TraceModel.
	Data any
}

type MeltSink interface {
	Put(melt *MeltModel)
}

// Non-thread safe
type MeltList struct {
	Set			[]MeltModel
}

type MeltWriter struct {
	capacity	int
	buff     	[]MeltModel
	out			chan<- []MeltModel
	mu			sync.Mutex
}

func (receiver *MeltModel) UnmarshalJSON(data []byte) error {
	var dict map[string]any
	err := json.Unmarshal(data, &dict)
	if err != nil {
		return err
	}

	var model MeltModel
	err = mapstructure.Decode(dict, &model)
	if err != nil {
		return err
	}

	meltData := model.Data.(map[string]any)
	switch model.Type {
	case Metric:
		var metricModel MetricModel
		err := mapstructure.Decode(meltData, &metricModel)
		if err != nil {
			return err
		}
		model.Data = metricModel
	case Event:
		var eventModel EventModel
		err := mapstructure.Decode(meltData, &eventModel)
		if err != nil {
			return err
		}
		model.Data = eventModel
	case Log:
		var logModel LogModel
		err := mapstructure.Decode(meltData, &logModel)
		if err != nil {
			return err
		}
		model.Data = logModel
	case Trace:
		//TODO: unmarshal Trace model
	default:
		return errors.New("'Type' contains an invalid value " + strconv.Itoa(int(model.Type)))
	}

	*receiver = model
	return nil
}

func (m *MeltModel) Metric() (MetricModel, bool) {
	model, ok := m.Data.(MetricModel)
	return model, ok
}

// Event obtains an EventModel from the MeltModel.
// If the inner data is a LogModel, it will be converted into an EventModel.
// This transformation may cause data loss: if the Log had a key in `Attributes` named "message", it will be overwritten
// with the contents of the `Message` field.
func (m *MeltModel) Event() (EventModel, bool) {
	if m.Type == Log {
		// Convert Log into an Event
		logModel := m.Data.(LogModel)
		model := EventModel{
			Type: logModel.Type,
		}
		if m.Attributes == nil {
			m.Attributes = map[string]any{}
		}
		// Warning: if the log already had an attribute named "message", it will be overwritten
		if _, ok := m.Attributes["message"]; ok {
			log.Warn("Log2Event: Log already had an attribute named 'message', overwriting it")
		}
		m.Attributes["message"] = logModel.Message
		return model, true
	} else {
		model, ok := m.Data.(EventModel)
		return model, ok
	}
}

func (m *MeltModel) Log() (LogModel, bool) {
	// TODO: convert Event to Log -> check if the event contains an attribute named "message" of type string.
	model, ok := m.Data.(LogModel)
	return model, ok
}

func (m *MeltModel) Trace() (TraceModel, bool) {
	model, ok := m.Data.(TraceModel)
	return model, ok
}

// Numeric model.
type Numeric struct {
	IntOrFlt bool // true = Int, false = Float
	IntVal   int64
	FltVal   float64
}

// Numeric holds an integer.
func (n *Numeric) IsInt() bool {
	return n.IntOrFlt
}

// Numeric holds a float.
func (n *Numeric) IsFloat() bool {
	return !n.IntOrFlt
}

// Get float from Numeric.
func (n *Numeric) Float() float64 {
	if n.IsFloat() {
		return n.FltVal
	} else {
		return float64(n.IntVal)
	}
}

// Get int from Numeric.
func (n *Numeric) Int() int64 {
	if n.IsInt() {
		return n.IntVal
	} else {
		return int64(n.FltVal)
	}
}

// Get whatever it is.
func (n *Numeric) Value() any {
	if n.IsInt() {
		return n.IntVal
	} else {
		return n.FltVal
	}
}

// Make a Numeric from either an int, int64 or float64. Other types will cause a panic.
func MakeNumeric(val any) Numeric {
	switch val := val.(type) {
	case int:
		return Numeric{
			IntOrFlt: true,
			IntVal:   int64(val),
			FltVal:   0.0,
		}
	case int64:
		return Numeric{
			IntOrFlt: true,
			IntVal:   val,
			FltVal:   0.0,
		}
	case float64:
		return Numeric{
			IntOrFlt: false,
			IntVal:   0,
			FltVal:   val,
		}
	default:
		panic("Numeric must be either an int, int64 or float64")
	}
}

func NewMeltList(capacity int) *MeltList {
	return &MeltList{
		make([]MeltModel, 0, capacity),
	}
}

func (m *MeltList) Put(melt *MeltModel) {
	m.Set = append(m.Set, *melt)
}

func NewMeltWriter(capacity int, out chan<- []MeltModel) *MeltWriter {
	return &MeltWriter{
		capacity: capacity,
		buff: make([]MeltModel, 0, capacity),
		out: out,
	}
}

func (ms *MeltWriter) Put(melt *MeltModel) {
	ms.mu.Lock()
	if len(ms.buff) >= ms.capacity {
		ms.flush()
	}
	ms.buff = append(ms.buff, *melt)
	ms.mu.Unlock()
}

func (ms *MeltWriter) Flush() {
	ms.mu.Lock()
	ms.flush()
	ms.mu.Unlock()
}

func (ms *MeltWriter) flush() {
	ms.out <- ms.buff
	ms.buff = make([]MeltModel, 0, ms.capacity)
}