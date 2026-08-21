package anondaemon

import (
	"fmt"

	"gitlab.com/parrotsec/packages/anonsurf/internal/systemd"
)

type startTransition struct {
	torWasActive      bool
	torStartAttempted bool
	torrcGenerated    bool
	dnsTouched        bool
	ipv6Touched       bool
	firewallTouched   bool
}

func (c Controller) failStart(operation error, state startTransition) error {
	rollback := c.rollbackStart(state)
	if rollback == nil {
		return operation
	}
	return transitionError{operation: operation, rollback: rollback}
}

func (c Controller) rollbackStart(state startTransition) error {
	var failures errorCollector
	if state.torStartAttempted {
		failures.add("stop Tor service", systemd.NewTorController().Stop())
	}
	if state.firewallTouched {
		failures.add("restore firewall", c.restoreFirewall())
	}
	if state.ipv6Touched {
		failures.add("restore IPv6 state", c.restoreIPv6State())
	}
	if state.dnsTouched {
		failures.add("restore DNS settings", c.restoreDNS())
	}
	if state.torrcGenerated {
		failures.add("restore torrc", c.restoreTorrc())
	}
	if err := failures.err(); err != nil {
		return err
	}
	if state.torWasActive {
		if err := systemd.NewTorController().Start(); err != nil {
			return fmt.Errorf("restart original Tor service: %w", err)
		}
		if err := clearTorState(); err != nil {
			return err
		}
	}
	return nil
}
