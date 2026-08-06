package main

import (
	"fmt"
	"io"
	"os"

	"gitlab.com/parrotsec/packages/anonsurf/internal/anondaemon"
)

type daemonController interface {
	Start() error
	Stop() error
	Restart() error
}

type dependencies struct {
	daemon daemonController
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr, defaultDependencies()); err != nil {
		fmt.Fprintln(os.Stderr, "anondaemon:", err)
		os.Exit(1)
	}
}

func defaultDependencies() dependencies {
	return dependencies{
		daemon: anondaemon.NewController(),
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
