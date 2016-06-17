package protocol

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/op/go-logging"
	"github.com/subgraph/go-procsnitch"
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
	conn       net.Conn
	connReader *bufio.Reader
	procInfo   procsnitch.ProcInfo
}

func NewProcSnitchSession(conn net.Conn) *ProcSnitchSession {
	p := ProcSnitchSession{
		conn:       conn,
		connReader: bufio.NewReader(conn),
		procInfo:   procsnitch.SystemProcInfo{},
	}
	return &p
}

func (s *ProcSnitchSession) readLine() (cmd string, splitCmd []string, rawLine []byte, err error) {
	if rawLine, err = s.connReader.ReadBytes('\n'); err != nil {
		return
	}
	trimmedLine := bytes.TrimSpace(rawLine)
	splitCmd = strings.Split(string(trimmedLine), " ")
	cmd = strings.ToUpper(strings.TrimSpace(splitCmd[0]))
	return
}

func (s *ProcSnitchSession) writeProcInfo(info *procsnitch.Info) {
	out := fmt.Sprintf("%d %d %d %s %s", info.UID, info.Pid, info.ParentPid, info.ExePath, info.CmdLine)
	s.conn.Write([]byte(out + "\r\n"))
}

func (s *ProcSnitchSession) onCmdUnixInfo(args []string) error {
	if len(args) < 2 {
		log.Error("invalid number of arguments")
		return fmt.Errorf("invalid number of arguments")
	}
	info := s.procInfo.LookupUNIXSocketProcess(args[1])
	s.writeProcInfo(info)
	return nil
}

func (s *ProcSnitchSession) onCmdTcpInfo(args []string) error {
	var srcPort, dstPort int64
	var dstAddr net.IP
	var err error

	if len(args) < 4 {
		log.Error("invalid number of arguments")
		return fmt.Errorf("invalid number of arguments")
	}

	srcPort, err = strconv.ParseInt(args[1], 10, 16)
	if err != nil {
		log.Error("failed to parse TCP port")
		return fmt.Errorf("failed to parse TCP port")
	}
	dstAddr = net.ParseIP(args[2])
	if dstAddr == nil {
		log.Error("failed to parse IP")
		return fmt.Errorf("failed to parse IP")
	}
	dstPort, err = strconv.ParseInt(args[3], 10, 16)
	if err != nil {
		log.Error("failed to parse TCP port")
		return fmt.Errorf("failed to parse TCP port")
	}

	info := s.procInfo.LookupTCPSocketProcess(uint16(srcPort), dstAddr, uint16(dstPort))
	s.writeProcInfo(info)
	return nil
}

func (s *ProcSnitchSession) onCmdUdpInfo(args []string) error {
	var srcPort int64
	var err error

	srcPort, err = strconv.ParseInt(args[1], 10, 16)
	if err != nil {
		log.Error("failed to parse UDP port")
		return fmt.Errorf("failed to parse UDP port")
	}
	info := s.procInfo.LookupUDPSocketProcess(uint16(srcPort))
	s.writeProcInfo(info)
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
