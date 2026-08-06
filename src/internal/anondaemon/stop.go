package anondaemon

import "fmt"

func (c Controller) Stop() error {
	if err := c.stopSurf(); err != nil {
		return err
	}
	if err := c.restoreTorrc(); err != nil {
		return fmt.Errorf("restore torrc: %w", err)
	}
	return nil
}

func (c Controller) stopSurf() error {
	if err := c.enableIPv6(); err != nil {
		return err
	}
	if err := c.flushIPTablesRules(); err != nil {
		return err
	}
	if err := c.restoreIPTablesRules(); err != nil {
		return err
	}
	if err := c.stopTor(); err != nil {
		return err
	}
	if err := c.dnsManager().RestoreBackupOrDHCP(); err != nil {
		return fmt.Errorf("restore DNS settings: %w", err)
	}
	return nil
}

func (c Controller) enableIPv6() error {
	for _, setting := range []string{
		"net.ipv6.conf.all.disable_ipv6=0",
		"net.ipv6.conf.default.disable_ipv6=0",
	} {
		if err := c.run(c.sysctlPath(), "-w", setting); err != nil {
			return fmt.Errorf("enable IPv6: %w", err)
		}
	}
	return nil
}

func (c Controller) flushIPTablesRules() error {
	for _, args := range [][]string{{"-F"}, {"-t", "nat", "-F"}} {
		if err := c.run(c.iptablesPath(), args...); err != nil {
			return fmt.Errorf("flush iptables rules: %w", err)
		}
	}
	return nil
}

func (c Controller) restoreIPTablesRules() error {
	path := c.iptablesRulesPath()
	exists, err := c.fileExists(path)
	if err != nil {
		return fmt.Errorf("check iptables backup: %w", err)
	}
	if !exists {
		return nil
	}

	rules, err := c.readFile(path)
	if err != nil {
		return fmt.Errorf("read iptables backup: %w", err)
	}
	if err := c.runWithInput(c.iptablesRestorePath(), rules); err != nil {
		return fmt.Errorf("restore iptables rules: %w", err)
	}
	if err := c.removeFile(path); err != nil {
		return fmt.Errorf("remove iptables backup: %w", err)
	}
	return nil
}

func (c Controller) stopTor() error {
	active, err := c.torService().IsActive()
	if err != nil {
		return fmt.Errorf("check Tor service: %w", err)
	}
	if active {
		if err := c.torService().Stop(); err != nil {
			return fmt.Errorf("stop Tor service: %w", err)
		}
	}
	if err := c.run(c.killallPath(), "tor"); err != nil {
		c.warnf("Tor process cleanup failed: %v", err)
	}
	return nil
}

func (c Controller) restoreTorrc() error {
	return c.run(c.makeTorrcPath(), "restore")
}
