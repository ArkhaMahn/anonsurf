package anondaemon

import "fmt"

const stateMarkerContent = "anonsurf\n"

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
