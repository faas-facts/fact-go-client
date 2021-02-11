package factclient

import (
	"encoding/json"
	"fmt"
	"github.com/faas-facts/fact/fact"
)

type ConsoleLogger struct{}

func (c ConsoleLogger) Connect(m map[string]string) error {
	return nil
}

func (c ConsoleLogger) Send(trace fact.Trace) error {
	data, _ := json.Marshal(trace)
	fmt.Println(data)

	return nil
}
