package anonsurf

import (
	"fmt"
	"strings"

	"gitlab.com/parrotsec/packages/anonsurf/internal/firewall"
	"gitlab.com/parrotsec/packages/anonsurf/internal/systemd"
	"gitlab.com/parrotsec/packages/anonsurf/internal/tor"
)

type restartSnapshot struct {
	torrc      string
	routing    tor.RoutingConfig
	dnsManaged bool
}

func (c Controller) Restart() error {
	active, err := fileExists(ActiveStatePath)
	if err != nil {
		return fmt.Errorf("check active state: %w", err)
	}
	if !active {
		return ErrNotRunning
	}

	snapshot, err := c.captureRestartSnapshot()
	if err != nil {
		return err
	}

	torService := systemd.NewTorController()
	if err := torService.Stop(); err != nil {
		return c.failRestart(fmt.Errorf("stop Tor service: %w", err), snapshot)
	}
	if err := c.generateTorrc(); err != nil {
		return c.failRestart(fmt.Errorf("generate torrc: %w", err), snapshot)
	}
	routing, err := c.readTorRoutingConfig()
	if err != nil {
		return c.failRestart(err, snapshot)
	}
	if err := (firewall.Manager{}).Apply(routing); err != nil {
		return c.failRestart(err, snapshot)
	}
	if err := c.startTorWithDNSMode(!routing.PreserveSystemDNS); err != nil {
		return c.failRestart(err, snapshot)
	}
	return nil
}

func (c Controller) captureRestartSnapshot() (restartSnapshot, error) {
	data, err := readFile(tor.TorrcPath)
	if err != nil {
		return restartSnapshot{}, fmt.Errorf("read current torrc: %w", err)
	}
	routing, err := tor.ParseRoutingConfig(strings.NewReader(string(data)))
	if err != nil {
		return restartSnapshot{}, fmt.Errorf("parse current torrc: %w", err)
	}
	dnsManaged, err := fileExists(DNSStatePath)
	if err != nil {
		return restartSnapshot{}, fmt.Errorf("check DNS state: %w", err)
	}
	return restartSnapshot{
		torrc:      string(data),
		routing:    routing,
		dnsManaged: dnsManaged,
	}, nil
}

func (c Controller) startTorWithDNSMode(managed bool) error {
	if err := c.setDNSManaged(false); err != nil {
		return fmt.Errorf("prepare system DNS for Tor: %w", err)
	}
	if err := systemd.NewTorController().Start(); err != nil {
		return fmt.Errorf("start Tor service: %w", err)
	}
	if managed {
		if err := c.setDNSManaged(true); err != nil {
			return fmt.Errorf("restore Tor DNS mode: %w", err)
		}
	}
	return nil
}

func (c Controller) failRestart(operation error, snapshot restartSnapshot) error {
	rollback := c.rollbackRestart(snapshot)
	if rollback == nil {
		return operation
	}
	return transitionError{operation: operation, rollback: rollback}
}

func (c Controller) rollbackRestart(snapshot restartSnapshot) error {
	torService := systemd.NewTorController()
	var failures errorCollector
	failures.add("stop Tor service", torService.Stop())
	failures.add("restore previous torrc", (tor.TorrcManager{}).Write(snapshot.torrc))
	failures.add("restore previous firewall", (firewall.Manager{}).Apply(snapshot.routing))
	failures.add("prepare system DNS for Tor", c.setDNSManaged(false))
	if err := failures.err(); err != nil {
		return err
	}
	if err := torService.Start(); err != nil {
		return fmt.Errorf("restart previous Tor service: %w", err)
	}
	if snapshot.dnsManaged {
		if err := c.setDNSManaged(true); err != nil {
			return fmt.Errorf("restore previous DNS mode: %w", err)
		}
	}
	return nil
}
