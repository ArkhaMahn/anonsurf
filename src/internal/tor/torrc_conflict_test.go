package tor

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestTorrcManagerRejectsConflictingBackupState(t *testing.T) {
	manager := testTorrcManager(t)
	for _, path := range []string{manager.BackupPath, manager.MissingMarkerPath} {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("state\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}

	if err := manager.Backup(); !errors.Is(err, ErrTorrcBackupConflict) {
		t.Fatalf("Backup() error = %v, want %v", err, ErrTorrcBackupConflict)
	}
	if err := manager.Restore(); !errors.Is(err, ErrTorrcBackupConflict) {
		t.Fatalf("Restore() error = %v, want %v", err, ErrTorrcBackupConflict)
	}
}
