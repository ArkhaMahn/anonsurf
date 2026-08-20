package main

import (
	"errors"
	"fmt"
	"os"

	"gitlab.com/parrotsec/packages/anonsurf/internal/config"
	"gitlab.com/parrotsec/packages/anonsurf/internal/tor"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "make-torrc:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return generateTorrc()
	}
	if len(args) > 1 {
		return fmt.Errorf("expected at most one argument")
	}

	switch args[0] {
	case "new-torrc":
		return generateTorrc()
	case "restore":
		if err := (tor.TorrcManager{}).Restore(); err != nil && !errors.Is(err, tor.ErrTorrcBackupNotFound) {
			return err
		}
		return nil
	default:
		return fmt.Errorf("unknown argument %q", args[0])
	}
}

func generateTorrc() error {
	cfg, err := config.Load(config.SystemPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	options := tor.TorrcOptionsFromConfig(cfg)
	options, err = (tor.BridgeResolver{}).Resolve(options)
	if err != nil {
		return fmt.Errorf("resolve bridges: %w", err)
	}

	rendered, err := tor.RenderTorrc(options)
	if err != nil {
		return fmt.Errorf("render torrc: %w", err)
	}

	if err := (tor.NyxrcManager{}).Ensure(); err != nil {
		return fmt.Errorf("ensure nyxrc: %w", err)
	}
	manager := tor.TorrcManager{}
	if err := manager.Backup(); err != nil && !errors.Is(err, tor.ErrTorrcBackupExists) {
		return fmt.Errorf("backup torrc: %w", err)
	}
	if err := manager.Write(rendered); err != nil {
		return fmt.Errorf("write torrc: %w", err)
	}

	return nil
}
