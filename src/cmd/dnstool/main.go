package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/parrotsec/packages/anonsurf/internal/dns"
	"gitlab.com/parrotsec/packages/anonsurf/internal/sys"
	"gitlab.com/parrotsec/packages/anonsurf/internal/tor"
	"gitlab.com/parrotsec/packages/anonsurf/internal/ui"
)

func dnstShowStatus() {
	if _, err := os.Stat(dns.ResolvConfPath); os.IsNotExist(err) {
		ui.PrintErrorResolvNotFound(dns.ResolvConfPath)
		return
	}

	addresses := dns.ParseDNSAddresses(dns.ResolvConfPath)

	if len(addresses) == 0 {
		ui.PrintErrorResolvEmpty(dns.ResolvConfPath)
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
}

func dnstRestoreBackup() {
	if _, err := os.Stat(dns.ResolvConfBackup); !os.IsNotExist(err) {
		dns.ResolvRestoreFromBackup()
	} else {
		dns.NMCreateHookDHCP()
	}
	dnstShowStatus()
}

func dnstCreateBackup() {
	dns.ResolvMakeBackup()
}

func dnstWriteAddresses(hasDHCP bool, listAddr []string) {
	if hasDHCP {
		if len(listAddr) == 0 {
			dns.ResolvCreateDHCPOnly()
		} else {
			addrFromDHCP := dns.GetListValidAddresses(dns.NMGetDNSServers())
			dns.ResolvCreateAddrs(append(listAddr, addrFromDHCP...))
		}
	} else {
		dns.ResolvCreateAddrs(listAddr)
	}

	fmt.Println("[*] Applied DNS settings")
	dnstShowStatus()
}

func main() {
	var rootCmd = &cobra.Command{
		Use: "dnstool",
		Run: func(cmd *cobra.Command, args []string) {
			ui.ShowHelp()
			dnstShowStatus()
		},
	}

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show current DNS status",
		Run: func(cmd *cobra.Command, args []string) {
			dnstShowStatus()
		},
	}

	createBackupCmd := &cobra.Command{
		Use:   "create-backup",
		Short: "Create a backup of resolv.conf",
		Run: func(cmd *cobra.Command, args []string) {
			if !sys.IsAdmin() {
				ui.PrintError("Require admin's privilege")
				return
			}
			dnstCreateBackup()
		},
	}

	restoreBackupCmd := &cobra.Command{
		Use:   "restore-backup",
		Short: "Restore resolv.conf from backup",
		Run: func(cmd *cobra.Command, args []string) {
			if !sys.IsAdmin() {
				ui.PrintError("Require admin's privilege")
				return
			}
			dnstRestoreBackup()
		},
	}

	addressCmd := &cobra.Command{
		Use:     "address [dhcp] [addr...]",
		Aliases: []string{"addr"},
		Short:   "Write custom DNS addresses or use DHCP",
		Run: func(cmd *cobra.Command, args []string) {
			if !sys.IsAdmin() {
				ui.PrintError("Require admin's privilege")
				return
			}
			addrResult := dns.ParseAddrFromParams(args)
			dnstWriteAddresses(addrResult.HasDHCPFlag, addrResult.ListAddr)
		},
	}

	rootCmd.AddCommand(statusCmd, createBackupCmd, restoreBackupCmd, addressCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
