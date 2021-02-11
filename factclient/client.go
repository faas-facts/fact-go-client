package factclient

import (
	"fmt"
	"github.com/faas-facts/fact/fact"
	"github.com/google/uuid"
	"github.com/prometheus/common/log"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"os"
	"runtime"
	"strings"
	"time"
)

var containerID string

func init(){
	containerID = uuid.New().String()
}

func (fc *FactClient) Boot(conf FactClientConfig)  {
	switch conf.Receiver {
		//TODO:XXX
	}
	fc.base = fact.NewTrace()

	fc.base.BootTime= timestamppb.Now()
	fc.base.ContainerID= containerID
	fc.base.Runtime= fmt.Sprintf("%s %s %s",runtime.GOOS,runtime.GOARCH,runtime.Version())
	fc.base.Timestamp= timestamppb.Now()


	if conf.Platform != nil {
		fc.inspectorFromPlatformType(*conf.Platform)
	}

	if fc.platformInspector == nil {
		fc.inspectorFromEnvironment()
	}

	fc.platformInspector.Init(&fc.base)

	fc.sendOnUpdate = conf.SendOnUpdate

	if conf.IncludeEnvironment {
		log.Warn("includeEnvironment is set, this can leak sensetive information")
		for _, env := range os.Environ() {
			kv := strings.Split(env, "=")
			fc.trace.Env[kv[0]] = kv[1]
		}

	}

	fc.base = fc.platformInspector.Collect(fc.trace, nil)
}

func (fc *FactClient) inspectorFromPlatformType(pf string) {
	switch strings.ToUpper(pf) {
	case "AWS":
		fc.platformInspector = &AWSInspector{}
	case "GCF":
		fc.platformInspector = &GCFInspector{}
	case "ACF":
		fc.platformInspector = &ACFInspector{}
	case "ICF":
		fc.platformInspector = &ICFInspector{}
	case "OW":
		fc.platformInspector = &OWInspector{}
	}
}

func (fc *FactClient) Parent(parent string ){
	fc.trace.ChildOf = parent
}

func (fc *FactClient) Start(context interface{}, event interface{})  {
	trace := fc.platformInspector.Collect(fc.base,context)
	trace.ID = uuid.New().String()
	trace.StartTime = timestamppb.Now()

	if fc.sendOnUpdate {
		fc.send(trace)
	}

	fc.trace = trace
}

func (fc *FactClient) Update(context interface{}, msg *string, tags map[string]string)  {
	trace := fc.platformInspector.Collect(fc.trace,context)
	if msg != nil{
		trace.Logs[uint64(time.Now().Unix())] = *msg
	}
	for k, v := range tags {
		trace.Tags[k] = v
	}

	if fc.sendOnUpdate {
		fc.send(trace)
	}

	fc.trace = trace
}

func (fc *FactClient) Done(context interface{}, msg *string, args ... string) fact.Trace  {
	trace := fc.platformInspector.Collect(fc.trace,context)
	trace.EndTime = timestamppb.Now()

	if msg != nil{
		trace.Logs[uint64(time.Now().Unix())] = *msg
	}

	for _, arg := range args {
		trace.Args = append(trace.Args,arg)
	}
	trace.ExecutionLatency = durationpb.New(trace.EndTime.AsTime().Sub(trace.StartTime.AsTime()))

	fc.send(trace)

	return trace
}

func (fc *FactClient) send(trace fact.Trace) {
	if fc.io != nil{
		err := fc.io.Send(trace)
		if err != nil {
			log.Errorf("failed to send trace:%+v - %f",trace,err)
		}
	}
}