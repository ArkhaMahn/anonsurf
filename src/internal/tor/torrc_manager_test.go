package tor

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func testTorrcManager(t *testing.T) TorrcManager {
	t.Helper()

	dir := t.TempDir()
	return TorrcManager{
		Path:              filepath.Join(dir, "etc", "tor", "torrc"),
		BackupPath:        filepath.Join(dir, "etc", "tor", "torrc.bak"),
		MissingMarkerPath: filepath.Join(dir, "var", "lib", "anonsurf", "torrc.missing"),
	}
}

func TestNewTorrcManagerUsesParrotPaths(t *testing.T) {
	manager := NewTorrcManager()

	if manager.Path != TorrcPath {
		t.Fatalf("Path = %q, want %q", manager.Path, TorrcPath)
	}
	if manager.BackupPath != TorrcBackupPath {
		t.Fatalf("BackupPath = %q, want %q", manager.BackupPath, TorrcBackupPath)
	}
	if manager.MissingMarkerPath != TorrcMissingMarkerPath {
		t.Fatalf("MissingMarkerPath = %q, want %q", manager.MissingMarkerPath, TorrcMissingMarkerPath)
	}
}

func TestTorrcManagerBackupCreatesBackupAndPreservesExistingBackup(t *testing.T) {
	manager := testTorrcManager(t)
	if err := os.MkdirAll(filepath.Dir(manager.Path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manager.Path, []byte("original torrc\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(manager.Path, 0600); err != nil {
		t.Fatal(err)
	}

	if err := manager.Backup(); err != nil {
		t.Fatal(err)
	}
	backup, err := os.ReadFile(manager.BackupPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(backup) != "original torrc\n" {
		t.Fatalf("backup = %q, want original torrc", string(backup))
	}
	backupInfo, err := os.Stat(manager.BackupPath)
	if err != nil {
		t.Fatal(err)
	}
	if backupInfo.Mode().Perm() != 0600 {
		t.Fatalf("backup mode = %v, want 0600", backupInfo.Mode().Perm())
	}

	if err := os.WriteFile(manager.Path, []byte("generated torrc\n"), 0644); err != nil {
		t.Fatal(err)
	}
	err = manager.Backup()
	if !errors.Is(err, ErrTorrcBackupExists) {
		t.Fatalf("error = %v, want %v", err, ErrTorrcBackupExists)
	}
	backup, err = os.ReadFile(manager.BackupPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(backup) != "original torrc\n" {
		t.Fatalf("backup was overwritten: %q", string(backup))
	}
}

func TestTorrcManagerRestoresMissingTorrcState(t *testing.T) {
	manager := testTorrcManager(t)

	if err := manager.Backup(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(manager.MissingMarkerPath); err != nil {
		t.Fatalf("missing marker was not created: %v", err)
	}
	if err := manager.Write("generated torrc\n"); err != nil {
		t.Fatal(err)
	}
	if err := manager.Restore(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(manager.Path); !os.IsNotExist(err) {
		t.Fatalf("generated torrc still exists: %v", err)
	}
	if _, err := os.Stat(manager.MissingMarkerPath); !os.IsNotExist(err) {
		t.Fatalf("missing marker still exists: %v", err)
	}
}

func TestTorrcManagerRestoreMovesBackupOverTorrc(t *testing.T) {
	manager := testTorrcManager(t)
	if err := os.MkdirAll(filepath.Dir(manager.Path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manager.Path, []byte("generated torrc\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manager.BackupPath, []byte("original torrc\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := manager.Restore(); err != nil {
		t.Fatal(err)
	}
	restored, err := os.ReadFile(manager.Path)
	if err != nil {
		t.Fatal(err)
	}
	if string(restored) != "original torrc\n" {
		t.Fatalf("restored torrc = %q, want original torrc", string(restored))
	}
	if _, err := os.Stat(manager.BackupPath); !os.IsNotExist(err) {
		t.Fatalf("backup should be moved away, stat error = %v", err)
	}
}

func TestTorrcManagerRestoreMissingBackup(t *testing.T) {
	manager := testTorrcManager(t)

	err := manager.Restore()
	if !errors.Is(err, ErrTorrcBackupNotFound) {
		t.Fatalf("error = %v, want %v", err, ErrTorrcBackupNotFound)
	}
}

func TestTorrcManagerWriteCreatesDirectoryAndWritesAtomically(t *testing.T) {
	manager := testTorrcManager(t)

	if err := manager.Write("generated torrc\n"); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(manager.Path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "generated torrc\n" {
		t.Fatalf("torrc = %q, want generated torrc", string(content))
	}
}

func TestTorrcManagerWriteRejectsEmptyContentWithoutReplacingTorrc(t *testing.T) {
	manager := testTorrcManager(t)
	if err := os.MkdirAll(filepath.Dir(manager.Path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manager.Path, []byte("existing torrc\n"), 0644); err != nil {
		t.Fatal(err)
	}

	err := manager.Write(" \n\t")
	if !errors.Is(err, ErrEmptyTorrc) {
		t.Fatalf("error = %v, want %v", err, ErrEmptyTorrc)
	}
	content, err := os.ReadFile(manager.Path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "existing torrc\n" {
		t.Fatalf("torrc changed to %q", string(content))
	}
}
