package fact_go_client

import (
	"bytes"
	"context"
	"fmt"
	"github.com/faas-facts/fact/fact"
	"google.golang.org/protobuf/types/known/durationpb"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"
)

type AWSInspector struct {
	GenericInspector
	path string
}

func (A AWSInspector) Name() string {
	return "AWS"
}

func (A *AWSInspector) Init(trace *fact.Trace) {
	trace.Platform = A.Name()
	trace.ContainerID = os.Getenv("AWS_LAMBDA_LOG_STREAM_NAME")
	trace.Region = os.Getenv("AWS_REGION")

	uptime := uptime()
	trace.Tags["uptime"] = strconv.FormatInt(uptime, 10)

	err := A.readCGroupIDs(trace)
	if err != nil {
		trace.Tags["host"] = fmt.Sprintf("U%d", uptime)
		trace.HostID = fmt.Sprintf("U%d", uptime)
		trace.Tags["service"] = "undefined"
		trace.Tags["sandbox"] = "undefined"
		trace.Tags["freezer"] = "undefined"
	}

}

const freezerOffset = len("freezer:/sandbox-")

func (A *AWSInspector) readCGroupIDs(trace *fact.Trace) error {
	if A.path == "" {
		A.path = "/proc/self/cgroup"
	}

	data, err := ioutil.ReadFile(A.path)
	if err != nil {
		return err
	}

	lines := bytes.Split(data, []byte("\n"))
	found := 0
	for _, line := range lines {
		if index := bytes.Index(line, []byte("freezer")); index >= 0 {
			trace.Tags["freezer"] = string(line[index+freezerOffset:])
			found++
		}
		if index := bytes.Index(line, []byte("sandbox-root-")); index >= 0 {
			line = line[index:]

			if len(line) < 57 {
				return fmt.Errorf("sandbox line maleformed %s", string(line))
			}

			host := string(line[13:19])

			trace.Tags["host"] = host
			trace.HostID = host
			trace.Tags["service"] = string(line[36:42])
			trace.Tags["sandbox"] = string(line[51:57])
			found++
		}
		if found >= 2 {
			break
		}
	}

	return nil
}
func (A *AWSInspector) Collect(trace fact.Trace, ctx interface{}) fact.Trace {
	trace.ExecutionLatency = durationpb.New(time.Now().Sub(trace.StartTime.AsTime()))

	if _ctx, ok := ctx.(context.Context); ok {
		lc, _ := lambdacontext.FromContext(_ctx)
		deadline, _ := _ctx.Deadline()
		if lc != nil {
			trace.Tags["fname"] = lc.InvokedFunctionArn
			trace.Tags["fver"] = _ctx.Value("FunctionVersion").(string)
			trace.Tags["rid"] = lc.AwsRequestID
		}

		trace.Logs[uint64(time.Now().Unix())] = fmt.Sprintf("RenamingTime %s", time.Until(deadline))

		mem, _ := strconv.ParseInt(_ctx.Value("MemoryLimitInMB").(string), 10, 32)
		trace.Memory = int32(mem)
	}

	trace.Cost = A.calculateCost(trace)

	return trace
}

func (A AWSInspector) calculateCost(trace fact.Trace) float32 {
	mb := trace.Memory
	duration := float32(trace.ExecutionLatency.Seconds)
	if mb <= 128 {
		return 0.0000002083 * duration
	} else if mb <= 512 {
		return 0.0000008333 * duration
	} else if mb <= 1024 {
		return 0.0000016667 * duration
	} else if mb <= 1536 {
		return 0.0000025000 * duration
	} else if mb <= 2048 {
		return 0.0000033333 * duration
	} else {
		return 0.0000048958 * duration
	}
}
