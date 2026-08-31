package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/parrotsec/packages/anonsurf/internal/anonsurf"
	"gitlab.com/parrotsec/packages/anonsurf/internal/console"
	"gitlab.com/parrotsec/packages/anonsurf/internal/netcheck"
)

func changeIDCommand(changer anonsurf.IdentityClient) *cobra.Command {
	return &cobra.Command{
		Use:   "changeid",
		Short: "Request a new Tor identity",
		Long:  "Request a new Tor identity through Tor's control port.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := changer.Change(); err != nil {
				return fmt.Errorf("changeid: %w", err)
			}
			console.New(cmd.OutOrStdout(), cmd.ErrOrStderr()).Success("Tor identity change requested.")
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
			printer := console.New(cmd.OutOrStdout(), cmd.ErrOrStderr())
			printer.Resultf("Public IP: %s\n", result.IP)
			if result.IsTor {
				printer.Result("Tor:       yes")
			} else {
				printer.Result("Tor:       no")
			}
			return nil
		},
	}
}
