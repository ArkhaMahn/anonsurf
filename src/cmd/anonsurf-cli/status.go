package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/parrotsec/packages/anonsurf/internal/anonsurf"
	"gitlab.com/parrotsec/packages/anonsurf/internal/console"
	"gitlab.com/parrotsec/packages/anonsurf/internal/status"
)

func statusCommand(reader status.Reader, monitor anonsurf.MonitorClient) *cobra.Command {
	var useNyx bool
	var useJSON bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show Tor and DNS health",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if useNyx {
				if err := monitor.Run(); err != nil {
					return fmt.Errorf("status --nyx: %w", err)
				}
				return nil
			}
			snapshot, err := reader.Snapshot()
			if err != nil {
				return fmt.Errorf("status: %w", err)
			}
			if useJSON {
				if err := writeJSON(cmd.OutOrStdout(), newStatusOutput(snapshot)); err != nil {
					return fmt.Errorf("write status JSON: %w", err)
				}
				return nil
			}
			printStatus(console.New(cmd.OutOrStdout(), cmd.ErrOrStderr()), snapshot)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&useNyx, "nyx", "n", false, "Open Nyx with the AnonSurf configuration instead of printing a snapshot")
	cmd.Flags().BoolVar(&useJSON, "json", false, "Print the status snapshot as JSON")
	cmd.MarkFlagsMutuallyExclusive("nyx", "json")
	return cmd
}

func monitorCommand(monitor anonsurf.MonitorClient) *cobra.Command {
	return &cobra.Command{
		Use:   "monitor",
		Short: "Open the Nyx Tor monitor",
		Long:  "Open Nyx using the AnonSurf configuration.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := monitor.Run(); err != nil {
				return fmt.Errorf("monitor: %w", err)
			}
			return nil
		},
	}
}

func printStatus(printer console.Printer, snapshot status.Snapshot) {
	printer.Result("AnonSurf status")
	printer.Resultf("  AnonSurf: %s\n", activeLabel(snapshot.AnonSurfActive))
	printer.Resultf("  Tor:      %s\n", activeLabel(snapshot.TorActive))
	printer.Resultf("  DNS:      %s\n", dnsLabel(snapshot.DNSState))
	if len(snapshot.DNSAddresses) > 0 {
		printer.Resultf("  Servers:  %s\n", strings.Join(snapshot.DNSAddresses, ", "))
	}
}

type statusOutput struct {
	AnonSurfActive bool            `json:"anonsurf_active"`
	TorActive      bool            `json:"tor_active"`
	DNSState       status.DNSState `json:"dns_state"`
	DNSAddresses   []string        `json:"dns_addresses"`
}

func newStatusOutput(snapshot status.Snapshot) statusOutput {
	return statusOutput{
		AnonSurfActive: snapshot.AnonSurfActive,
		TorActive:      snapshot.TorActive,
		DNSState:       snapshot.DNSState,
		DNSAddresses:   append([]string{}, snapshot.DNSAddresses...),
	}
}

func activeLabel(active bool) string {
	if active {
		return "active"
	}
	return "inactive"
}

func dnsLabel(state status.DNSState) string {
	switch state {
	case status.DNSTor:
		return "using Tor DNS"
	case status.DNSTorRedirect:
		return "using Tor DNS redirection"
	case status.DNSClearnet:
		return "clearnet DNS"
	case status.DNSLocalhostOnly:
		return "localhost only without AnonSurf"
	case status.DNSEmpty:
		return "no nameservers configured"
	case status.DNSMissing:
		return "resolv.conf missing"
	default:
		return "unknown"
	}
}
