package anondaemon

import (
	"fmt"
	"strings"

	"gitlab.com/parrotsec/packages/anonsurf/internal/tor"
)

func (c Controller) saveIPTablesRules() error {
	path := c.iptablesRulesPath()
	if exists, err := c.fileExists(path); err != nil {
		return fmt.Errorf("check iptables backup: %w", err)
	} else if exists {
		return nil
	}

	rules, err := c.runner().Output(c.iptablesSavePath())
	if err != nil {
		return fmt.Errorf("save iptables rules: %w", err)
	}
	if err := c.writeFile(path, rules, 0644); err != nil {
		return fmt.Errorf("write iptables backup: %w", err)
	}
	return nil
}

func (c Controller) applyIPTablesRules(routing tor.RoutingConfig) error {
	for _, args := range c.iptablesRules(routing) {
		if err := c.run(c.iptablesPath(), args...); err != nil {
			return fmt.Errorf("apply iptables rule %s: %w", strings.Join(args, " "), err)
		}
	}
	return nil
}

func (c Controller) iptablesRules(routing tor.RoutingConfig) [][]string {
	torUser := c.torUser()
	commands := [][]string{
		{"-F"},
		{"-t", "nat", "-F"},
		{"-t", "nat", "-A", "OUTPUT", "-m", "owner", "--uid-owner", torUser, "-j", "RETURN"},
		{"-t", "nat", "-A", "OUTPUT", "-d", "127.0.0.1/32", "-p", "udp", "-m", "udp", "--dport", "53", "-j", "REDIRECT", "--to-ports", routing.DNSPort},
		{"-t", "nat", "-A", "OUTPUT", "-p", "tcp", "-d", routing.VirtualAddrNetwork, "-j", "REDIRECT", "--to-ports", routing.TransPort},
		{"-t", "nat", "-A", "OUTPUT", "-p", "udp", "-d", routing.VirtualAddrNetwork, "-j", "REDIRECT", "--to-ports", routing.TransPort},
	}
	for _, network := range c.torExclude() {
		commands = append(commands,
			[]string{"-t", "nat", "-A", "OUTPUT", "-d", network, "-j", "RETURN"},
			[]string{"-A", "OUTPUT", "-d", network, "-j", "ACCEPT"},
		)
	}
	return append(commands,
		[]string{"-t", "nat", "-A", "OUTPUT", "-p", "tcp", "--syn", "-j", "REDIRECT", "--to-ports", routing.TransPort},
		[]string{"-t", "nat", "-A", "OUTPUT", "-p", "udp", "-j", "REDIRECT", "--to-ports", routing.TransPort},
		[]string{"-t", "nat", "-A", "OUTPUT", "-p", "icmp", "-j", "REDIRECT", "--to-ports", routing.TransPort},
		[]string{"-A", "OUTPUT", "-m", "state", "--state", "ESTABLISHED,RELATED", "-j", "ACCEPT"},
		[]string{"-A", "OUTPUT", "-m", "owner", "--uid-owner", torUser, "-j", "ACCEPT"},
		[]string{"-A", "OUTPUT", "-j", "REJECT"},
		[]string{"-A", "INPUT", "-m", "state", "--state", "ESTABLISHED", "-j", "ACCEPT"},
		[]string{"-A", "INPUT", "-i", "lo", "-j", "ACCEPT"},
		[]string{"-A", "INPUT", "-j", "DROP"},
		[]string{"-A", "FORWARD", "-j", "DROP"},
		[]string{"-A", "OUTPUT", "-m", "state", "--state", "INVALID", "-j", "DROP"},
		[]string{"-A", "OUTPUT", "-m", "state", "--state", "ESTABLISHED", "-j", "ACCEPT"},
		[]string{"-A", "OUTPUT", "-m", "owner", "--uid-owner", torUser, "-p", "tcp", "-m", "tcp", "--tcp-flags", "FIN,SYN,RST,ACK", "SYN", "-m", "state", "--state", "NEW", "-j", "ACCEPT"},
		[]string{"-A", "OUTPUT", "-d", "127.0.0.1/32", "-o", "lo", "-j", "ACCEPT"},
		[]string{"-A", "OUTPUT", "-d", "127.0.0.1/32", "-p", "tcp", "-m", "tcp", "--dport", routing.TransPort, "--tcp-flags", "FIN,SYN,RST,ACK", "SYN", "-j", "ACCEPT"},
	)
}
