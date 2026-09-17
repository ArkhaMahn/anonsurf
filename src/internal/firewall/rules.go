package firewall

import (
	"fmt"
	"strings"

	"gitlab.com/parrotsec/packages/anonsurf/internal/tor"
)

type iptablesChain struct {
	name   string
	policy string
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

func (m Manager) renderRules(routing tor.RoutingConfig, backup []byte) string {
	var builder strings.Builder
	writeIPTablesTable(&builder, "filter", iptablesChainDefinitions(backup, "filter"), m.filterRules(routing))
	writeIPTablesTable(&builder, "nat", iptablesChainDefinitions(backup, "nat"), m.natRules(routing))
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
