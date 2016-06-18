package client

import (
	"net"
	"net/rpc"

	"github.com/op/go-logging"
	"github.com/subgraph/go-procsnitch"
)

var log = logging.MustGetLogger("procsnitchd_protocol")

// SetLogger allows setting a custom go-logging instance
func SetLogger(logger *logging.Logger) {
	log = logger
}

type SnitchClient struct {
	conn       net.Conn
	client     *rpc.Client
	socketFile string
}

func NewSnitchClient(socketFile string) *SnitchClient {
	s := SnitchClient{
		socketFile: socketFile,
	}
	return &s
}

func (s *SnitchClient) Start() error {
	var err error
	s.conn, err = net.Dial("unix", s.socketFile)
	if err != nil {
		return err
	}
	s.client = rpc.NewClient(s.conn)
	return nil
}

// implements the go-procsnitch ProcInfo interface

func (s *SnitchClient) LookupUNIXSocketProcess(socketFile string) *procsnitch.Info {
	var err error
	var info procsnitch.Info
	err = s.client.Call("ProcsnitchRPC.LookupUNIXSocketProcess", &socketFile, &info)
	if err != nil {
		panic("wtf")
	}
	return &info
}

func (s *SnitchClient) LookupTCPSocketProcess(srcPort uint16, dstAddr net.IP, dstPort uint16) *procsnitch.Info {
	return nil
}

func (s *SnitchClient) LookupUDPSocketProcess(srcPort uint16) *procsnitch.Info {
	return nil
}
