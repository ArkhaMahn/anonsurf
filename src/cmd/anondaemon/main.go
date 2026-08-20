package main

import (
	"fmt"
	"io"
	"os"

	"gitlab.com/parrotsec/packages/anonsurf/internal/anondaemon"
	"gitlab.com/parrotsec/packages/anonsurf/internal/tor"
)

type daemonController interface {
	Start() error
	Stop() error
	Restart() error
}

type identityChanger interface {
	Change() error
}

type dependencies struct {
	daemon   daemonController
	identity identityChanger
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr, defaultDependencies()); err != nil {
		fmt.Fprintln(os.Stderr, "anondaemon:", err)
		os.Exit(1)
	}
}

func defaultDependencies() dependencies {
	return dependencies{
		daemon:   anondaemon.NewController(),
		identity: tor.NewIdentityChanger(),
	}
}

func run(args []string, stdout, stderr io.Writer, deps dependencies) error {
	if len(args) != 1 {
		return fmt.Errorf("expected one command: start, stop, restart, or changeid")
	}

	switch args[0] {
	case "start":
		if deps.daemon == nil {
			return fmt.Errorf("daemon backend is missing")
		}
		return deps.daemon.Start()
	case "stop":
		if deps.daemon == nil {
			return fmt.Errorf("daemon backend is missing")
		}
		return deps.daemon.Stop()
	case "restart":
		if deps.daemon == nil {
			return fmt.Errorf("daemon backend is missing")
		}
		return deps.daemon.Restart()
	case "changeid":
		if deps.identity == nil {
			return fmt.Errorf("identity backend is missing")
		}
		return deps.identity.Change()
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}
