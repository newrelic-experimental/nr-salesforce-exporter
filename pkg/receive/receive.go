package receive

import (
	"context"
	"newrelic/multienv/pkg/connect"
	"newrelic/multienv/pkg/deser"
	"newrelic/multienv/pkg/model"
	"time"
)

type (
	ModelFactoryFunc func() interface{}
	TransformerFunc  func(model interface{}, sink model.MeltSink) error
)

type Receiver interface {
	Poll(ctx context.Context, set model.MeltSink) error
}

type simpleReceiverOpt func(receiver *SimpleReceiver)

type SimpleReceiver struct {
	connector    connect.HttpConnector
	modelFactory ModelFactoryFunc
	deserializer deser.DeserFunc
	transformer  TransformerFunc
}

func NewSimpleReceiver(
	url string,
	headers map[string]string,
	modelFactory ModelFactoryFunc,
	transformer TransformerFunc,
	simpleReceiverOpts ...simpleReceiverOpt,
) *SimpleReceiver {
	receiver := &SimpleReceiver{
		connect.MakeHttpGetConnector(url, headers),
		modelFactory,
		deser.DeserJson,
		transformer,
	}

	for _, opt := range simpleReceiverOpts {
		opt(receiver)
	}

	return receiver
}

func (s *SimpleReceiver) Poll(
	ctx context.Context,
	sink model.MeltSink,
) error {
	data, err := s.connector.Request()
	if err != nil {
		return err
	}

	model := s.modelFactory()

	err = s.deserializer(data, model)
	if err != nil {
		return err
	}

	return s.transformer(model, sink)
}

func WithBuilder(builder connect.BuilderFunc) simpleReceiverOpt {
	return func(r *SimpleReceiver) {
		r.connector.SetReqBuilder(builder)
	}
}

func WithAuthenticator(authenticator connect.Authenticator) simpleReceiverOpt {
	return func(r *SimpleReceiver) {
		r.connector.SetAuthenticator(authenticator)
	}
}

func WithMethod(method connect.HttpMethod) simpleReceiverOpt {
	return func(r *SimpleReceiver) {
		r.connector.SetMethod(method)
	}
}

func WithBody(body any) simpleReceiverOpt {
	return func(r *SimpleReceiver) {
		r.connector.SetBody(body)
	}
}

func WithHeaders(headers map[string]string) simpleReceiverOpt {
	return func(r *SimpleReceiver) {
		r.connector.SetHeaders(headers)
	}
}

func WithTimeout(timeout time.Duration) simpleReceiverOpt {
	return func(r *SimpleReceiver) {
		r.connector.SetTimeout(timeout)
	}
}

type DataSource interface {
	Fetch() (interface{}, error)
}

type GenericReceiver struct {
	dataSource  DataSource
	transformer TransformerFunc
}

func NewGenericReceiver(
	dataSource DataSource,
	transformer TransformerFunc,
) *GenericReceiver {
	return &GenericReceiver{
		dataSource,
		transformer,
	}
}

func (r *GenericReceiver) Poll(
	ctx context.Context,
	sink model.MeltSink,
) error {
	model, err := r.dataSource.Fetch()
	if err != nil {
		return err
	}

	return r.transformer(model, sink)
}
