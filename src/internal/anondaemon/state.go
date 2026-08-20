package anondaemon

import (
	"fmt"

	"gitlab.com/parrotsec/packages/anonsurf/internal/dns"
)

const stateMarkerContent = "anonsurf\n"

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
		IPTablesRulesPath,
		DNSStatePath,
		IPv6StatePath,
		TorrcBackupPath,
		TorrcMissingPath,
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
