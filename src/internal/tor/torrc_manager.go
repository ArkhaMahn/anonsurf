package tor

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	TorrcPath       = "/etc/tor/torrc"
	TorrcBackupPath = "/etc/tor/torrc.bak"
)

var (
	ErrEmptyTorrc          = errors.New("torrc content is empty")
	ErrTorrcBackupExists   = errors.New("torrc backup already exists")
	ErrTorrcBackupNotFound = errors.New("torrc backup not found")
)

type TorrcManager struct {
	Path       string
	BackupPath string
}

func NewTorrcManager() TorrcManager {
	return TorrcManager{
		Path:       TorrcPath,
		BackupPath: TorrcBackupPath,
	}
}

// Backup preserves the first backup so repeated generation cannot overwrite the original torrc.
func (m TorrcManager) Backup() error {
	if exists, err := fileExists(m.BackupPath); err != nil {
		return fmt.Errorf("check torrc backup: %w", err)
	} else if exists {
		return ErrTorrcBackupExists
	}

	info, err := os.Stat(m.Path)
	if err != nil {
		return fmt.Errorf("stat torrc: %w", err)
	}
	content, err := os.ReadFile(m.Path)
	if err != nil {
		return fmt.Errorf("read torrc: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(m.BackupPath), 0755); err != nil {
		return fmt.Errorf("create torrc backup directory: %w", err)
	}
	if err := writeFileAtomically(m.BackupPath, content, info.Mode().Perm()); err != nil {
		return fmt.Errorf("write torrc backup: %w", err)
	}
	return nil
}

// Restore moves the backup back into place, matching the legacy make-torrc restore behavior.
func (m TorrcManager) Restore() error {
	if exists, err := fileExists(m.BackupPath); err != nil {
		return fmt.Errorf("check torrc backup: %w", err)
	} else if !exists {
		return ErrTorrcBackupNotFound
	}

	if err := os.Rename(m.BackupPath, m.Path); err != nil {
		return fmt.Errorf("restore torrc backup: %w", err)
	}
	return nil
}

// Write rejects empty rendered config and replaces torrc with a same-directory atomic rename.
func (m TorrcManager) Write(content string) error {
	if strings.TrimSpace(content) == "" {
		return ErrEmptyTorrc
	}
	if err := os.MkdirAll(filepath.Dir(m.Path), 0755); err != nil {
		return fmt.Errorf("create torrc directory: %w", err)
	}
	if err := writeFileAtomically(m.Path, []byte(content), 0644); err != nil {
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
