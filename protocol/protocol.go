package protocol

import (
	"bufio"
	"net"
)

type ProcSnitchSession struct {
	appConn       net.Conn
	appConnReader *bufio.Reader
}

func NewProcSnitchSession(conn net.Conn) *ProcSnitchSession {
	p := ProcSnitchSession{
		appConn:       conn,
		appConnReader: bufio.NewReader(conn),
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

func (s *ProcSnitchSession) Start() error {

	for {
		cmd, splitCmd, _, err := s.readLine()
		if err != nil {
			log.Errorf("Failed reading client request: %s", err)
			return err
		}

	}

	return nil
}
