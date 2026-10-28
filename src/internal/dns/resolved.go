package dns

import (
	"fmt"
	"strings"

	"gitlab.com/parrotsec/packages/anonsurf/internal/command"
	"gitlab.com/parrotsec/packages/anonsurf/internal/fileops"
)

const (
	// ResolvedUnit is the systemd-resolved service used by Arch, CachyOS and
	// most modern distributions to serve glibc DNS lookups through nss-resolve.
	ResolvedUnit = "systemd-resolved.service"
	// ResolvedStatePath records that AnonSurf suspended systemd-resolved and
	// must start it again when routing stops.
	ResolvedStatePath = "/var/lib/anonsurf/systemd-resolved.was-active"

	systemctlPath = "/usr/bin/systemctl"
)

// SuspendResolved stops systemd-resolved for the lifetime of the AnonSurf
// session.
//
// nss-resolve answers glibc lookups directly from systemd-resolved and ignores
// /etc/resolv.conf, so writing Tor's DNSPort into resolv.conf is not enough on
// systems that enable it. Stopping the service makes the nsswitch "resolve"
// module unavailable, so lookups fall through to the Tor-backed resolver.
//
// The unit is also masked at runtime because it is socket- and D-Bus-activated:
// without the mask the next lookup through its varlink socket would start it
// again. A runtime mask only lives until the next boot, and the condition
// marker lets the stop path and boot recovery remove it again.
//
// The previous state is recorded before the service is touched, which lets the
// stop path and boot recovery restore it. A failure to stop the unit is not
// fatal: transparent redirect rules still send UDP/53 from resolved to Tor.
func (Manager) SuspendResolved() error {
	active, err := resolvedActive()
	if err != nil {
		return err
	}
	if !active {
		return nil
	}
	if err := fileops.WriteAtomicWithDir(ResolvedStatePath, []byte("active\n"), 0600, 0700); err != nil {
		return fmt.Errorf("record systemd-resolved state: %w", err)
	}
	if err := runSystemctl("mask", "--runtime", ResolvedUnit); err != nil {
		return fmt.Errorf("mask %s: %w", ResolvedUnit, err)
	}
	if err := runSystemctl("stop", ResolvedUnit); err != nil {
		return fmt.Errorf("stop %s: %w", ResolvedUnit, err)
	}
	return nil
}

// RestoreResolved starts systemd-resolved again when AnonSurf suspended it.
// It is safe to call when no session suspended the service.
func (Manager) RestoreResolved() error {
	exists, err := fileops.Exists(ResolvedStatePath)
	if err != nil {
		return fmt.Errorf("check systemd-resolved state: %w", err)
	}
	if !exists {
		return nil
	}
	// Unmask first: a runtime mask (possibly left over from this or an
	// interrupted session) would otherwise block the start below.
	if err := runSystemctl("unmask", "--runtime", ResolvedUnit); err != nil {
		return fmt.Errorf("unmask %s: %w", ResolvedUnit, err)
	}
	if err := runSystemctl("start", ResolvedUnit); err != nil {
		return fmt.Errorf("start %s: %w", ResolvedUnit, err)
	}
	if err := fileops.RemoveIfExists(ResolvedStatePath); err != nil {
		return fmt.Errorf("remove systemd-resolved state: %w", err)
	}
	return nil
}

func resolvedActive() (bool, error) {
	output, err := (command.Runner{}).Output(systemctlPath, "is-active", ResolvedUnit)
	return parseActiveState(string(output), err != nil)
}

func runSystemctl(args ...string) error {
	return (command.Runner{}).Run(systemctlPath, args...)
}

// parseActiveState interprets `systemctl is-active` output. systemctl exits
// non-zero for inactive units, so a known state wins over the exit status.
func parseActiveState(output string, failed bool) (bool, error) {
	state := strings.TrimSpace(output)
	switch state {
	case "active", "activating", "reloading":
		return true, nil
	case "inactive", "failed", "deactivating", "maintenance", "unknown":
		return false, nil
	case "":
		if failed {
			// Missing unit or no systemd: nothing to suspend.
			return false, nil
		}
	}
	if failed {
		return false, fmt.Errorf("unexpected systemctl is-active result %q", state)
	}
	return false, fmt.Errorf("unexpected systemctl is-active state %q", state)
}
