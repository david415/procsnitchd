package protocol

import (
	"bufio"
	"bytes"
	"net"
	"strings"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("procsnitchd_protocol")

// SetLogger allows setting a custom go-logging instance
func SetLogger(logger *logging.Logger) {
	log = logger
}

const (
	cmdTcpInfo  = "TCPINFO"
	cmdUnixInfo = "UNIXINFO"
	cmdUdpInfo  = "UDPINFO"
)

func HandleNewConnection(conn net.Conn) error {
	s := NewProcSnitchSession(conn)
	return s.Start()
}

type ProcSnitchSession struct {
	appConn       net.Conn
	appConnReader *bufio.Reader
	procInfo      procsnitch.ProcInfo
}

func NewProcSnitchSession(conn net.Conn) *ProcSnitchSession {
	p := ProcSnitchSession{
		appConn:       conn,
		appConnReader: bufio.NewReader(conn),
		procInfo:      procsnitch.SystemProcInfo{},
	}
	return &p
}

func (s *ProcSnitchSession) readLine() (cmd string, splitCmd []string, rawLine []byte, err error) {
	if rawLine, err = s.appConnReader.ReadBytes('\n'); err != nil {
		return
	}
	trimmedLine := bytes.TrimSpace(rawLine)
	splitCmd = strings.Split(string(trimmedLine), " ")
	cmd = strings.ToUpper(strings.TrimSpace(splitCmd[0]))
	return
}

/*
type ProcInfo interface {
	LookupTCPSocketProcess(srcPort uint16, dstAddr net.IP, dstPort uint16) *Info
	LookupUNIXSocketProcess(socketFile string) *Info
	LookupUDPSocketProcess(srcPort uint16) *Info
}
// Info is a struct containing the result of a socket proc query
type Info struct {
	UID       int
	Pid       int
	ParentPid int
	loaded    bool
	ExePath   string
	CmdLine   string
}
*/
func (s *ProcSnitchSession) onCmdUnixInfo(args []string) error {
	if len(args) < 2 {
		log.Error("invalid number of arguments")
		return fmt.Error("invalid number of arguments")
	}
	info := s.procInfo.LookupUNIXSocketProcess(args[1])
	fmt.Sprintf("")
	return nil
}

func (s *ProcSnitchSession) onCmdTcpInfo(args []string) error {
	return nil
}

func (s *ProcSnitchSession) onCmdUdpInfo(args []string) error {
	return nil
}

func (s *ProcSnitchSession) Start() error {
	for {
		cmd, splitCmd, _, err := s.readLine()
		if err != nil {
			log.Errorf("Failed reading client request: %s", err)
			return err
		}
		switch cmd {
		case cmdUnixInfo:
			if err = s.onCmdUnixInfo(splitCmd); err != nil {
				return err
			}
		case cmdTcpInfo:
			if err = s.onCmdTcpInfo(splitCmd); err != nil {
				return err
			}
		case cmdUdpInfo:
			if err = s.onCmdUdpInfo(splitCmd); err != nil {
				return err
			}
		}
	}
	return nil
}
