package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/parrotsec/packages/anonsurf/internal/dns"
	"gitlab.com/parrotsec/packages/anonsurf/internal/sys"
	"gitlab.com/parrotsec/packages/anonsurf/internal/tor"
	"gitlab.com/parrotsec/packages/anonsurf/internal/ui"
)

func dnstShowStatus(manager dns.Manager) error {
	addresses, err := dns.ParseDNSAddresses(manager.ResolvConfPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			ui.PrintErrorResolvNotFound(manager.ResolvConfPath)
			return nil
		}
		return err
	}

	if len(addresses) == 0 {
		ui.PrintErrorResolvEmpty(manager.ResolvConfPath)
	} else if tor.IsAnonsurfRunning() {
		if dns.HasOnlyLocalhost(addresses) {
			ui.PrintUnderTorDNS()
		} else {
			ui.PrintErrorDNSLeak()
		}
	} else {
		if dns.HasOnlyLocalhost(addresses) {
			ui.PrintErrorLocalHost()
		} else {
			ui.PrintDNSAddresses(addresses)
		}
	}
	return nil
}

func dnstRestoreBackup(manager dns.Manager) error {
	if err := manager.RestoreBackupOrDHCP(); err != nil {
		return err
	}
	return dnstShowStatus(manager)
}

func dnstWriteAddresses(manager dns.Manager, hasDHCP bool, listAddr []string) error {
	if hasDHCP {
		if len(listAddr) == 0 {
			if err := manager.CreateDHCPOnly(); err != nil {
				return err
			}
		} else {
			addrFromDHCP, err := manager.DHCPAddresses()
			if err != nil {
				return fmt.Errorf("read DHCP DNS addresses: %w", err)
			}
			if err := manager.CreateAddrs(append(listAddr, addrFromDHCP...)); err != nil {
				return err
			}
		}
	} else {
		if err := manager.CreateAddrs(listAddr); err != nil {
			return err
		}
	}

	fmt.Println("[*] Applied DNS settings")
	return dnstShowStatus(manager)
}

func requireAdmin() error {
	if !sys.IsAdmin() {
		return errors.New("require admin's privilege")
	}
	return nil
}

func main() {
	manager := dns.NewManager()

	var rootCmd = &cobra.Command{
		Use:           "dnstool",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.ShowHelp()
			return dnstShowStatus(manager)
		},
	}

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show current DNS status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return dnstShowStatus(manager)
		},
	}

	createBackupCmd := &cobra.Command{
		Use:   "create-backup",
		Short: "Create a backup of resolv.conf",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAdmin(); err != nil {
				return err
			}
			return manager.MakeBackup()
		},
	}

	restoreBackupCmd := &cobra.Command{
		Use:   "restore-backup",
		Short: "Restore resolv.conf from backup",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAdmin(); err != nil {
				return err
			}
			return dnstRestoreBackup(manager)
		},
	}

	addressCmd := &cobra.Command{
		Use:     "address [dhcp] [addr...]",
		Aliases: []string{"addr"},
		Short:   "Write custom DNS addresses or use DHCP",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAdmin(); err != nil {
				return err
			}
			addrResult := dns.ParseAddrFromParams(args)
			if len(addrResult.Invalid) > 0 {
				return fmt.Errorf("invalid DNS address: %s", strings.Join(addrResult.Invalid, ", "))
			}
			return dnstWriteAddresses(manager, addrResult.HasDHCPFlag, addrResult.ListAddr)
		},
	}

	rootCmd.AddCommand(statusCmd, createBackupCmd, restoreBackupCmd, addressCmd)

	if err := rootCmd.Execute(); err != nil {
		ui.PrintError(err.Error())
		os.Exit(1)
	}
}
