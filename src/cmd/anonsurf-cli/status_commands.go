package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/parrotsec/packages/anonsurf/internal/status"
)

func statusCommand(reader statusReader, monitor monitorRunner) *cobra.Command {
	var useNyx bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show AnonSurf, Tor, and DNS status",
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
			printStatus(cmd.OutOrStdout(), snapshot)
			return nil
		},
	}
	cmd.Flags().BoolVar(&useNyx, "nyx", false, "Open Nyx with the AnonSurf configuration instead of printing a snapshot")
	return cmd
}

func monitorCommand(monitor monitorRunner) *cobra.Command {
	return &cobra.Command{
		Use:   "monitor",
		Short: "Open Nyx Tor monitor",
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

func printStatus(output io.Writer, snapshot status.Snapshot) {
	fmt.Fprintln(output, "AnonSurf status")
	fmt.Fprintf(output, "  AnonSurf: %s\n", activeLabel(snapshot.AnonSurfActive))
	fmt.Fprintf(output, "  Tor:      %s\n", activeLabel(snapshot.TorActive))
	fmt.Fprintf(output, "  DNS:      %s\n", dnsLabel(snapshot.DNSState))
	if len(snapshot.DNSAddresses) > 0 {
		fmt.Fprintf(output, "  Servers:  %s\n", strings.Join(snapshot.DNSAddresses, ", "))
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
	case status.DNSLeak:
		return "possible leak: non-Tor DNS while AnonSurf is active"
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
