package anonsurf

import (
	"errors"
	"fmt"

	"gitlab.com/parrotsec/packages/anonsurf/internal/firewall"
	"gitlab.com/parrotsec/packages/anonsurf/internal/ipv6"
	"gitlab.com/parrotsec/packages/anonsurf/internal/systemd"
	"gitlab.com/parrotsec/packages/anonsurf/internal/tor"
)

func (c Controller) Stop() error {
	managed, err := c.hasManagedState()
	if err != nil {
		return err
	}
	if !managed {
		return nil
	}
	restartTor, err := torWasActive()
	if err != nil {
		return err
	}

	var failures errorCollector
	failures.add("stop Tor service", systemd.NewTorController().Stop())
	failures.add("restore firewall", (firewall.Manager{}).Restore())
	failures.add("restore DNS settings", c.restoreDNS())
	failures.add("restore torrc", c.restoreTorrc())
	failures.add("restore IPv6 state", (ipv6.Manager{}).Restore())
	if err := failures.err(); err != nil {
		return err
	}
	if restartTor && !systemd.IsSystemStopping() {
		if err := systemd.NewTorController().Start(); err != nil {
			return fmt.Errorf("restart original Tor service: %w", err)
		}
	}
	if err := c.clearActive(); err != nil {
		return err
	}
	if restartTor {
		return clearTorState()
	}
	return nil
}

func (c Controller) restoreTorrc() error {
	if err := (tor.TorrcManager{}).Restore(); err != nil && !errors.Is(err, tor.ErrTorrcBackupNotFound) {
		return fmt.Errorf("restore torrc: %w", err)
	}
	return nil
}
