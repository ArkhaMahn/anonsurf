package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

var errBackendNotImplemented = errors.New("backend is not implemented")

type daemonController interface {
	Start() error
	Stop() error
	Restart() error
}

type dependencies struct {
	daemon daemonController
}

type unsupportedDaemon struct{}

func (unsupportedDaemon) Start() error {
	return errBackendNotImplemented
}

func (unsupportedDaemon) Stop() error {
	return errBackendNotImplemented
}

func (unsupportedDaemon) Restart() error {
	return errBackendNotImplemented
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr, defaultDependencies()); err != nil {
		fmt.Fprintln(os.Stderr, "anondaemon:", err)
		os.Exit(1)
	}
}

func defaultDependencies() dependencies {
	return dependencies{
		daemon: unsupportedDaemon{},
	}
}

func run(args []string, stdout, stderr io.Writer, deps dependencies) error {
	if len(args) != 1 {
		return fmt.Errorf("expected one command: start, stop, or restart")
	}
	if deps.daemon == nil {
		return fmt.Errorf("daemon backend is missing")
	}

	switch args[0] {
	case "start":
		return deps.daemon.Start()
	case "stop":
		return deps.daemon.Stop()
	case "restart":
		return deps.daemon.Restart()
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}
