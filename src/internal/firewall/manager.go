package firewall

import (
	"fmt"
	"os"

	"gitlab.com/parrotsec/packages/anonsurf/internal/command"
	"gitlab.com/parrotsec/packages/anonsurf/internal/fileops"
	"gitlab.com/parrotsec/packages/anonsurf/internal/tor"
)

const (
	SavePath       = "/usr/sbin/iptables-save"
	RestorePath    = "/usr/sbin/iptables-restore"
	StatePath      = "/run/anonsurf/iptables.rules"
	defaultTorUser = "debian-tor"
)

var excludedNetworks = [...]string{
	"192.168.0.0/16",
	"172.16.0.0/12",
	"10.0.0.0/8",
	"127.0.0.0/8",
}

type Manager struct{}

func (Manager) Backup() error {
	if exists, err := fileops.Exists(StatePath); err != nil {
		return fmt.Errorf("check iptables backup: %w", err)
	} else if exists {
		return nil
	}

	rules, err := (command.Runner{}).Output(SavePath)
	if err != nil {
		return fmt.Errorf("save iptables rules: %w", err)
	}
	if err := fileops.WriteAtomicWithDir(StatePath, rules, 0600, 0700); err != nil {
		return fmt.Errorf("write iptables backup: %w", err)
	}
	return nil
}

func (m Manager) Apply(routing tor.RoutingConfig) error {
	backup, err := os.ReadFile(StatePath)
	if err != nil {
		return fmt.Errorf("read iptables backup for apply: %w", err)
	}
	rules := m.renderRules(routing, backup)
	if err := (command.Runner{}).RunWithInput(RestorePath, []byte(rules)); err != nil {
		return fmt.Errorf("apply iptables rules: %w", err)
	}
	return nil
}

func (m Manager) Restore() error {
	exists, err := fileops.Exists(StatePath)
	if err != nil {
		return fmt.Errorf("check iptables backup: %w", err)
	}
	if !exists {
		return nil
	}
	return m.restoreRules()
}

func (Manager) restoreRules() error {
	rules, err := os.ReadFile(StatePath)
	if err != nil {
		return fmt.Errorf("read iptables backup: %w", err)
	}
	rules = ensureManagedIPTablesTables(rules)
	if err := (command.Runner{}).RunWithInput(RestorePath, rules); err != nil {
		return fmt.Errorf("restore iptables rules: %w", err)
	}
	if err := fileops.RemoveIfExists(StatePath); err != nil {
		return fmt.Errorf("remove iptables backup: %w", err)
	}
	return nil
}

func (Manager) natRules(routing tor.RoutingConfig) [][]string {
	torUser := defaultTorUser
	rules := [][]string{
		{"-A", "OUTPUT", "-m", "owner", "--uid-owner", torUser, "-j", "RETURN"},
		{"-A", "OUTPUT", "-p", "udp", "-m", "udp", "--sport", "68", "--dport", "67", "-j", "RETURN"},
		{"-A", "OUTPUT", "-p", "udp", "-m", "udp", "--dport", "53", "-j", "REDIRECT", "--to-ports", routing.DNSPort},
		{"-A", "OUTPUT", "-p", "tcp", "-m", "tcp", "--dport", "53", "-j", "REDIRECT", "--to-ports", routing.DNSPort},
		{"-A", "OUTPUT", "-p", "tcp", "-d", routing.VirtualAddrNetwork, "-j", "REDIRECT", "--to-ports", routing.TransPort},
		{"-A", "OUTPUT", "-p", "udp", "-d", routing.VirtualAddrNetwork, "-j", "REDIRECT", "--to-ports", routing.TransPort},
	}
	for _, network := range excludedNetworks {
		rules = append(rules, []string{"-A", "OUTPUT", "-d", network, "-j", "RETURN"})
	}
	return append(rules,
		[]string{"-A", "OUTPUT", "-p", "tcp", "--syn", "-j", "REDIRECT", "--to-ports", routing.TransPort},
		[]string{"-A", "OUTPUT", "-p", "udp", "-j", "REDIRECT", "--to-ports", routing.TransPort},
	)
}

func (Manager) filterRules(routing tor.RoutingConfig) [][]string {
	torUser := defaultTorUser
	rules := [][]string{
		{"-A", "OUTPUT", "-p", "udp", "-m", "udp", "--sport", "68", "--dport", "67", "-j", "ACCEPT"},
		{"-A", "INPUT", "-p", "udp", "-m", "udp", "--sport", "67", "--dport", "68", "-j", "ACCEPT"},
	}
	for _, network := range excludedNetworks {
		rules = append(rules, []string{"-A", "OUTPUT", "-d", network, "-j", "ACCEPT"})
	}
	return append(rules,
		[]string{"-A", "OUTPUT", "-m", "state", "--state", "ESTABLISHED,RELATED", "-j", "ACCEPT"},
		[]string{"-A", "OUTPUT", "-m", "owner", "--uid-owner", torUser, "-j", "ACCEPT"},
		[]string{"-A", "INPUT", "-m", "state", "--state", "ESTABLISHED", "-j", "ACCEPT"},
		[]string{"-A", "INPUT", "-i", "lo", "-j", "ACCEPT"},
		[]string{"-A", "INPUT", "-j", "DROP"},
		[]string{"-A", "FORWARD", "-j", "DROP"},
		[]string{"-A", "OUTPUT", "-m", "state", "--state", "INVALID", "-j", "DROP"},
		[]string{"-A", "OUTPUT", "-m", "state", "--state", "ESTABLISHED", "-j", "ACCEPT"},
		[]string{"-A", "OUTPUT", "-m", "owner", "--uid-owner", torUser, "-p", "tcp", "-m", "tcp", "--tcp-flags", "FIN,SYN,RST,ACK", "SYN", "-m", "state", "--state", "NEW", "-j", "ACCEPT"},
		[]string{"-A", "OUTPUT", "-d", "127.0.0.1/32", "-o", "lo", "-j", "ACCEPT"},
		[]string{"-A", "OUTPUT", "-d", "127.0.0.1/32", "-p", "tcp", "-m", "tcp", "--dport", routing.TransPort, "--tcp-flags", "FIN,SYN,RST,ACK", "SYN", "-j", "ACCEPT"},
		[]string{"-A", "OUTPUT", "-j", "REJECT"},
	)
}
