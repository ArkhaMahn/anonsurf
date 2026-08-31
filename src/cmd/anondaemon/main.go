package main

import (
	"fmt"
	"os"

	"gitlab.com/parrotsec/packages/anonsurf/internal/anonsurf"
	"gitlab.com/parrotsec/packages/anonsurf/internal/monitor"
	"gitlab.com/parrotsec/packages/anonsurf/internal/systemd"
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

	controller := anonsurf.NewController(newStatusReporter())
	switch args[0] {
	case "start":
		return controller.Start()
	case "stop":
		return controller.Stop()
	case "restart":
		return controller.Restart()
	case "recover":
		return controller.Recover()
	case "changeid":
		return controller.ChangeIdentity()
	case "monitor":
		return monitor.NewNyxMonitor(os.Stdin, os.Stdout, os.Stderr).Run()
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func newStatusReporter() anonsurf.StatusReporter {
	available := true
	return func(status string) {
		if !available {
			return
		}
		if err := systemd.NotifyStatus(status); err != nil {
			fmt.Fprintln(os.Stderr, "anondaemon: report service progress:", err)
			available = false
		}
	}
}
