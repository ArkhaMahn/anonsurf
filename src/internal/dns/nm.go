package dns

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	NMHookScript               = "/etc/NetworkManager/conf.d/90-dns-none.conf"
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

func (m Manager) createNetworkManagerHook(content string) error {
	if err := os.MkdirAll(filepath.Dir(m.NetworkManagerHookPath), 0755); err != nil {
		return err
	}
	if err := writeFileAtomically(m.NetworkManagerHookPath, []byte(content), 0644); err != nil {
		return err
	}
	if m.ReloadNetworkManager == nil {
		return nil
	}
	return m.ReloadNetworkManager()
}

func NMCreateHookUnmanage() error {
	return NewManager().createNetworkManagerHook(NMHookDontManageResolvConf)
}

func NMCreateHookDHCP() error {
	return NewManager().createNetworkManagerHook(NMHookHandleResolvConf)
}
