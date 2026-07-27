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
	rootCmd := newRootCommand(stdout, stderr)
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}

func newRootCommand(stdout, stderr io.Writer) *cobra.Command {
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
		plannedCommand("start", nil, "Start system-wide Tor transparent routing"),
		plannedCommand("stop", nil, "Stop AnonSurf and return to clearnet"),
		plannedCommand("restart", nil, "Restart AnonSurf routing"),
		plannedCommand("status", nil, "Show AnonSurf, Tor, and DNS status"),
		plannedCommand("changeid", []string{"newnym"}, "Request a new Tor identity"),
		plannedCommand("myip", []string{"ip"}, "Check the current public Tor status"),
		plannedCommand("enable-boot", nil, "Enable AnonSurf at system startup"),
		plannedCommand("disable-boot", nil, "Disable AnonSurf at system startup"),
		plannedCommand("status-boot", nil, "Show whether AnonSurf starts at boot"),
	)

	return rootCmd
}

func plannedCommand(use string, aliases []string, short string) *cobra.Command {
	return &cobra.Command{
		Use:     use,
		Aliases: aliases,
		Short:   short,
		Long:    short + ".\n\nThis command is registered in the Go CLI and will be wired to the system controller in a dedicated migration step.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("%s is not implemented yet in the Go CLI", cmd.CommandPath())
		},
	}
}

const helpTemplate = `AnonSurf

{{.Long}}

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
