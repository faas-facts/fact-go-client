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
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/faas-facts/fact/fact"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var containerID string

func init() {
	containerID = uuid.New().String()
}

func (fc *FactClient) Boot(conf FactClientConfig) {
	if conf.Receiver != nil {
		switch *conf.Receiver {
		case Console:
			fc.io = &ConsoleLogger{}
		case TCP:
			fc.io = &TCPLogger{}
		}
		if fc.io == nil {
			fc.io = &ConsoleLogger{}
		}
	} else {
		fc.io = &ConsoleLogger{}
	}

	if fc.io.Connect(conf.IOArgs) != nil {
		log.Error("could not connect to logger, falling back to console logger")
		fc.io = &ConsoleLogger{}
	}

	fc.base = fact.NewTrace()

	fc.base.BootTime = timestamppb.Now()
	fc.base.ContainerID = containerID
	fc.base.Runtime = fmt.Sprintf("%s %s %s", runtime.GOOS, runtime.GOARCH, runtime.Version())
	fc.base.Timestamp = timestamppb.Now()

	if conf.Platform != nil {
		fc.inspectorFromPlatformType(*conf.Platform)
	}

	if fc.platformInspector == nil {
		fc.inspectorFromEnvironment()
	}

	log.Infof("Detected %s platfrom", fc.platformInspector.Name())

	fc.platformInspector.Init(&fc.base)

	fc.sendOnUpdate = conf.SendOnUpdate

	fc.base = fc.platformInspector.Collect(fc.base, nil)

	if conf.IncludeEnvironment {
		log.Warn("includeEnvironment is set, this can leak sensetive information")
		for _, env := range os.Environ() {
			kv := strings.Split(env, "=")
			fc.base.Env[kv[0]] = kv[1]
		}

	}
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

func (fc *FactClient) Parent(parent string) {
	fc.trace.ChildOf = parent
}

func (fc *FactClient) Start(context interface{}, event interface{}) {
	trace := fc.platformInspector.Collect(fc.base, context)
	trace.ID = uuid.New().String()
	trace.StartTime = timestamppb.Now()

	if fc.sendOnUpdate {
		fc.send(trace)
	}

	fc.trace = trace
}

func (fc *FactClient) Update(context interface{}, msg *string, tags map[string]string) {
	trace := fc.platformInspector.Collect(fc.trace, context)
	if msg != nil {
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

func (fc *FactClient) Done(context interface{}, msg *string, args ...string) fact.Trace {
	trace := fc.platformInspector.Collect(fc.trace, context)
	trace.EndTime = timestamppb.Now()

	if msg != nil {
		trace.Logs[uint64(time.Now().Unix())] = *msg
	}

	for _, arg := range args {
		trace.Args = append(trace.Args, arg)
	}
	trace.ExecutionLatency = durationpb.New(trace.EndTime.AsTime().Sub(trace.StartTime.AsTime()))

	fc.send(trace)

	return trace
}

func (fc *FactClient) send(trace fact.Trace) {
	if fc.io != nil {
		err := fc.io.Send(trace)
		if err != nil {
			log.Errorf("failed to send trace:%+v - %f", trace, err)
		}
	}
}
