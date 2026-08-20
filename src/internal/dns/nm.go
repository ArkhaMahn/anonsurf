package dns

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	NMHookScript               = "/etc/NetworkManager/conf.d/90-dns-none.conf"
	NMHookBackup               = "/var/lib/anonsurf/network-manager-hook.backup"
	NMHookMissingMarker        = "/var/lib/anonsurf/network-manager-hook.missing"
	NMHookDontManageResolvConf = "[main]\ndns=none\nrc-manager=unmanaged\n"
	NMHookHandleResolvConf     = "[main]\ndns=default\nrc-manager=symlink\n"
)

func NMReloadService() error {
	cmd := exec.Command("/usr/bin/systemctl", "reload", "NetworkManager")
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
	if err := os.MkdirAll(filepath.Dir(m.NetworkManagerHookPath), 0755); err != nil {
		return err
	}
	if err := writeFileAtomically(m.NetworkManagerHookPath, []byte(content), perm); err != nil {
		return err
	}
	if m.ReloadNetworkManager == nil {
		return nil
	}
	return m.ReloadNetworkManager()
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

	info, err := os.Stat(m.NetworkManagerHookPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(m.networkManagerHookMissingMarker()), 0700); err != nil {
				return err
			}
			return writeFileAtomically(m.networkManagerHookMissingMarker(), []byte("missing\n"), 0600)
		}
		return err
	}
	content, err := os.ReadFile(m.NetworkManagerHookPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(m.networkManagerHookBackup()), 0700); err != nil {
		return err
	}
	return writeFileAtomically(m.networkManagerHookBackup(), content, info.Mode().Perm())
}

type networkManagerHookState int

const (
	networkManagerHookUnmanaged networkManagerHookState = iota
	networkManagerHookBackedUp
	networkManagerHookMissing
)

func (m Manager) networkManagerHookState() (networkManagerHookState, error) {
	backupExists, err := fileExists(m.networkManagerHookBackup())
	if err != nil {
		return networkManagerHookUnmanaged, err
	}
	markerExists, err := fileExists(m.networkManagerHookMissingMarker())
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
		info, err := os.Stat(m.networkManagerHookBackup())
		if err != nil {
			return state, err
		}
		content, err := os.ReadFile(m.networkManagerHookBackup())
		if err != nil {
			return state, err
		}
		if err := m.writeNetworkManagerHook(string(content), info.Mode().Perm()); err != nil {
			return state, err
		}
	case networkManagerHookMissing:
		if err := removeFileIfExists(m.NetworkManagerHookPath); err != nil {
			return state, err
		}
		if m.ReloadNetworkManager != nil {
			if err := m.ReloadNetworkManager(); err != nil {
				return state, err
			}
		}
	default:
		if err := m.createNetworkManagerHook(NMHookHandleResolvConf); err != nil {
			return state, err
		}
	}
	return state, nil
}

func (m Manager) clearNetworkManagerHookState(state networkManagerHookState) error {
	switch state {
	case networkManagerHookBackedUp:
		return removeFileIfExists(m.networkManagerHookBackup())
	case networkManagerHookMissing:
		return removeFileIfExists(m.networkManagerHookMissingMarker())
	default:
		return nil
	}
}

func (m Manager) networkManagerHookBackup() string {
	if m.NetworkManagerHookBackup != "" {
		return m.NetworkManagerHookBackup
	}
	return NMHookBackup
}

func (m Manager) networkManagerHookMissingMarker() string {
	if m.NetworkManagerHookMissingMarker != "" {
		return m.NetworkManagerHookMissingMarker
	}
	return NMHookMissingMarker
}
