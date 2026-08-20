package anondaemon

import "fmt"

const stateMarkerContent = "anonsurf\n"

func (c Controller) markActive() error {
	return c.writeFile(c.activeStatePath(), []byte(stateMarkerContent), 0600)
}

func (c Controller) clearActive() error {
	if err := c.removeFile(c.activeStatePath()); err != nil {
		return fmt.Errorf("remove active state: %w", err)
	}
	return nil
}

func (c Controller) markDNSChanged() error {
	return c.writeFile(c.dnsStatePath(), []byte(stateMarkerContent), 0600)
}

func (c Controller) restoreDNS() error {
	path := c.dnsStatePath()
	exists, err := c.fileExists(path)
	if err != nil {
		return fmt.Errorf("check DNS state: %w", err)
	}
	if !exists {
		return nil
	}
	if err := c.dnsManager().RestoreBackupOrDHCP(); err != nil {
		return err
	}
	if err := c.removeFile(path); err != nil {
		return fmt.Errorf("remove DNS state: %w", err)
	}
	return nil
}

func (c Controller) hasManagedState() (bool, error) {
	paths := []string{
		c.activeStatePath(),
		c.iptablesRulesPath(),
		c.dnsStatePath(),
		c.ipv6StatePath(),
		c.torrcBackupPath(),
		c.torrcMissingMarkerPath(),
	}
	for _, path := range paths {
		exists, err := c.fileExists(path)
		if err != nil {
			return false, fmt.Errorf("check managed state %s: %w", path, err)
		}
		if exists {
			return true, nil
		}
	}
	return false, nil
}
