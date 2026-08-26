package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/parrotsec/packages/anonsurf/internal/anondaemon"
	"gitlab.com/parrotsec/packages/anonsurf/internal/netcheck"
)

func changeIDCommand(changer anondaemon.IdentityClient) *cobra.Command {
	return &cobra.Command{
		Use:   "changeid",
		Short: "Request a new Tor identity",
		Long:  "Request a new Tor identity through Tor's control port.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := changer.Change(); err != nil {
				return fmt.Errorf("changeid: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Tor identity change requested.")
			return nil
		},
	}
}

func ipCommand(checker netcheck.Checker) *cobra.Command {
	return &cobra.Command{
		Use:   "ip",
		Short: "Verify the public IP and Tor",
		Long:  "Check the current public IP address and whether it is seen as Tor traffic.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := checker.Check(cmd.Context())
			if err != nil {
				return fmt.Errorf("ip: %w", err)
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
