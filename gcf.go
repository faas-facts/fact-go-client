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
	"math"
	"os"
	"strconv"
)

type GCFInspector struct {
	GenericInspector
}

func (G GCFInspector) Name() string {
	return "GCF"
}

func (G GCFInspector) Init(trace *fact.Trace) {
	trace.Tags["service"] = os.Getenv("SUPERVISOR_HOSTNAME")
	trace.Platform = G.Name()
	utime := strconv.FormatInt(uptime(), 10)
	trace.Region = os.Getenv("X_GOOGLE_FUNCTION_REGION")
	trace.Tags["uptime"] = utime
	trace.Tags["fname"] = os.Getenv("X_GOOGLE_FUNCTION_NAME")
	trace.Tags["service"] = os.Getenv("X_GOOGLE_SUPERVISOR_HOSTNAME")
	mem, _ := strconv.ParseInt(os.Getenv("X_GOOGLE_FUNCTION_MEMORY_MB"), 10, 32)
	trace.Memory = int32(mem)
	trace.HostID = utime
}

func (G GCFInspector) Collect(trace fact.Trace, context interface{}) fact.Trace {
	t := G.GenericInspector.Collect(trace, context)

	t.Cost = float32(G.cost(t))

	return t

}

func (G GCFInspector) cost(t fact.Trace) float64 {
	duration := math.Floor(float64(t.ExecutionLatency.Seconds / 100.0))
	mb, _ := strconv.ParseInt(os.Getenv("X_GOOGLE_FUNCTION_MEMORY_MB"), 10, 32)

	if mb <= 128 {
		return 0.000000231 * duration
	} else if mb <= 256 {
		return 0.000000463 * duration
	} else if mb <= 512 {
		return 0.000000925 * duration
	} else if mb <= 1024 {
		return 0.000001650 * duration
	} else if mb <= 2048 {
		return 0.000002900 * duration
	} else {
		return 0.000005800 * duration
	}
}
