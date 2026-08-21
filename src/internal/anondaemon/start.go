package anondaemon

import (
	"errors"
	"fmt"
	"strings"

	"gitlab.com/parrotsec/packages/anonsurf/internal/dns"
	"gitlab.com/parrotsec/packages/anonsurf/internal/systemd"
	"gitlab.com/parrotsec/packages/anonsurf/internal/tor"
)

const torDNSAddress = "127.0.0.1"

func (c Controller) Start() error {
	state := startTransition{}
	active, err := systemd.NewTorController().IsActive()
	if err != nil {
		return fmt.Errorf("check Tor service: %w", err)
	}
	state.torWasActive, err = captureTorState(active)
	if err != nil {
		return err
	}

	state.torrcGenerated = true
	if err := c.generateTorrc(); err != nil {
		return c.failStart(fmt.Errorf("generate torrc: %w", err), state)
	}

	if exists, err := fileExists(TorrcPath); err != nil {
		return c.failStart(fmt.Errorf("check torrc: %w", err), state)
	} else if !exists {
		return c.failStart(ErrTorrcNotFound, state)
	}
	routing, err := c.readTorRoutingConfig()
	if err != nil {
		return c.failStart(err, state)
	}
	if err := c.startSurf(routing, &state); err != nil {
		return c.failStart(err, state)
	}

	state.torStartAttempted = true
	if err := systemd.NewTorController().Start(); err != nil {
		return c.failStart(fmt.Errorf("start Tor service: %w", err), state)
	}
	if !routing.PreserveSystemDNS {
		if err := c.configureDNS(&state); err != nil {
			return c.failStart(err, state)
		}
	}
	if err := c.markActive(); err != nil {
		return c.failStart(fmt.Errorf("mark daemon active: %w", err), state)
	}
	return nil
}

func (c Controller) startSurf(routing tor.RoutingConfig, state *startTransition) error {
	if err := systemd.NewTorController().Stop(); err != nil {
		return fmt.Errorf("stop Tor service before routing: %w", err)
	}

	if err := c.captureIPv6State(); err != nil {
		return fmt.Errorf("capture IPv6 state: %w", err)
	}
	state.ipv6Touched = true
	if err := c.disableIPv6(); err != nil {
		return err
	}

	if err := c.saveIPTablesRules(); err != nil {
		return err
	}
	state.firewallTouched = true
	return c.applyIPTablesRules(routing)
}

func (c Controller) configureDNS(state *startTransition) error {
	if err := (dns.Manager{}).MakeBackup(); err != nil && !isSkippableDNSBackupError(err) {
		return fmt.Errorf("backup DNS settings: %w", err)
	}
	if err := c.markDNSChanged(); err != nil {
		return fmt.Errorf("mark DNS state: %w", err)
	}
	state.dnsTouched = true
	if err := (dns.Manager{}).CreateAddrs([]string{torDNSAddress}); err != nil {
		return fmt.Errorf("set Tor DNS: %w", err)
	}
	return nil
}

func (c Controller) readTorRoutingConfig() (tor.RoutingConfig, error) {
	data, err := readFile(TorrcPath)
	if err != nil {
		return tor.RoutingConfig{}, fmt.Errorf("read torrc: %w", err)
	}
	return tor.ParseRoutingConfig(strings.NewReader(string(data)))
}

func isSkippableDNSBackupError(err error) bool {
	return errors.Is(err, dns.ErrBackupSkippedLocalhost) || errors.Is(err, dns.ErrBackupSkippedSymlink)
}
