package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

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
