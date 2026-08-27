package anonsurf

import (
	"errors"
	"fmt"

	"gitlab.com/parrotsec/packages/anonsurf/internal/dns"
)

const torDNSAddress = "127.0.0.1"

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

func isSkippableDNSBackupError(err error) bool {
	return errors.Is(err, dns.ErrBackupExists) || errors.Is(err, dns.ErrBackupSkippedSymlink)
}
