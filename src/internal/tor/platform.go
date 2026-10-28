package tor

import (
	"strings"

	"gitlab.com/parrotsec/packages/anonsurf/internal/platform"
)

// archTorCookieAuthFile lives inside Tor's DataDirectory so that the hardened
// tor.service shipped by Arch and CachyOS can create it. Arch mounts the whole
// filesystem read-only for Tor and only allows writes to /var/lib/tor and
// /var/log/tor, while the upstream template points at /var/run/tor.
const archTorCookieAuthFile = "/var/lib/tor/control.authcookie"

// archTorDataDirectory is Tor's data directory on Arch-based systems. It is
// created by the tor package's tmpfiles entry and is the only location the
// hardened tor.service allows Tor to write to.
const archTorDataDirectory = "/var/lib/tor"

// applyPlatformDefaults adapts the upstream torrc template to the Tor unit of
// the running distribution.
func applyPlatformDefaults(rendered string) string {
	return applyTorrcPlatformDefaults(rendered, platform.TorRunsUnderForegroundUnit())
}

// applyTorrcPlatformDefaults rewrites the rendered torrc when Tor runs under a
// foreground systemd unit such as Arch's tor.service.
//
// That unit runs `tor -f /etc/tor/torrc` without --RunAsDaemon 0 (unlike
// Debian's tor@default.service) and starts Tor as root with a reduced
// capability set, so the configuration has to disable daemonizing, drop to the
// tor user explicitly and keep the control cookie inside the writable data
// directory.
func applyTorrcPlatformDefaults(rendered string, foregroundUnit bool) string {
	if !foregroundUnit {
		return rendered
	}

	lines := strings.Split(rendered, "\n")
	sawUser := false
	sawDataDirectory := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "RunAsDaemon "):
			lines[i] = "RunAsDaemon 0"
		case strings.HasPrefix(trimmed, "CookieAuthFile "):
			lines[i] = "CookieAuthFile " + archTorCookieAuthFile
		case strings.HasPrefix(trimmed, "DataDirectory "):
			sawDataDirectory = true
		case strings.HasPrefix(trimmed, "User "):
			sawUser = true
		}
	}

	out := strings.Join(lines, "\n")
	if !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	// Arch's tor.service has no defaults file and starts Tor for the root
	// user, so the data directory must be explicit before Tor drops to the
	// tor account. Otherwise Tor tries /root/.tor and fails.
	if !sawDataDirectory {
		out += "DataDirectory " + archTorDataDirectory + "\n"
	}
	// Group is obsolete in current Tor releases; the primary group of the tor
	// account is used automatically.
	if !sawUser {
		out += "User tor\n"
	}
	return out
}
