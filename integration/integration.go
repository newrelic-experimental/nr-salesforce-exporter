package integration

import (
	"newrelic/multienv/examples/integrations/newrelic/status"
	"newrelic/multienv/pkg/env"
	"newrelic/multienv/pkg/process"
	"newrelic/multienv/pkg/receive"
)

/*
// Integration Receiver Initializer
func InitRecv() (receive.Receiver, error) {
	return eventlogfile.InitRecv()
}

// Integration Processor Initializer
func InitProc() (process.Processor, error) {
	return eventlogfile.InitProc()
}
*/

// Integration Receiver Initializer
func InitRecv(env *env.Environment) (receive.Receiver, error) {
	return status.InitRecv(env)
}

// Integration Processor Initializer
func InitProc(env *env.Environment) (process.Processor, error) {
	return status.InitProc(env)
}

/*
// Integration Receiver Initializer
func InitRecv(pipeConfig *config.PipelineConfig) (receive.Receiver, error) {
	return hostnet.InitRecv(pipeConfig)
}

// Integration Processor Initializer
func InitProc(pipeConfig *config.PipelineConfig) (process.Processor, error) {
	return hostnet.InitProc(pipeConfig)
}
*/

/*
// Integration Receiver Initializer
func InitRecv(pipeConfig *config.PipelineConfig) (receive.Receiver, error) {
	return ipify.InitRecv(pipeConfig)
}

// Integration Processor Initializer
func InitProc(pipeConfig *config.PipelineConfig) (process.Processor, error) {
	return ipify.InitProc(pipeConfig)
}
*/
