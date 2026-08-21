package anondaemon

import (
	"fmt"

	"gitlab.com/parrotsec/packages/anonsurf/internal/systemd"
	"gitlab.com/parrotsec/packages/anonsurf/internal/tor"
)

func (c Controller) saveIPTablesRules() error {
	path := IPTablesRulesPath
	if exists, err := fileExists(path); err != nil {
		return fmt.Errorf("check iptables backup: %w", err)
	} else if exists {
		return nil
	}

	rules, err := (systemd.ExecRunner{}).Output(IPTablesSavePath)
	if err != nil {
		return fmt.Errorf("save iptables rules: %w", err)
	}
	if err := writeFile(path, rules, 0600); err != nil {
		return fmt.Errorf("write iptables backup: %w", err)
	}
	return nil
}

func (c Controller) applyIPTablesRules(routing tor.RoutingConfig) error {
	backup, err := readFile(IPTablesRulesPath)
	if err != nil {
		return fmt.Errorf("read iptables backup for apply: %w", err)
	}
	rules := c.renderIPTablesRules(routing, backup)
	if err := runWithInput(IPTablesRestorePath, []byte(rules)); err != nil {
		return fmt.Errorf("apply iptables rules: %w", err)
	}
	return nil
}

func (c Controller) restoreFirewall() error {
	exists, err := fileExists(IPTablesRulesPath)
	if err != nil {
		return fmt.Errorf("check iptables backup: %w", err)
	}
	if !exists {
		return nil
	}
	return c.restoreIPTablesRules()
}

func (c Controller) restoreIPTablesRules() error {
	path := IPTablesRulesPath
	rules, err := readFile(path)
	if err != nil {
		return fmt.Errorf("read iptables backup: %w", err)
	}
	rules = ensureManagedIPTablesTables(rules)
	if err := runWithInput(IPTablesRestorePath, rules); err != nil {
		return fmt.Errorf("restore iptables rules: %w", err)
	}
	if err := removeFile(path); err != nil {
		return fmt.Errorf("remove iptables backup: %w", err)
	}
	return nil
}

func (c Controller) natIPTablesRules(routing tor.RoutingConfig) [][]string {
	torUser := defaultTorUser
	rules := [][]string{
		{"-A", "OUTPUT", "-m", "owner", "--uid-owner", torUser, "-j", "RETURN"},
		{"-A", "OUTPUT", "-p", "udp", "-m", "udp", "--sport", "68", "--dport", "67", "-j", "RETURN"},
		{"-A", "OUTPUT", "-p", "udp", "-m", "udp", "--dport", "53", "-j", "REDIRECT", "--to-ports", routing.DNSPort},
		{"-A", "OUTPUT", "-p", "tcp", "-m", "tcp", "--dport", "53", "-j", "REDIRECT", "--to-ports", routing.DNSPort},
		{"-A", "OUTPUT", "-p", "tcp", "-d", routing.VirtualAddrNetwork, "-j", "REDIRECT", "--to-ports", routing.TransPort},
		{"-A", "OUTPUT", "-p", "udp", "-d", routing.VirtualAddrNetwork, "-j", "REDIRECT", "--to-ports", routing.TransPort},
	}
	for _, network := range torExcludedNetworks {
		rules = append(rules, []string{"-A", "OUTPUT", "-d", network, "-j", "RETURN"})
	}
	return append(rules,
		[]string{"-A", "OUTPUT", "-p", "tcp", "--syn", "-j", "REDIRECT", "--to-ports", routing.TransPort},
		[]string{"-A", "OUTPUT", "-p", "udp", "-j", "REDIRECT", "--to-ports", routing.TransPort},
		[]string{"-A", "OUTPUT", "-p", "icmp", "-j", "REDIRECT", "--to-ports", routing.TransPort},
	)
}

func (c Controller) filterIPTablesRules(routing tor.RoutingConfig) [][]string {
	torUser := defaultTorUser
	rules := [][]string{
		{"-A", "OUTPUT", "-p", "udp", "-m", "udp", "--sport", "68", "--dport", "67", "-j", "ACCEPT"},
		{"-A", "INPUT", "-p", "udp", "-m", "udp", "--sport", "67", "--dport", "68", "-j", "ACCEPT"},
	}
	for _, network := range torExcludedNetworks {
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
