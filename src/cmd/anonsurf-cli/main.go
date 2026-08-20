package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
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
