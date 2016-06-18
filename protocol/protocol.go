package protocol

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

type ProcsnitchRPC struct {
	procInfo procsnitch.ProcInfo
}

func NewProcsnitchRPC(procInfo procsnitch.ProcInfo) *ProcsnitchRPC {
	rpc := ProcsnitchRPC{
		procInfo: procInfo,
	}
	return &rpc
}

func (t *ProcsnitchRPC) LookupUNIXSocketProcess(socketFile *string, info *procsnitch.Info) error {
	info = t.procInfo.LookupUNIXSocketProcess(*socketFile)
	return nil
}

func HandleNewConnection(conn net.Conn) error {
	s := NewProcSnitchSession(conn)
	return s.Start()
}

type ProcSnitchSession struct {
	conn      net.Conn
	rpcServer *rpc.Server
}

func NewProcSnitchSession(conn net.Conn) *ProcSnitchSession {
	p := ProcSnitchSession{
		conn:      conn,
		rpcServer: rpc.NewServer(),
	}
	rpc := NewProcsnitchRPC(procsnitch.SystemProcInfo{})
	p.rpcServer.Register(rpc)
	return &p
}

func (s *ProcSnitchSession) Start() error {
	s.rpcServer.ServeConn(s.conn)
	return nil
}
