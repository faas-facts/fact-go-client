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
