package anondaemon

import "fmt"

func (c Controller) Stop() error {
	managed, err := c.hasManagedState()
	if err != nil {
		return err
	}
	if !managed {
		return nil
	}

	var failures errorCollector
	failures.add("stop Tor service", c.torService().Stop())
	failures.add("restore firewall", c.restoreFirewall())
	failures.add("restore DNS settings", c.restoreDNS())
	failures.add("restore torrc", c.restoreTorrc())
	failures.add("restore IPv6 state", c.restoreIPv6State())
	if err := failures.err(); err != nil {
		return err
	}
	return c.clearActive()
}

func (c Controller) restoreTorrc() error {
	if err := c.run(c.makeTorrcPath(), "restore"); err != nil {
		return fmt.Errorf("run torrc restore: %w", err)
	}
	return nil
}
