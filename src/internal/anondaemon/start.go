package anondaemon

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gitlab.com/parrotsec/packages/anonsurf/internal/dns"
	"gitlab.com/parrotsec/packages/anonsurf/internal/tor"
)

const torDNSAddress = "127.0.0.1"

func (c Controller) Start() error {
	if err := c.generateTorrc(); err != nil {
		return fmt.Errorf("generate torrc: %w", err)
	}
	if exists, err := c.fileExists(c.torrcPath()); err != nil {
		return fmt.Errorf("check torrc: %w", err)
	} else if !exists {
		return ErrTorrcNotFound
	}

	if err := c.startSurf(); err != nil {
		return err
	}

	active, err := c.torService().IsActive()
	if err != nil {
		return fmt.Errorf("check Tor service: %w", err)
	}
	if !active {
		if err := c.torService().Start(); err != nil {
			return fmt.Errorf("start Tor service: %w", err)
		}
	}

	c.prepareControlCookie()
	return nil
}

func (c Controller) readTorRoutingConfig() (tor.RoutingConfig, error) {
	data, err := c.readFile(c.torrcPath())
	if err != nil {
		return tor.RoutingConfig{}, fmt.Errorf("read torrc: %w", err)
	}
	return tor.ParseRoutingConfig(strings.NewReader(string(data)))
}

func (c Controller) startSurf() error {
	active, err := c.torService().IsActive()
	if err != nil {
		return fmt.Errorf("check Tor service before routing: %w", err)
	}
	if active {
		if err := c.torService().Stop(); err != nil {
			return fmt.Errorf("stop Tor service before routing: %w", err)
		}
	}

	routing, err := c.readTorRoutingConfig()
	if err != nil {
		return err
	}

	if err := c.dnsManager().MakeBackup(); err != nil && !isSkippableDNSBackupError(err) {
		return fmt.Errorf("backup DNS settings: %w", err)
	}
	if err := c.dnsManager().CreateAddrs([]string{torDNSAddress}); err != nil {
		return fmt.Errorf("set Tor DNS: %w", err)
	}
	if err := c.disableIPv6(); err != nil {
		return err
	}
	if err := c.saveIPTablesRules(); err != nil {
		return err
	}
	return c.applyIPTablesRules(routing)
}

func (c Controller) disableIPv6() error {
	for _, setting := range []string{
		"net.ipv6.conf.all.disable_ipv6=1",
		"net.ipv6.conf.default.disable_ipv6=1",
	} {
		if err := c.run(c.sysctlPath(), "-w", setting); err != nil {
			return fmt.Errorf("disable IPv6: %w", err)
		}
	}
	return nil
}

func isSkippableDNSBackupError(err error) bool {
	return errors.Is(err, dns.ErrBackupSkippedLocalhost) || errors.Is(err, dns.ErrBackupSkippedSymlink)
}

func (c Controller) prepareControlCookie() {
	cookiePath := c.torCookiePath()
	exists, err := c.waitForFile(cookiePath, c.cookieTimeout(), c.cookiePollInterval())
	if err != nil {
		c.warnf("Tor control cookie check failed for %s: %v", cookiePath, err)
		return
	}
	if !exists {
		c.warnf("Tor control cookie %s was not created; identity changes may be unavailable", cookiePath)
		return
	}
	if err := c.chmodReadable(cookiePath); err != nil {
		c.warnf("Tor control cookie permissions were not updated for %s: %v", cookiePath, err)
	}
}

func (c Controller) waitForFile(path string, timeout, interval time.Duration) (bool, error) {
	exists, err := c.fileExists(path)
	if exists || err != nil || timeout <= 0 {
		return exists, err
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		c.sleep(interval)
		exists, err = c.fileExists(path)
		if exists || err != nil {
			return exists, err
		}
	}
	return false, nil
}
