package tor

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	TorrcPath              = "/etc/tor/torrc"
	TorrcBackupPath        = "/var/lib/anonsurf/torrc.backup"
	TorrcMissingMarkerPath = "/var/lib/anonsurf/torrc.missing"
)

var (
	ErrEmptyTorrc          = errors.New("torrc content is empty")
	ErrTorrcBackupExists   = errors.New("torrc backup already exists")
	ErrTorrcBackupNotFound = errors.New("torrc backup not found")
	ErrTorrcBackupConflict = errors.New("torrc backup state is inconsistent")
)

type TorrcManager struct {
	Path              string
	BackupPath        string
	MissingMarkerPath string
}

func NewTorrcManager() TorrcManager {
	return TorrcManager{
		Path:              TorrcPath,
		BackupPath:        TorrcBackupPath,
		MissingMarkerPath: TorrcMissingMarkerPath,
	}
}

func (m TorrcManager) Backup() error {
	backupExists, err := fileExists(m.BackupPath)
	if err != nil {
		return fmt.Errorf("check torrc backup: %w", err)
	}
	markerPath := m.missingMarkerPath()
	markerExists, err := fileExists(markerPath)
	if err != nil {
		return fmt.Errorf("check missing torrc marker: %w", err)
	}
	if backupExists && markerExists {
		return ErrTorrcBackupConflict
	}
	if backupExists || markerExists {
		return ErrTorrcBackupExists
	}

	info, err := os.Stat(m.Path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(markerPath), 0700); err != nil {
				return fmt.Errorf("create torrc state directory: %w", err)
			}
			if err := writeFileAtomically(markerPath, []byte("missing\n"), 0600); err != nil {
				return fmt.Errorf("write missing torrc marker: %w", err)
			}
			return nil
		}
		return fmt.Errorf("stat torrc: %w", err)
	}
	content, err := os.ReadFile(m.Path)
	if err != nil {
		return fmt.Errorf("read torrc: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(m.BackupPath), 0700); err != nil {
		return fmt.Errorf("create torrc backup directory: %w", err)
	}
	if err := writeFileAtomically(m.BackupPath, content, info.Mode().Perm()); err != nil {
		return fmt.Errorf("write torrc backup: %w", err)
	}
	return nil
}

func (m TorrcManager) Restore() error {
	backupExists, err := fileExists(m.BackupPath)
	if err != nil {
		return fmt.Errorf("check torrc backup: %w", err)
	}
	markerPath := m.missingMarkerPath()
	markerExists, err := fileExists(markerPath)
	if err != nil {
		return fmt.Errorf("check missing torrc marker: %w", err)
	}
	if backupExists && markerExists {
		return ErrTorrcBackupConflict
	}

	if backupExists {
		info, err := os.Stat(m.BackupPath)
		if err != nil {
			return fmt.Errorf("stat torrc backup: %w", err)
		}
		content, err := os.ReadFile(m.BackupPath)
		if err != nil {
			return fmt.Errorf("read torrc backup: %w", err)
		}
		if err := writeFileAtomically(m.Path, content, info.Mode().Perm()); err != nil {
			return fmt.Errorf("restore torrc backup: %w", err)
		}
		if err := os.Remove(m.BackupPath); err != nil {
			return fmt.Errorf("remove torrc backup: %w", err)
		}
		return nil
	}
	if !markerExists {
		return ErrTorrcBackupNotFound
	}
	if err := os.Remove(m.Path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove generated torrc: %w", err)
	}
	if err := os.Remove(markerPath); err != nil {
		return fmt.Errorf("remove missing torrc marker: %w", err)
	}
	return nil
}

func (m TorrcManager) Write(content string) error {
	if strings.TrimSpace(content) == "" {
		return ErrEmptyTorrc
	}
	if err := os.MkdirAll(filepath.Dir(m.Path), 0755); err != nil {
		return fmt.Errorf("create torrc directory: %w", err)
	}
	if err := writeFileAtomically(m.Path, []byte(content), 0600); err != nil {
		return fmt.Errorf("write torrc: %w", err)
	}
	return nil
}

func (m TorrcManager) missingMarkerPath() string {
	if m.MissingMarkerPath != "" {
		return m.MissingMarkerPath
	}
	return TorrcMissingMarkerPath
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func writeFileAtomically(path string, content []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tempFile, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	if _, err := tempFile.Write(content); err != nil {
		_ = tempFile.Close()
		return err
	}
	if err := tempFile.Chmod(perm); err != nil {
		_ = tempFile.Close()
		return err
	}
	if err := tempFile.Sync(); err != nil {
		_ = tempFile.Close()
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}

	return os.Rename(tempPath, path)
}
