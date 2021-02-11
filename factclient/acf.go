package factclient

import (
	"github.com/faas-facts/fact/fact"
	"os"
	"strconv"
)

type ACFInspector struct {
	GenericInspector
}

func (A ACFInspector) Name() string {
	return "ACF"
}

func (A ACFInspector) Init(trace *fact.Trace){
	trace.Platform=A.Name()
	trace.ContainerID=os.Getenv("WEBSITE_HOSTNAME")
	trace.Region=os.Getenv("REGION_NAME")
	trace.HostID=os.Getenv("COMPUTERNAME")
	trace.Tags["service"]=os.Getenv("WEBSITE_INSTANCE_ID")
	trace.Tags["decrpytion_key"]=os.Getenv("MACHINEKEY_DecryptionKey")
	trace.Tags["uptime"]= strconv.FormatInt(uptime(), 10)
}

