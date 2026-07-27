package main

import (
	"errors"
	"fmt"
	"os"

	"gitlab.com/parrotsec/packages/anonsurf/internal/config"
	"gitlab.com/parrotsec/packages/anonsurf/internal/tor"
)

type torrcManager interface {
	Backup() error
	Restore() error
	Write(content string) error
}

type nyxrcManager interface {
	Ensure() error
}

type bridgeResolver interface {
	Resolve(options tor.TorrcOptions) (tor.TorrcOptions, error)
}

type dependencies struct {
	configPath     string
	loadConfig     func(path string) (config.Config, error)
	resolveBridges bridgeResolver
	renderTorrc    func(options tor.TorrcOptions) (string, error)
	manager        torrcManager
	nyxrc          nyxrcManager
}

func main() {
	if err := run(os.Args[1:], defaultDependencies()); err != nil {
		fmt.Fprintln(os.Stderr, "make-torrc:", err)
		os.Exit(1)
	}
}

func defaultDependencies() dependencies {
	return dependencies{
		configPath:     config.SystemPath,
		loadConfig:     config.Load,
		resolveBridges: tor.NewBridgeResolver(),
		renderTorrc:    tor.RenderTorrc,
		manager:        tor.NewTorrcManager(),
		nyxrc:          tor.NewNyxrcManager(),
	}
}

func run(args []string, deps dependencies) error {
	if len(args) == 0 {
		return generateTorrc(deps)
	}
	if len(args) > 1 {
		return fmt.Errorf("expected at most one argument")
	}

	switch args[0] {
	case "new-torrc":
		return generateTorrc(deps)
	case "restore":
		return deps.manager.Restore()
	default:
		return fmt.Errorf("unknown argument %q", args[0])
	}
}

func generateTorrc(deps dependencies) error {
	cfg, err := deps.loadConfig(deps.configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	options := tor.TorrcOptionsFromConfig(cfg)
	options, err = deps.resolveBridges.Resolve(options)
	if err != nil {
		return fmt.Errorf("resolve bridges: %w", err)
	}

	rendered, err := deps.renderTorrc(options)
	if err != nil {
		return fmt.Errorf("render torrc: %w", err)
	}

	if err := deps.manager.Backup(); err != nil && !errors.Is(err, tor.ErrTorrcBackupExists) {
		return fmt.Errorf("backup torrc: %w", err)
	}
	if err := deps.manager.Write(rendered); err != nil {
		return fmt.Errorf("write torrc: %w", err)
	}
	if err := deps.nyxrc.Ensure(); err != nil {
		return fmt.Errorf("ensure nyxrc: %w", err)
	}

	return nil
}
