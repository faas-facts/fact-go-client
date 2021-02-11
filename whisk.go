package fact_go_client

import (
	"github.com/faas-facts/fact/fact"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"
)

type ICFInspector struct {
	OWInspector
}

func (O ICFInspector) Name() string {
	return "ICF"
}

func (O ICFInspector) Init(trace *fact.Trace) {
	O.OWInspector.Init(trace)
	uuid, err := ioutil.ReadFile("/sys/hypervisor/uuid")
	if err == nil {
		trace.HostID = strings.TrimSpace(string(uuid))
	}
}

func (O ICFInspector) Collect(trace fact.Trace, context interface{}) fact.Trace {
	t := O.GenericInspector.Collect(trace, context)
	t.Cost = float32(math.Floor(float64(t.Memory)/1024.0)) * 0.000017 * float32(t.ExecutionLatency.Seconds)
	return t
}

type OWInspector struct {
	GenericInspector
}

func (O OWInspector) Name() string {
	return "OW"
}

func (O OWInspector) Init(trace *fact.Trace) {
	trace.Platform = O.Name()
	trace.Region = os.Getenv("__OW_API_HOST")
	trace.Tags["uptime"] = strconv.FormatInt(uptime(), 10)
	trace.Tags["fname"] = os.Getenv("__OW_ACTION_NAME")

	hostname, _ := os.Hostname()
	trace.HostID = hostname

	bytes, err := ioutil.ReadFile("/sys/fs/cgroup/memory/memory.limit_in_bytes")
	if err == nil {
		memory, err := strconv.ParseInt(string(bytes), 10, 32)
		if err == nil {
			trace.Memory = int32(memory)
		}
	}

}
