# Fact-Go-Client

This library is a go based client implementation for [Fact](https://github.com/faas-facts/fact).

Usage:

```go
package main

import (
	"encoding/json"
	fact "github.com/faas-facts/fact-go-client"
	"net/http"
)

var client *fact.FactClient

func init() {
	client = &fact.FactClient{}
	client.Boot(fact.FactClientConfig{
		SendOnUpdate:       false,
		IncludeEnvironment: false,
	})
}

//Faas Function handler (e.g. GCF)
func ServeHTTP(w http.ResponseWriter, req http.Request) {
	client.Start(nil, req)
	//...
	client.Update(nil, "some important event")
	//...
	t := client.Done(nil)
	b, _ := json.Marshal(t)
	_, _ = w.Write(b)
}
```