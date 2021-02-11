package fact_go_client

import "github.com/faas-facts/fact/fact"

type Inspector interface {
	Name() string
	Init(trace *fact.Trace)
	Collect(trace fact.Trace, context interface{}) fact.Trace
}

type FactClient struct {
	base              fact.Trace
	trace             fact.Trace
	io                FactReceiver
	sendOnUpdate      bool
	platformInspector Inspector
}

type ReceiverType int

const (
	Console ReceiverType = iota
	TCP
)

type FactReceiver interface {
	Connect(map[string]string) error
	Send(trace fact.Trace) error
}

type FactClientConfig struct {
	Platform           *string
	Receiver           *ReceiverType
	SendOnUpdate       bool
	IncludeEnvironment bool
	IOArgs             map[string]string
}
