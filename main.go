// procsnitch daemon - UNIX domain socket service providing process information for local network connections

package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"unsafe"

	"github.com/op/go-logging"
	"github.com/subgraph/go-procsnitch"
	"github.com/subgraph/procsnitchd/protocol"
	"github.com/subgraph/procsnitchd/service"
)

var log = logging.MustGetLogger("procsnitchd")

var logFormat = logging.MustStringFormatter(
	"%{level:.4s} %{id:03x} %{message}",
)
var ttyFormat = logging.MustStringFormatter(
	"%{color}%{time:15:04:05} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}",
)

const ioctlReadTermios = 0x5401

func isTerminal(fd int) bool {
	var termios syscall.Termios
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), ioctlReadTermios, uintptr(unsafe.Pointer(&termios)), 0, 0, 0)
	return err == 0
}

func stringToLogLevel(level string) (logging.Level, error) {

	switch level {
	case "DEBUG":
		return logging.DEBUG, nil
	case "INFO":
		return logging.INFO, nil
	case "NOTICE":
		return logging.NOTICE, nil
	case "WARNING":
		return logging.WARNING, nil
	case "ERROR":
		return logging.ERROR, nil
	case "CRITICAL":
		return logging.CRITICAL, nil
	}
	return -1, fmt.Errorf("invalid logging level %s", level)
}

func setupLoggerBackend(level logging.Level) logging.LeveledBackend {
	format := logFormat
	if isTerminal(int(os.Stderr.Fd())) {
		format = ttyFormat
	}
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	formatter := logging.NewBackendFormatter(backend, format)
	leveler := logging.AddModuleLevel(formatter)
	leveler.SetLevel(level, "procsnitchd")
	return leveler
}

func main() {
	var level logging.Level
	var logLevel string
	var err error

	socketFile := flag.String("socket", "", "UNIX domain socket file")
	group := flag.String("group", "", "Group ownership of the socket file")
	flag.StringVar(&logLevel, "log_level", "INFO", "logging level could be set to: DEBUG, INFO, NOTICE, WARNING, ERROR, CRITICAL")
	flag.Parse()

	level, err = stringToLogLevel(logLevel)
	if err != nil {
		log.Critical("Invalid logging-level specified.")
		os.Exit(1)
	}
	logBackend := setupLoggerBackend(level)
	log.SetBackend(logBackend)
	procsnitch.SetLogger(log)
	protocol.SetLogger(log)
	service.SetLogger(log)

	if os.Geteuid() != 0 {
		log.Error("Must be run as root")
		os.Exit(1)
	}

	if *socketFile == "" {
		log.Critical("UNIX domain socket file must be specified!")
		os.Exit(1)
	}
	if *group == "" {
		log.Critical("group ownership of our UNIX domain socket file must be specified!")
		os.Exit(1)
	}

	procInfo := procsnitch.SystemProcInfo{}
	service := service.NewMortalService("unix", *socketFile, protocol.ConnectionHandlerFactory(procInfo))
	err = service.Start()
	if err != nil {
		log.Criticalf("failed to start listener %s", err)
	}

	log.Notice("procsnitchd starting")

	// change the group ownership / permissions of the UNIX domain socket
	cmd := exec.Command("/bin/chgrp", *group, *socketFile)
	err = cmd.Run()
	if err != nil {
		log.Criticalf("failed to chmod socket: %s", err)
		panic("wtf")
	}
	mode := 0775
	err = os.Chmod(*socketFile, os.FileMode(mode))
	if err != nil {
		log.Critical("cannot chmod socket file")
		panic("wtf")
	}

	// wait for a control-c or kill signal
	sigKillChan := make(chan os.Signal, 1)
	signal.Notify(sigKillChan, os.Interrupt, os.Kill)
	for {
		select {
		case <-sigKillChan:
			log.Notice("procsnitchd stopping")
			service.Stop()
			return
		}
	}
}
