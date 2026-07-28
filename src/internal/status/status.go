package status

import (
	"errors"
	"fmt"
	"os"

	"gitlab.com/parrotsec/packages/anonsurf/internal/dns"
	"gitlab.com/parrotsec/packages/anonsurf/internal/systemd"
)

type DNSState string

const (
	DNSMissing       DNSState = "missing"
	DNSEmpty         DNSState = "empty"
	DNSTor           DNSState = "tor"
	DNSLeak          DNSState = "leak"
	DNSClearnet      DNSState = "clearnet"
	DNSLocalhostOnly DNSState = "localhost-only"
)

type Snapshot struct {
	AnonSurfActive bool
	TorActive      bool
	DNSState       DNSState
	DNSAddresses   []string
}

type Reader struct {
	AnonSurfActive func() (bool, error)
	TorActive      func() (bool, error)
	DNSAddresses   func() ([]string, error)
}

func NewReader() Reader {
	anonsurfd := systemd.NewAnonsurfdController()
	torService := systemd.NewTorController()

	return Reader{
		AnonSurfActive: anonsurfd.IsActive,
		TorActive:      torService.IsActive,
		DNSAddresses: func() ([]string, error) {
			return dns.ParseDNSAddresses(dns.ResolvConfPath)
		},
	}
}

func (r Reader) Snapshot() (Snapshot, error) {
	if r.AnonSurfActive == nil {
		return Snapshot{}, fmt.Errorf("missing AnonSurf status reader")
	}
	if r.TorActive == nil {
		return Snapshot{}, fmt.Errorf("missing Tor status reader")
	}
	if r.DNSAddresses == nil {
		return Snapshot{}, fmt.Errorf("missing DNS status reader")
	}

	anonsurfActive, err := r.AnonSurfActive()
	if err != nil {
		return Snapshot{}, fmt.Errorf("read AnonSurf status: %w", err)
	}
	torActive, err := r.TorActive()
	if err != nil {
		return Snapshot{}, fmt.Errorf("read Tor status: %w", err)
	}
	addresses, err := r.DNSAddresses()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return Snapshot{}, fmt.Errorf("read DNS status: %w", err)
	}

	return Snapshot{
		AnonSurfActive: anonsurfActive,
		TorActive:      torActive,
		DNSState:       classifyDNS(anonsurfActive, addresses, err),
		DNSAddresses:   addresses,
	}, nil
}

func classifyDNS(anonsurfActive bool, addresses []string, readErr error) DNSState {
	if errors.Is(readErr, os.ErrNotExist) {
		return DNSMissing
	}
	if len(addresses) == 0 {
		return DNSEmpty
	}
	if anonsurfActive {
		if dns.HasOnlyLocalhost(addresses) {
			return DNSTor
		}
		return DNSLeak
	}
	if dns.HasOnlyLocalhost(addresses) {
		return DNSLocalhostOnly
	}
	return DNSClearnet
}
