package dns

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	ResolvConfBackup = "/etc/anonsurf/resolv.conf.bak"
	ResolvConfPath   = "/etc/resolv.conf"
)

var (
	ErrBackupSkippedLocalhost = errors.New("skip backup because resolv.conf only points to localhost")
	ErrBackupSkippedSymlink   = errors.New("skip backup because resolv.conf is a symlink")
)

type Manager struct{}

func removeFileIfExists(path string) error {
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
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

func writeFileAtomically(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func copyFile(src, dst string, perm os.FileMode) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return writeFileAtomically(dst, data, perm)
}

func (m Manager) CreateDHCPOnly() error {
	if err := m.backupNetworkManagerHook(); err != nil {
		return fmt.Errorf("backup NetworkManager DNS settings: %w", err)
	}
	if err := removeFileIfExists(ResolvConfPath); err != nil {
		return fmt.Errorf("remove current DNS settings: %w", err)
	}
	return m.createNetworkManagerHook(NMHookHandleResolvConf)
}

func (m Manager) CreateAddrs(listAddr []string) error {
	if len(listAddr) == 0 {
		return ErrNoValidDNSAddress
	}
	if err := m.backupNetworkManagerHook(); err != nil {
		return fmt.Errorf("backup NetworkManager DNS settings: %w", err)
	}
	if err := WriteDNSToSystem(listAddr, ResolvConfPath); err != nil {
		return fmt.Errorf("write DNS settings: %w", err)
	}
	return m.createNetworkManagerHook(NMHookDontManageResolvConf)
}

func (m Manager) restoreResolverBackup() (bool, error) {
	exists, err := fileExists(ResolvConfBackup)
	if err != nil {
		return false, fmt.Errorf("check backup: %w", err)
	}
	if !exists {
		return false, nil
	}
	info, err := os.Stat(ResolvConfBackup)
	if err != nil {
		return false, fmt.Errorf("stat backup: %w", err)
	}
	content, err := os.ReadFile(ResolvConfBackup)
	if err != nil {
		return false, fmt.Errorf("read backup: %w", err)
	}
	if err := writeFileAtomically(ResolvConfPath, content, info.Mode().Perm()); err != nil {
		return false, fmt.Errorf("restore backup: %w", err)
	}
	return true, nil
}

func (m Manager) RestoreBackupOrDHCP() error {
	restoredBackup, err := m.restoreResolverBackup()
	if err != nil {
		return err
	}
	if !restoredBackup {
		if err := removeFileIfExists(ResolvConfPath); err != nil {
			return fmt.Errorf("remove current DNS settings: %w", err)
		}
	}

	hookState, err := m.restoreNetworkManagerHook()
	if err != nil {
		return fmt.Errorf("restore NetworkManager DNS ownership: %w", err)
	}
	if restoredBackup {
		if err := removeFileIfExists(ResolvConfBackup); err != nil {
			return fmt.Errorf("remove restored backup: %w", err)
		}
	}
	if err := m.clearNetworkManagerHookState(hookState); err != nil {
		return fmt.Errorf("clear NetworkManager hook backup: %w", err)
	}
	return nil
}

func (m Manager) MakeBackup() error {
	exists, err := fileExists(ResolvConfPath)
	if err != nil {
		return fmt.Errorf("check resolv.conf: %w", err)
	}
	if !exists {
		return fmt.Errorf("missing %s", ResolvConfPath)
	}

	hasOnlyLocalhost, err := SystemHasOnlyLocalhost(ResolvConfPath)
	if err != nil {
		return fmt.Errorf("read resolv.conf: %w", err)
	}
	if hasOnlyLocalhost {
		return ErrBackupSkippedLocalhost
	}

	isSymlink, err := SystemDNSFileIsSymlink(ResolvConfPath)
	if err != nil {
		return fmt.Errorf("check resolv.conf symlink: %w", err)
	}
	if isSymlink {
		return ErrBackupSkippedSymlink
	}

	info, err := os.Stat(ResolvConfPath)
	if err != nil {
		return fmt.Errorf("stat resolv.conf: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(ResolvConfBackup), 0755); err != nil {
		return fmt.Errorf("create backup directory: %w", err)
	}
	if err := copyFile(ResolvConfPath, ResolvConfBackup, info.Mode().Perm()); err != nil {
		return fmt.Errorf("create backup: %w", err)
	}
	return nil
}

func (m Manager) DHCPAddresses() ([]string, error) {
	addresses, err := NMGetDNSServers()
	if err != nil {
		return nil, err
	}
	return GetListValidAddresses(addresses), nil
}
