package anondaemon

import (
	"fmt"
	"strings"

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

func ensureManagedIPTablesTables(rules []byte) []byte {
	content := string(rules)
	for _, table := range []string{"filter", "nat"} {
		if hasIPTablesTable(content, table) {
			continue
		}
		if content != "" && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += emptyIPTablesTable(table)
	}
	return []byte(content)
}

func hasIPTablesTable(rules, table string) bool {
	for _, line := range strings.Split(rules, "\n") {
		if strings.TrimSpace(line) == "*"+table {
			return true
		}
	}
	return false
}

type iptablesChain struct {
	name   string
	policy string
}

func defaultIPTablesChains(table string) []iptablesChain {
	switch table {
	case "filter":
		return []iptablesChain{
			{name: "INPUT", policy: "ACCEPT"},
			{name: "FORWARD", policy: "ACCEPT"},
			{name: "OUTPUT", policy: "ACCEPT"},
		}
	case "nat":
		return []iptablesChain{
			{name: "PREROUTING", policy: "ACCEPT"},
			{name: "INPUT", policy: "ACCEPT"},
			{name: "OUTPUT", policy: "ACCEPT"},
			{name: "POSTROUTING", policy: "ACCEPT"},
		}
	default:
		return nil
	}
}

func emptyIPTablesTable(table string) string {
	var builder strings.Builder
	builder.WriteString("*")
	builder.WriteString(table)
	builder.WriteString("\n")
	for _, chain := range defaultIPTablesChains(table) {
		fmt.Fprintf(&builder, ":%s %s [0:0]\n", chain.name, chain.policy)
	}
	builder.WriteString("COMMIT\n")
	return builder.String()
}

func iptablesChainDefinitions(backup []byte, table string) []string {
	var definitions []string
	seen := make(map[string]bool)
	inTable := false

	for _, line := range strings.Split(string(backup), "\n") {
		line = strings.TrimSpace(line)
		switch {
		case line == "*"+table:
			inTable = true
			continue
		case strings.HasPrefix(line, "*") || line == "COMMIT":
			inTable = false
		}
		if !inTable || !strings.HasPrefix(line, ":") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		name := strings.TrimPrefix(fields[0], ":")
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		definitions = append(definitions, fmt.Sprintf(":%s %s [0:0]", name, fields[1]))
	}

	for _, chain := range defaultIPTablesChains(table) {
		if !seen[chain.name] {
			definitions = append(definitions, fmt.Sprintf(":%s %s [0:0]", chain.name, chain.policy))
		}
	}
	return definitions
}

func (c Controller) renderIPTablesRules(routing tor.RoutingConfig, backup []byte) string {
	var builder strings.Builder
	writeIPTablesTable(&builder, "filter", iptablesChainDefinitions(backup, "filter"), c.filterIPTablesRules(routing))
	writeIPTablesTable(&builder, "nat", iptablesChainDefinitions(backup, "nat"), c.natIPTablesRules(routing))
	return builder.String()
}

func writeIPTablesTable(builder *strings.Builder, table string, definitions []string, rules [][]string) {
	builder.WriteString("*")
	builder.WriteString(table)
	builder.WriteString("\n")
	for _, definition := range definitions {
		builder.WriteString(definition)
		builder.WriteString("\n")
	}
	for _, rule := range rules {
		builder.WriteString(strings.Join(rule, " "))
		builder.WriteString("\n")
	}
	builder.WriteString("COMMIT\n")
}

func (c Controller) natIPTablesRules(routing tor.RoutingConfig) [][]string {
	torUser := defaultTorUser
	rules := [][]string{
		{"-A", "OUTPUT", "-m", "owner", "--uid-owner", torUser, "-j", "RETURN"},
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
	var rules [][]string
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
