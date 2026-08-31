package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/parrotsec/packages/anonsurf/internal/console"
	"gitlab.com/parrotsec/packages/anonsurf/internal/safekill"
	"gitlab.com/parrotsec/packages/anonsurf/internal/systemd"
)

func serviceCommand(use, short, activity string, run func() error, success string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			printer := console.New(cmd.OutOrStdout(), cmd.ErrOrStderr())
			printer.Running(activity)
			if err := run(); err != nil {
				return fmt.Errorf("%s: %w", use, err)
			}
			printer.Success(success)
			return nil
		},
	}
}

func serviceCommandWithSafeKill(
	use, short, activity string,
	run func() error,
	success string,
	safeKill safekill.Runner,
) *cobra.Command {
	var useSafeKill bool
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			printer := console.New(cmd.OutOrStdout(), cmd.ErrOrStderr())
			if useSafeKill {
				printer.Running("Closing networked applications and clearing caches...")
				if err := safeKill.Run(); err != nil {
					return fmt.Errorf("%s safe-kill: %w", use, err)
				}
				printer.Success("SafeKill completed.")
			}
			printer.Running(activity)
			if err := run(); err != nil {
				return fmt.Errorf("%s: %w", use, err)
			}
			printer.Success(success)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&useSafeKill, "safe-kill", "k", false, "Run safekill before changing AnonSurf service state")
	return cmd
}

func bootCommand(service systemd.ServiceController) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "boot",
		Short: "Manage startup integration",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(
		serviceCommand(
			"enable", "Enable AnonSurf at system startup", "Enabling AnonSurf at startup...",
			service.Enable, "AnonSurf enabled at startup.",
		),
		serviceCommand(
			"disable", "Disable AnonSurf at system startup", "Disabling AnonSurf at startup...",
			service.Disable, "AnonSurf disabled at startup.",
		),
		statusBootCommand(service),
	)
	return cmd
}

func statusBootCommand(service systemd.ServiceController) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show whether AnonSurf starts at boot",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			enabled, err := service.IsEnabled()
			if err != nil {
				return fmt.Errorf("boot status: %w", err)
			}
			printer := console.New(cmd.OutOrStdout(), cmd.ErrOrStderr())
			if enabled {
				printer.Result("AnonSurf is enabled at startup.")
			} else {
				printer.Result("AnonSurf is disabled at startup.")
			}
			return nil
		},
	}
}
