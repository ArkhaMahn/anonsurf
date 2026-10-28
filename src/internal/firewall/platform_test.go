package firewall

import (
	"path/filepath"
	"testing"
)

func TestTorUsernameIsAKnownCandidate(t *testing.T) {
	got := torUsername()

	known := false
	for _, candidate := range torUserCandidates {
		if got == candidate {
			known = true
			break
		}
	}
	if !known {
		t.Fatalf("torUsername() returned unknown account %q", got)
	}
}

func TestIptablesPathKeepsName(t *testing.T) {
	got := iptablesPath("iptables-save")

	if filepath.Base(got) != "iptables-save" {
		t.Fatalf("expected path ending in iptables-save, got %q", got)
	}
}
