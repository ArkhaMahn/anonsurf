package platform

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestDetectTorUnitPrefersDefaultInstance(t *testing.T) {
	dir := t.TempDir()
	writeUnit(t, dir, "tor@default.service")
	writeUnit(t, dir, "tor.service")

	restore := overrideUnitDirs(dir)
	defer restore()

	if got := detectTorUnit(); got != "tor@default.service" {
		t.Fatalf("expected tor@default.service, got %q", got)
	}
}

func TestDetectTorUnitFallsBackToArchUnit(t *testing.T) {
	dir := t.TempDir()
	writeUnit(t, dir, "tor.service")

	restore := overrideUnitDirs(dir)
	defer restore()

	if got := detectTorUnit(); got != "tor.service" {
		t.Fatalf("expected tor.service, got %q", got)
	}
	if !TorRunsUnderForegroundUnit() {
		t.Fatal("expected foreground Tor unit on Arch")
	}
}

func TestDetectTorUnitDefaultWhenMissing(t *testing.T) {
	dir := t.TempDir()

	restore := overrideUnitDirs(dir)
	defer restore()

	if got := detectTorUnit(); got != "tor@default.service" {
		t.Fatalf("expected tor@default.service fallback, got %q", got)
	}
}

func overrideUnitDirs(dir string) func() {
	previous := unitDirs
	unitDirs = []string{dir}
	torUnit = ""
	torUnitOnce = sync.Once{}
	return func() {
		unitDirs = previous
		torUnit = ""
		torUnitOnce = sync.Once{}
	}
}

func writeUnit(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte("[Unit]\n"), 0o644); err != nil {
		t.Fatalf("write unit: %v", err)
	}
}
