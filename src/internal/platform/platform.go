// Package platform holds runtime detection for distribution differences that
// affect how AnonSurf drives system services.
package platform

import (
	"os"
	"path/filepath"
	"sync"
)

// systemd unit search path, highest priority first.
var unitDirs = []string{
	"/etc/systemd/system",
	"/run/systemd/system",
	"/usr/local/lib/systemd/system",
	"/usr/lib/systemd/system",
	"/lib/systemd/system",
}

var (
	torUnitOnce sync.Once
	torUnit     string
)

// TorUnit returns the systemd unit that runs Tor on this system.
//
// Debian and Parrot expose the default instance as tor@default.service while
// Arch, CachyOS and other distributions ship a plain tor.service.
func TorUnit() string {
	torUnitOnce.Do(func() {
		torUnit = detectTorUnit()
	})
	return torUnit
}

func detectTorUnit() string {
	if unitExists("tor@default.service") {
		return "tor@default.service"
	}
	if unitExists("tor.service") {
		return "tor.service"
	}
	return "tor@default.service"
}

// TorRunsUnderForegroundUnit reports whether Tor is supervised by a unit that
// already keeps the daemon in the foreground, which is the case for Arch's
// tor.service. Debian's tor@default.service passes --RunAsDaemon 0 on the
// command line, so its generated torrc may keep the upstream default.
func TorRunsUnderForegroundUnit() bool {
	return TorUnit() == "tor.service"
}

func unitExists(name string) bool {
	for _, dir := range unitDirs {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			return true
		}
	}
	return false
}
