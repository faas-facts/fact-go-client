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
