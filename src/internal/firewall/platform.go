package firewall

import (
	"os"
	"os/user"
	"path/filepath"
)

// defaultTorUser is the account Tor runs as on Debian and Parrot. Arch-based
// systems use "tor" instead, so the account is resolved at runtime.
const defaultTorUser = "debian-tor"

var torUserCandidates = []string{"debian-tor", "tor", "_tor"}

// torUsername returns the account that owns Tor's outbound connections.
//
// The firewall lets that account bypass the transparent redirect, otherwise
// Tor's own traffic would be routed back into Tor and never reach the network.
func torUsername() string {
	for _, name := range torUserCandidates {
		if _, err := user.Lookup(name); err == nil {
			return name
		}
	}
	return defaultTorUser
}

// iptablesPath resolves an iptables helper through the merged /usr layout used
// by Arch-based systems and the split /usr/sbin layout used by Debian.
func iptablesPath(name string) string {
	candidates := []string{
		filepath.Join("/usr/sbin", name),
		filepath.Join("/usr/bin", name),
		filepath.Join("/sbin", name),
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return candidates[0]
}
