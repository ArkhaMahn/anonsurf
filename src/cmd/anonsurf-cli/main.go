package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
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
	service serviceController
}

func defaultDependencies() dependencies {
	return dependencies{
		service: systemd.NewAnonsurfdController(),
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
		serviceCommand("start", "Start system-wide Tor transparent routing", deps.service.Start, "AnonSurf started."),
		serviceCommand("stop", "Stop AnonSurf and return to clearnet", deps.service.Stop, "AnonSurf stopped."),
		serviceCommand("restart", "Restart AnonSurf routing", deps.service.Restart, "AnonSurf restarted."),
		plannedCommand("status", nil, "Show AnonSurf, Tor, and DNS status"),
		plannedCommand("changeid", []string{"newnym"}, "Request a new Tor identity"),
		plannedCommand("myip", []string{"ip"}, "Check the current public Tor status"),
		serviceCommand("enable-boot", "Enable AnonSurf at system startup", deps.service.Enable, "AnonSurf enabled at startup."),
		serviceCommand("disable-boot", "Disable AnonSurf at system startup", deps.service.Disable, "AnonSurf disabled at startup."),
		statusBootCommand(deps.service),
	)

	return rootCmd
}

func serviceCommand(use, short string, run func() error, success string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Long:  short + ".",
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

func statusBootCommand(service serviceController) *cobra.Command {
	return &cobra.Command{
		Use:   "status-boot",
		Short: "Show whether AnonSurf starts at boot",
		Long:  "Show whether AnonSurf starts at boot.",
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

func plannedCommand(use string, aliases []string, short string) *cobra.Command {
	return &cobra.Command{
		Use:     use,
		Aliases: aliases,
		Short:   short,
		Long:    short + ".\n\nThis command is registered in the Go CLI and will be wired to the system controller in a dedicated migration step.",
		Args:    cobra.NoArgs,
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
