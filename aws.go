/*
 * Copyright (c) 2021. Sebastian Werner, TU Berlin, Germany
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package fact_go_client

import (
	"bytes"
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/faas-facts/fact/fact"
	"google.golang.org/protobuf/types/known/durationpb"

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
			if ver := _ctx.Value("FunctionVersion"); ver != nil {
				trace.Tags["fver"] = ver.(string)
			}
			trace.Tags["rid"] = lc.AwsRequestID
		}

		trace.Logs[uint64(time.Now().Unix())] = fmt.Sprintf("RenamingTime %s", time.Until(deadline))

		if val := _ctx.Value("MemoryLimitInMB"); val != nil {
			mem, _ := strconv.ParseInt(_ctx.Value("MemoryLimitInMB").(string), 10, 32)
			trace.Memory = int32(mem)
		}

	} else {
		log.Infof("%+v", ctx)
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
