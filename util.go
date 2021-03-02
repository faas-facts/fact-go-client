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
	"os"
	"syscall"
)

func uptime() int64 {
	si := &syscall.Sysinfo_t{}
	_ = syscall.Sysinfo(si)
	return si.Uptime
}

func (fc *FactClient) inspectorFromEnvironment() {
	var inspect Inspector

	awsKey := os.Getenv("AWS_LAMBDA_LOG_STREAM_NAME")
	gcfKey := os.Getenv("X_GOOGLE_FUNCTION_NAME")
	owKey := os.Getenv("__OW_ACTION_NAME")
	acfKey := os.Getenv("WEBSITE_HOSTNAME")

	if awsKey != "" {
		inspect = &AWSInspector{}
	} else if gcfKey != "" {
		inspect = &GCFInspector{}
	} else if acfKey != "" {
		inspect = &ACFInspector{}
	} else if owKey != "" {
		//TODO OW.init
		if _, err := os.Stat("/sys/hypervisor/uuid"); err == nil {
			inspect = &ICFInspector{}
		} else {
			inspect = &OWInspector{}
		}
	} else {
		//TODO OW.init
		inspect = &GenericInspector{}
	}

	fc.platformInspector = inspect
}
