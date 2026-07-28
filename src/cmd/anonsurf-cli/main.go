package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/parrotsec/packages/anonsurf/internal/monitor"
	"gitlab.com/parrotsec/packages/anonsurf/internal/netcheck"
	"gitlab.com/parrotsec/packages/anonsurf/internal/safekill"
	"gitlab.com/parrotsec/packages/anonsurf/internal/status"
	"gitlab.com/parrotsec/packages/anonsurf/internal/systemd"
	"gitlab.com/parrotsec/packages/anonsurf/internal/tor"
)

const version = "6.0.0"

func main() {
	if err := execute(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "anonsurf:", err)
		fmt.Fprintln(os.Stderr, "Run 'anonsurf --help' for usage.")
		os.Exit(1)
	}
}

func execute(args []string, stdout, stderr io.Writer) error {
	return executeWithDependencies(args, stdout, stderr, defaultDependencies())
}

func executeWithDependencies(args []string, stdout, stderr io.Writer, deps dependencies) error {
	rootCmd := newRootCommand(stdout, stderr, deps)
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}

type serviceController interface {
	Start() error
	Stop() error
	Restart() error
	Enable() error
	Disable() error
	IsEnabled() (bool, error)
}

type dependencies struct {
	service  serviceController
	status   statusReader
	identity identityChanger
	ipCheck  ipChecker
	monitor  monitorRunner
	safeKill safeKillRunner
}

type statusReader interface {
	Snapshot() (status.Snapshot, error)
}

type identityChanger interface {
	Change() error
}

type ipChecker interface {
	Check(ctx context.Context) (netcheck.Result, error)
}

type monitorRunner interface {
	Run() error
}

type safeKillRunner interface {
	Run() error
}

func defaultDependencies() dependencies {
	return dependencies{
		service:  systemd.NewAnonsurfdController(),
		status:   status.NewReader(),
		identity: tor.NewIdentityChanger(),
		ipCheck:  netcheck.NewChecker(),
		monitor:  monitor.NewNyxMonitor(os.Stdin, os.Stdout, os.Stderr),
		safeKill: safekill.NewRunner(),
	}
}

func newRootCommand(stdout, stderr io.Writer, deps dependencies) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "anonsurf",
		Short:         "Manage AnonSurf system-wide Tor routing",
		Long:          "AnonSurf manages Parrot OS system-wide Tor routing, DNS safety checks, identity rotation, and startup integration.",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		Example: `  anonsurf status
  anonsurf start
  anonsurf changeid
  anonsurf enable-boot`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetHelpTemplate(helpTemplate)
	rootCmd.SetVersionTemplate("AnonSurf {{.Version}}\n")
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(
		serviceCommandWithSafeKill("start", "Start system-wide Tor transparent routing", deps.service.Start, "AnonSurf started.", deps.safeKill),
		serviceCommandWithSafeKill("stop", "Stop AnonSurf and return to clearnet", deps.service.Stop, "AnonSurf stopped.", deps.safeKill),
		serviceCommandWithSafeKill("restart", "Restart AnonSurf routing", deps.service.Restart, "AnonSurf restarted.", deps.safeKill),
		statusCommand(deps.status, deps.monitor),
		monitorCommand(deps.monitor),
		changeIDCommand(deps.identity),
		myIPCommand(deps.ipCheck),
		serviceCommand("enable-boot", "Enable AnonSurf at system startup", deps.service.Enable, "AnonSurf enabled at startup."),
		serviceCommand("disable-boot", "Disable AnonSurf at system startup", deps.service.Disable, "AnonSurf disabled at startup."),
		statusBootCommand(deps.service),
	)

	return rootCmd
}

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

func changeIDCommand(changer identityChanger) *cobra.Command {
	return &cobra.Command{
		Use:     "changeid",
		Aliases: []string{"newnym"},
		Short:   "Request a new Tor identity",
		Long:    "Request a new Tor identity through Tor's control port.",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := changer.Change(); err != nil {
				return fmt.Errorf("changeid: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Tor identity change requested.")
			return nil
		},
	}
}

func myIPCommand(checker ipChecker) *cobra.Command {
	return &cobra.Command{
		Use:     "myip",
		Aliases: []string{"ip"},
		Short:   "Check the current public Tor status",
		Long:    "Check the current public IP address and whether it is seen as Tor traffic.",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := checker.Check(cmd.Context())
			if err != nil {
				return fmt.Errorf("myip: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Public IP: %s\n", result.IP)
			if result.IsTor {
				fmt.Fprintln(cmd.OutOrStdout(), "Tor:       yes")
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "Tor:       no")
			}
			return nil
		},
	}
}

func serviceCommand(use, short string, run func() error, success string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := run(); err != nil {
				return fmt.Errorf("%s: %w", use, err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), success)
			return nil
		},
	}
}

func serviceCommandWithSafeKill(use, short string, run func() error, success string, safeKill safeKillRunner) *cobra.Command {
	var useSafeKill bool
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if useSafeKill {
				if safeKill == nil {
					return fmt.Errorf("%s safe-kill: runner unavailable", use)
				}
				if err := safeKill.Run(); err != nil {
					return fmt.Errorf("%s safe-kill: %w", use, err)
				}
				fmt.Fprintln(cmd.OutOrStdout(), "SafeKill completed.")
			}
			if err := run(); err != nil {
				return fmt.Errorf("%s: %w", use, err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), success)
			return nil
		},
	}
	cmd.Flags().BoolVar(&useSafeKill, "safe-kill", false, "Run safekill before changing AnonSurf service state")
	return cmd
}

func statusBootCommand(service serviceController) *cobra.Command {
	return &cobra.Command{
		Use:   "status-boot",
		Short: "Show whether AnonSurf starts at boot",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			enabled, err := service.IsEnabled()
			if err != nil {
				return fmt.Errorf("status-boot: %w", err)
			}
			if enabled {
				fmt.Fprintln(cmd.OutOrStdout(), "AnonSurf is enabled at startup.")
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "AnonSurf is disabled at startup.")
			}
			return nil
		},
	}
}

const helpTemplate = `AnonSurf

{{if .Long}}{{.Long}}{{else}}{{.Short}}{{end}}

Usage:
  {{.UseLine}}

Commands:
{{range .Commands}}{{if or .IsAvailableCommand (eq .Name "help")}}  {{rpad .Name .NamePadding }} {{.Short}}
{{end}}{{end}}{{if .HasAvailableFlags}}
Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}
{{end}}{{if .Example}}
Examples:
{{.Example}}
{{end}}`
