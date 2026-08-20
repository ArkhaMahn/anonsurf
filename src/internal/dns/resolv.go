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
	ErrBackupNotFound         = errors.New("resolv.conf backup not found")
	ErrBackupSkippedLocalhost = errors.New("skip backup because resolv.conf only points to localhost")
	ErrBackupSkippedSymlink   = errors.New("skip backup because resolv.conf is a symlink")
)

type Manager struct {
	ResolvConfPath                  string
	ResolvConfBackup                string
	NetworkManagerHookPath          string
	NetworkManagerHookBackup        string
	NetworkManagerHookMissingMarker string
	ReloadNetworkManager            func() error
	NetworkManagerDNSServers        func() ([]string, error)
}

func NewManager() Manager {
	return Manager{
		ResolvConfPath:                  ResolvConfPath,
		ResolvConfBackup:                ResolvConfBackup,
		NetworkManagerHookPath:          NMHookScript,
		NetworkManagerHookBackup:        NMHookBackup,
		NetworkManagerHookMissingMarker: NMHookMissingMarker,
		ReloadNetworkManager:            NMReloadService,
		NetworkManagerDNSServers:        NMGetDNSServers,
	}
}

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
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return writeFileAtomically(dst, data, 0644)
}

func (m Manager) CreateDHCPOnly() error {
	if err := m.backupNetworkManagerHook(); err != nil {
		return fmt.Errorf("backup NetworkManager DNS settings: %w", err)
	}
	if err := removeFileIfExists(m.ResolvConfPath); err != nil {
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
	if err := WriteDNSToSystem(listAddr, m.ResolvConfPath); err != nil {
		return fmt.Errorf("write DNS settings: %w", err)
	}
	return m.createNetworkManagerHook(NMHookDontManageResolvConf)
}

func (m Manager) RestoreFromBackup() error {
	restored, err := m.restoreResolverBackup()
	if err != nil {
		return err
	}
	if !restored {
		return ErrBackupNotFound
	}
	if err := removeFileIfExists(m.ResolvConfBackup); err != nil {
		return fmt.Errorf("remove restored backup: %w", err)
	}
	return nil
}

func (m Manager) restoreResolverBackup() (bool, error) {
	exists, err := fileExists(m.ResolvConfBackup)
	if err != nil {
		return false, fmt.Errorf("check backup: %w", err)
	}
	if !exists {
		return false, nil
	}
	info, err := os.Stat(m.ResolvConfBackup)
	if err != nil {
		return false, fmt.Errorf("stat backup: %w", err)
	}
	content, err := os.ReadFile(m.ResolvConfBackup)
	if err != nil {
		return false, fmt.Errorf("read backup: %w", err)
	}
	if err := writeFileAtomically(m.ResolvConfPath, content, info.Mode().Perm()); err != nil {
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
		if err := removeFileIfExists(m.ResolvConfPath); err != nil {
			return fmt.Errorf("remove current DNS settings: %w", err)
		}
	}

	hookState, err := m.restoreNetworkManagerHook()
	if err != nil {
		return fmt.Errorf("restore NetworkManager DNS ownership: %w", err)
	}
	if restoredBackup {
		if err := removeFileIfExists(m.ResolvConfBackup); err != nil {
			return fmt.Errorf("remove restored backup: %w", err)
		}
	}
	if err := m.clearNetworkManagerHookState(hookState); err != nil {
		return fmt.Errorf("clear NetworkManager hook backup: %w", err)
	}
	return nil
}

func (m Manager) MakeBackup() error {
	exists, err := fileExists(m.ResolvConfPath)
	if err != nil {
		return fmt.Errorf("check resolv.conf: %w", err)
	}
	if !exists {
		return fmt.Errorf("missing %s", m.ResolvConfPath)
	}

	hasOnlyLocalhost, err := SystemHasOnlyLocalhost(m.ResolvConfPath)
	if err != nil {
		return fmt.Errorf("read resolv.conf: %w", err)
	}
	if hasOnlyLocalhost {
		return ErrBackupSkippedLocalhost
	}

	isSymlink, err := SystemDNSFileIsSymlink(m.ResolvConfPath)
	if err != nil {
		return fmt.Errorf("check resolv.conf symlink: %w", err)
	}
	if isSymlink {
		return ErrBackupSkippedSymlink
	}

	if err := os.MkdirAll(filepath.Dir(m.ResolvConfBackup), 0755); err != nil {
		return fmt.Errorf("create backup directory: %w", err)
	}
	if err := copyFile(m.ResolvConfPath, m.ResolvConfBackup); err != nil {
		return fmt.Errorf("create backup: %w", err)
	}
	return nil
}

func (m Manager) DHCPAddresses() ([]string, error) {
	addresses, err := m.NetworkManagerDNSServers()
	if err != nil {
		return nil, err
	}
	return GetListValidAddresses(addresses), nil
}

func ResolvCreateDHCPOnly() error {
	return NewManager().CreateDHCPOnly()
}

func ResolvCreateAddrs(listAddr []string) error {
	return NewManager().CreateAddrs(listAddr)
}

func ResolvRestoreFromBackup() error {
	return NewManager().RestoreFromBackup()
}

func ResolvMakeBackup() error {
	return NewManager().MakeBackup()
}
