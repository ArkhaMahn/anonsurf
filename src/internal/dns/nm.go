package dns

import (
	"os"
	"os/exec"
	"strings"

	"gitlab.com/parrotsec/packages/anonsurf/internal/ui"
)

const (
	NMHookScript               = "/etc/NetworkManager/conf.d/90-dns-none.conf"
	NMHookDontManageResolvConf = "[main]\ndns=none\nrc-manager=unmanaged\n"
	NMHookHandleResolvConf     = "[main]\ndns=default\nrc-manager=symlink\n"
)

func nmReloadService() {
	cmd := exec.Command("/usr/bin/systemctl", "reload", "NetworkManager")
	if err := cmd.Run(); err != nil {
		ui.PrintError("Failed to reload NetworkManager service")
	}
}

func NMGetDNSServers() []string {
	// Equivalent to: /usr/bin/nmcli dev show | grep DNS
	out, err := exec.Command("/usr/bin/nmcli", "-g", "IP4.DNS,IP6.DNS", "dev", "show").Output()
	if err != nil {
		return []string{}
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 {
		return []string{}
	}

	// Assuming the last line contains the DNS servers like: IP4.DNS[1]: 192.168.58.1,1.1.1.1
	// The Nim logic did: .replace("\n", "").split(" ")[^1].split(",") which combines all lines into one?
	// It's safer to just extract all DNS servers from all matching lines.
	var dnsServers []string
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) > 1 {
			// Usually parts[1] is the IP or comma-separated IPs
			ips := strings.Split(parts[1], ",")
			dnsServers = append(dnsServers, ips...)
		}
	}
	return dnsServers
}

func NMCreateHookUnmanage() {
	err := os.WriteFile(NMHookScript, []byte(NMHookDontManageResolvConf), 0644)
	if err != nil {
		ui.PrintError("Failed to write hook script to " + NMHookScript)
		return
	}
	nmReloadService()
}

func NMCreateHookDHCP() {
	err := os.WriteFile(NMHookScript, []byte(NMHookHandleResolvConf), 0644)
	if err != nil {
		ui.PrintError("Failed to write hook script to " + NMHookScript)
		return
	}
	nmReloadService()
}
