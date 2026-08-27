package anonsurf

import (
	"fmt"

	"gitlab.com/parrotsec/packages/anonsurf/internal/dns"
	"gitlab.com/parrotsec/packages/anonsurf/internal/firewall"
	"gitlab.com/parrotsec/packages/anonsurf/internal/ipv6"
	"gitlab.com/parrotsec/packages/anonsurf/internal/tor"
)

const (
	ActiveStatePath    = "/run/anonsurf/active"
	DNSStatePath       = "/var/lib/anonsurf/dns.changed"
	TorStatePath       = "/run/anonsurf/tor.was-active"
	stateMarkerContent = "anonsurf\n"
)

func (c Controller) markActive() error {
	return writeFile(ActiveStatePath, []byte(stateMarkerContent), 0600)
}

func (c Controller) clearActive() error {
	if err := removeFile(ActiveStatePath); err != nil {
		return fmt.Errorf("remove active state: %w", err)
	}
	return nil
}

func (c Controller) markDNSChanged() error {
	return writeFile(DNSStatePath, []byte(stateMarkerContent), 0600)
}

func captureTorState(active bool) (bool, error) {
	exists, err := fileExists(TorStatePath)
	if err != nil {
		return false, fmt.Errorf("check Tor state: %w", err)
	}
	if exists {
		return true, nil
	}
	if !active {
		return false, nil
	}
	if err := writeFile(TorStatePath, []byte(stateMarkerContent), 0600); err != nil {
		return false, fmt.Errorf("save Tor state: %w", err)
	}
	return true, nil
}

func clearTorState() error {
	if err := removeFile(TorStatePath); err != nil {
		return fmt.Errorf("remove Tor state: %w", err)
	}
	return nil
}

func torWasActive() (bool, error) {
	exists, err := fileExists(TorStatePath)
	if err != nil {
		return false, fmt.Errorf("check Tor state: %w", err)
	}
	return exists, nil
}

func (c Controller) restoreDNS() error {
	path := DNSStatePath
	exists, err := fileExists(path)
	if err != nil {
		return fmt.Errorf("check DNS state: %w", err)
	}
	if !exists {
		return nil
	}
	if err := (dns.Manager{}).RestoreBackupOrDHCP(); err != nil {
		return err
	}
	if err := removeFile(path); err != nil {
		return fmt.Errorf("remove DNS state: %w", err)
	}
	return nil
}

func (c Controller) hasManagedState() (bool, error) {
	paths := []string{
		ActiveStatePath,
		firewall.StatePath,
		DNSStatePath,
		ipv6.StatePath,
		TorStatePath,
		tor.TorrcBackupPath,
		tor.TorrcMissingMarkerPath,
	}
	for _, path := range paths {
		exists, err := fileExists(path)
		if err != nil {
			return false, fmt.Errorf("check managed state %s: %w", path, err)
		}
		if exists {
			return true, nil
		}
	}
	return false, nil
}

func (c Controller) Recover() error {
	active, err := fileExists(ActiveStatePath)
	if err != nil {
		return fmt.Errorf("check active state: %w", err)
	}
	if active {
		return nil
	}

	managed, err := c.hasManagedState()
	if err != nil {
		return err
	}
	if !managed {
		return nil
	}

	var failures errorCollector
	failures.add("restore firewall", (firewall.Manager{}).Restore())
	failures.add("restore DNS settings", c.restoreDNS())
	failures.add("restore torrc", c.restoreTorrc())
	failures.add("restore IPv6 state", (ipv6.Manager{}).Restore())
	if err := failures.err(); err != nil {
		return err
	}
	if err := c.clearActive(); err != nil {
		return err
	}
	return clearTorState()
}
