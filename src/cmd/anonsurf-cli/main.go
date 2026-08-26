package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/parrotsec/packages/anonsurf/internal/anondaemon"
	"gitlab.com/parrotsec/packages/anonsurf/internal/netcheck"
	"gitlab.com/parrotsec/packages/anonsurf/internal/safekill"
	"gitlab.com/parrotsec/packages/anonsurf/internal/status"
	"gitlab.com/parrotsec/packages/anonsurf/internal/systemd"
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
	rootCmd := newRootCommand(stdout, stderr)
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}

func newRootCommand(stdout, stderr io.Writer) *cobra.Command {
	cobra.EnableCommandSorting = false
	rootCmd := &cobra.Command{
		Use:           "anonsurf",
		Short:         "Manage AnonSurf system-wide Tor routing",
		Long:          "AnonSurf is a system-wide Tor routing tool for Parrot OS.",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		Example: `  anonsurf start
  anonsurf status
  anonsurf changeid
  anonsurf boot enable`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetHelpTemplate(helpTemplate)
	rootCmd.SetVersionTemplate("AnonSurf {{.Version}}\n")
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	service := systemd.NewAnonsurfdController()
	monitor := anondaemon.MonitorClient{}
	safeKill := safekill.Runner{}
	rootCmd.AddCommand(
		serviceCommandWithSafeKill("start", "Route traffic through Tor", service.Start, "AnonSurf started.", safeKill),
		serviceCommandWithSafeKill("stop", "Restore the original network", service.Stop, "AnonSurf stopped.", safeKill),
		serviceCommandWithSafeKill("restart", "Rebuild protected routing", service.Restart, "AnonSurf restarted.", safeKill),
		statusCommand(status.Reader{}, monitor),
		changeIDCommand(anondaemon.IdentityClient{}),
		ipCommand(netcheck.Checker{}),
		monitorCommand(monitor),
		bootCommand(service),
	)
	configureRootHelp(rootCmd)

	return rootCmd
}

const helpTemplate = `AnonSurf

{{if .Long}}{{.Long}}{{else}}{{.Short}}{{end}}

Usage:
  {{.UseLine}}
{{if .HasAvailableSubCommands}}
Commands:
{{range .Commands}}{{if or .IsAvailableCommand (eq .Name "help")}}  {{rpad .Name .NamePadding }} {{.Short}}
{{end}}{{end}}{{end}}
{{if .HasAvailableFlags}}Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}
{{end}}{{if .Example}}
Examples:
{{.Example}}
{{end}}`
