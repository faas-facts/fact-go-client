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
	trace.HostID = fmt.Sprintf("H_%d", uptime)
}

func (g GenericInspector) Collect(trace fact.Trace, context interface{}) fact.Trace {
	trace.ExecutionLatency = durationpb.New(time.Now().Sub(trace.StartTime.AsTime()))
	return trace
}
