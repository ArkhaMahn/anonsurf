package main

import (
	"fmt"
	"os"

	"gitlab.com/parrotsec/packages/anonsurf/internal/anondaemon"
	"gitlab.com/parrotsec/packages/anonsurf/internal/monitor"
	"gitlab.com/parrotsec/packages/anonsurf/internal/tor"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "anondaemon:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("expected one command: start, stop, restart, recover, changeid, or monitor")
	}

	switch args[0] {
	case "start":
		return (anondaemon.Controller{}).Start()
	case "stop":
		return (anondaemon.Controller{}).Stop()
	case "restart":
		return (anondaemon.Controller{}).Restart()
	case "recover":
		return (anondaemon.Controller{}).Recover()
	case "changeid":
		return (tor.IdentityChanger{}).Change()
	case "monitor":
		return monitor.NewNyxMonitor(os.Stdin, os.Stdout, os.Stderr).Run()
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}
