package main

import (
	"fmt"
	"os"

	"gitlab.com/parrotsec/packages/anonsurf/internal/anonsurf"
	"gitlab.com/parrotsec/packages/anonsurf/internal/monitor"
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
		return (anonsurf.Controller{}).Start()
	case "stop":
		return (anonsurf.Controller{}).Stop()
	case "restart":
		return (anonsurf.Controller{}).Restart()
	case "recover":
		return (anonsurf.Controller{}).Recover()
	case "changeid":
		return (anonsurf.Controller{}).ChangeIdentity()
	case "monitor":
		return monitor.NewNyxMonitor(os.Stdin, os.Stdout, os.Stderr).Run()
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}
