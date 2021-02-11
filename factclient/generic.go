package factclient

import (
	"fmt"
	"github.com/faas-facts/fact/fact"
	"google.golang.org/protobuf/types/known/durationpb"
	"time"
)

type GenericInspector struct {

}

func (g GenericInspector) Name() string {
	return "UKN"
}

func (g GenericInspector) Init(trace *fact.Trace) {
	trace.Platform = g.Name()
	uptime := uptime()
	trace.HostID = fmt.Sprintf("H_%d",uptime)
}

func (g GenericInspector) Collect(trace fact.Trace, context interface{}) fact.Trace {
	trace.ExecutionLatency = durationpb.New(time.Now().Sub(trace.StartTime.AsTime()))
	return trace
}

