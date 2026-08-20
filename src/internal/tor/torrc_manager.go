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

type TorrcManager struct{}

func (TorrcManager) Backup() error {
	backupExists, err := fileExists(TorrcBackupPath)
	if err != nil {
		return fmt.Errorf("check torrc backup: %w", err)
	}
	markerExists, err := fileExists(TorrcMissingMarkerPath)
	if err != nil {
		return fmt.Errorf("check missing torrc marker: %w", err)
	}
	if backupExists && markerExists {
		return ErrTorrcBackupConflict
	}
	if backupExists || markerExists {
		return ErrTorrcBackupExists
	}

	info, err := os.Stat(TorrcPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(TorrcMissingMarkerPath), 0700); err != nil {
				return fmt.Errorf("create torrc state directory: %w", err)
			}
			if err := writeFileAtomically(TorrcMissingMarkerPath, []byte("missing\n"), 0600); err != nil {
				return fmt.Errorf("write missing torrc marker: %w", err)
			}
			return nil
		}
		return fmt.Errorf("stat torrc: %w", err)
	}
	content, err := os.ReadFile(TorrcPath)
	if err != nil {
		return fmt.Errorf("read torrc: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(TorrcBackupPath), 0700); err != nil {
		return fmt.Errorf("create torrc backup directory: %w", err)
	}
	if err := writeFileAtomically(TorrcBackupPath, content, info.Mode().Perm()); err != nil {
		return fmt.Errorf("write torrc backup: %w", err)
	}
	return nil
}

func (TorrcManager) Restore() error {
	backupExists, err := fileExists(TorrcBackupPath)
	if err != nil {
		return fmt.Errorf("check torrc backup: %w", err)
	}
	markerExists, err := fileExists(TorrcMissingMarkerPath)
	if err != nil {
		return fmt.Errorf("check missing torrc marker: %w", err)
	}
	if backupExists && markerExists {
		return ErrTorrcBackupConflict
	}

	if backupExists {
		info, err := os.Stat(TorrcBackupPath)
		if err != nil {
			return fmt.Errorf("stat torrc backup: %w", err)
		}
		content, err := os.ReadFile(TorrcBackupPath)
		if err != nil {
			return fmt.Errorf("read torrc backup: %w", err)
		}
		if err := writeFileAtomically(TorrcPath, content, info.Mode().Perm()); err != nil {
			return fmt.Errorf("restore torrc backup: %w", err)
		}
		if err := os.Remove(TorrcBackupPath); err != nil {
			return fmt.Errorf("remove torrc backup: %w", err)
		}
		return nil
	}
	if !markerExists {
		return ErrTorrcBackupNotFound
	}
	if err := os.Remove(TorrcPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove generated torrc: %w", err)
	}
	if err := os.Remove(TorrcMissingMarkerPath); err != nil {
		return fmt.Errorf("remove missing torrc marker: %w", err)
	}
	return nil
}

func (TorrcManager) Write(content string) error {
	if strings.TrimSpace(content) == "" {
		return ErrEmptyTorrc
	}
	if err := os.MkdirAll(filepath.Dir(TorrcPath), 0755); err != nil {
		return fmt.Errorf("create torrc directory: %w", err)
	}
	if err := writeFileAtomically(TorrcPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("write torrc: %w", err)
	}
	return nil
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
