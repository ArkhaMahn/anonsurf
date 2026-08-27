package dns

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gitlab.com/parrotsec/packages/anonsurf/internal/fileops"
)

const (
	ResolvConfBackup = "/etc/anonsurf/resolv.conf.bak"
	ResolvConfPath   = "/etc/resolv.conf"
)

var (
	ErrBackupExists         = errors.New("resolv.conf backup already exists")
	ErrBackupSkippedSymlink = errors.New("skip backup because resolv.conf is a symlink")
)

type Manager struct{}

func (m Manager) CreateDHCPOnly() error {
	if err := m.backupNetworkManagerHook(); err != nil {
		return fmt.Errorf("backup NetworkManager DNS settings: %w", err)
	}
	if err := fileops.RemoveIfExists(ResolvConfPath); err != nil {
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
	exists, err := fileops.Exists(ResolvConfBackup)
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
	if err := fileops.WriteAtomic(ResolvConfPath, content, info.Mode().Perm()); err != nil {
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
		if err := removeManagedResolver(ResolvConfPath); err != nil {
			return fmt.Errorf("remove current DNS settings: %w", err)
		}
	}

	hookState, err := m.restoreNetworkManagerHook()
	if err != nil {
		return fmt.Errorf("restore NetworkManager DNS ownership: %w", err)
	}
	if restoredBackup {
		if err := fileops.RemoveIfExists(ResolvConfBackup); err != nil {
			return fmt.Errorf("remove restored backup: %w", err)
		}
	}
	if err := m.clearNetworkManagerHookState(hookState); err != nil {
		return fmt.Errorf("clear NetworkManager hook backup: %w", err)
	}
	return nil
}

func (m Manager) MakeBackup() error {
	backupExists, err := fileops.Exists(ResolvConfBackup)
	if err != nil {
		return fmt.Errorf("check resolv.conf backup: %w", err)
	}
	if backupExists {
		return ErrBackupExists
	}

	exists, err := fileops.Exists(ResolvConfPath)
	if err != nil {
		return fmt.Errorf("check resolv.conf: %w", err)
	}
	if !exists {
		return fmt.Errorf("missing %s", ResolvConfPath)
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
	if err := fileops.CopyAtomic(ResolvConfPath, ResolvConfBackup, info.Mode().Perm()); err != nil {
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
