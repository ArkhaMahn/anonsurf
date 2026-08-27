package dns

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gitlab.com/parrotsec/packages/anonsurf/internal/fileops"
)

const (
	NMHookScript               = "/etc/NetworkManager/conf.d/90-dns-none.conf"
	NMHookBackup               = "/var/lib/anonsurf/network-manager-hook.backup"
	NMHookMissingMarker        = "/var/lib/anonsurf/network-manager-hook.missing"
	NMHookDontManageResolvConf = "[main]\ndns=none\nrc-manager=unmanaged\n"
	NMHookHandleResolvConf     = "[main]\ndns=default\nrc-manager=symlink\n"
)

func reloadNetworkManager() error {
	cmd := exec.Command("/usr/bin/nmcli", "general", "reload")
	return cmd.Run()
}

func NMGetDNSServers() ([]string, error) {
	out, err := exec.Command("/usr/bin/nmcli", "-g", "IP4.DNS,IP6.DNS", "dev", "show").Output()
	if err != nil {
		return nil, err
	}
	return ParseNetworkManagerDNSOutput(string(out)), nil
}

func ParseNetworkManagerDNSOutput(output string) []string {
	var dnsServers []string
	for _, field := range strings.FieldsFunc(output, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r' || r == '\t' || r == ' '
	}) {
		dnsServers = append(dnsServers, field)
	}
	return GetListValidAddresses(dnsServers)
}

func (m Manager) writeNetworkManagerHook(content string, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(NMHookScript), 0755); err != nil {
		return err
	}
	if err := fileops.WriteAtomic(NMHookScript, []byte(content), perm); err != nil {
		return err
	}
	return reloadNetworkManager()
}

func (m Manager) createNetworkManagerHook(content string) error {
	return m.writeNetworkManagerHook(content, 0644)
}

func (m Manager) backupNetworkManagerHook() error {
	state, err := m.networkManagerHookState()
	if err != nil {
		return err
	}
	if state != networkManagerHookUnmanaged {
		return nil
	}

	info, err := os.Stat(NMHookScript)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(NMHookMissingMarker), 0700); err != nil {
				return err
			}
			return fileops.WriteAtomic(NMHookMissingMarker, []byte("missing\n"), 0600)
		}
		return err
	}
	content, err := os.ReadFile(NMHookScript)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(NMHookBackup), 0700); err != nil {
		return err
	}
	return fileops.WriteAtomic(NMHookBackup, content, info.Mode().Perm())
}

type networkManagerHookState int

const (
	networkManagerHookUnmanaged networkManagerHookState = iota
	networkManagerHookBackedUp
	networkManagerHookMissing
)

func (m Manager) networkManagerHookState() (networkManagerHookState, error) {
	backupExists, err := fileops.Exists(NMHookBackup)
	if err != nil {
		return networkManagerHookUnmanaged, err
	}
	markerExists, err := fileops.Exists(NMHookMissingMarker)
	if err != nil {
		return networkManagerHookUnmanaged, err
	}
	if backupExists && markerExists {
		return networkManagerHookUnmanaged, fmt.Errorf("network manager hook backup state is inconsistent")
	}
	if backupExists {
		return networkManagerHookBackedUp, nil
	}
	if markerExists {
		return networkManagerHookMissing, nil
	}
	return networkManagerHookUnmanaged, nil
}

func (m Manager) restoreNetworkManagerHook() (networkManagerHookState, error) {
	state, err := m.networkManagerHookState()
	if err != nil {
		return state, err
	}

	switch state {
	case networkManagerHookBackedUp:
		info, err := os.Stat(NMHookBackup)
		if err != nil {
			return state, err
		}
		content, err := os.ReadFile(NMHookBackup)
		if err != nil {
			return state, err
		}
		if err := m.writeNetworkManagerHook(string(content), info.Mode().Perm()); err != nil {
			return state, err
		}
	case networkManagerHookMissing:
		if err := fileops.RemoveIfExists(NMHookScript); err != nil {
			return state, err
		}
		if err := reloadNetworkManager(); err != nil {
			return state, err
		}
	default:
		content, readErr := os.ReadFile(NMHookScript)
		switch {
		case readErr == nil && string(content) == NMHookDontManageResolvConf:
			if err := m.createNetworkManagerHook(NMHookHandleResolvConf); err != nil {
				return state, err
			}
		case readErr == nil || os.IsNotExist(readErr):
			if err := reloadNetworkManager(); err != nil {
				return state, err
			}
		default:
			return state, readErr
		}
	}
	return state, nil
}

func (m Manager) clearNetworkManagerHookState(state networkManagerHookState) error {
	switch state {
	case networkManagerHookBackedUp:
		return fileops.RemoveIfExists(NMHookBackup)
	case networkManagerHookMissing:
		return fileops.RemoveIfExists(NMHookMissingMarker)
	default:
		return nil
	}
}
