package fact_go_client

import (
	"github.com/faas-facts/fact/fact"
	"github.com/magiconair/properties/assert"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAWSInspector_readCGroupIDs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cgroup.txt")
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0664)
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.Copy(file, strings.NewReader(cgroupExample))
	if err != nil {
		t.Fatal(err)
	}
	_ = file.Close()
	inspector := AWSInspector{
		GenericInspector{},
		path,
	}

	trace := fact.NewTrace()

	err = inspector.readCGroupIDs(&trace)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, trace.Tags["freezer"], "1d3254")
	assert.Equal(t, trace.Tags["host"], "siNKWg")
	assert.Equal(t, trace.Tags["service"], "9857cb")
	assert.Equal(t, trace.Tags["sandbox"], "36b98a")
	assert.Equal(t, trace.HostID, "siNKWg")

}

const cgroupExample = `11:freezer:/sandbox-1d3254
10:blkio:/
9:cpu,cpuacct:/sandbox-root-siNKWg/sandbox-service-9857cb/sandbox-36b98a
8:pids:/
7:hugetlb:/
6:devices:/
5:cpuset:/
4:memory:/sandbox-service-062f38/sandbox-dfb3bf
3:net_cls,net_prio:/
2:perf_event:/
1:name=systemd:/system.slice/sandbox.service`
