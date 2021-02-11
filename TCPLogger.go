package fact_go_client

import (
	"fmt"
	"github.com/faas-facts/fact/fact"
	"github.com/golang/protobuf/proto"
	"net"
	"strconv"
)

type TCPLogger struct {
	address string
	port    int64
}

func (T *TCPLogger) Connect(m map[string]string) error {
	if address, ok := m["fact_tcp_address"]; !ok {
		return fmt.Errorf("missing address in config")
	} else {
		T.address = address
	}
	if port, ok := m["fact_tcp_port"]; !ok {
		T.port = 9999
	} else {
		T.port, _ = strconv.ParseInt(port, 10, 32)
	}
	dial, err := net.Dial("tcp", fmt.Sprintf("%s:%d", T.address, T.port))
	if err != nil {
		return err
	}
	_ = dial.Close()

	return nil
}

func (T *TCPLogger) Send(trace fact.Trace) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", T.address, T.port))
	if err != nil {
		return err
	}
	defer conn.Close()

	data, err := proto.Marshal(&trace)
	if err != nil {
		return err
	}

	_, err = conn.Write(data)
	if err != nil {
		return err
	}

	err = conn.Close()

	return nil
}
