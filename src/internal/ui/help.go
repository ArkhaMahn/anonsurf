package ui

import (
	"fmt"
)

func printHelpLine(cmd, keyword, args, descr string) {
	printText := "\033[92m" + cmd + "\033[91m " + keyword + "\033[0m "
	if args != "" {
		printText += "\033[93m" + args + "\033[0m "
	}
	if descr != "" {
		printText += "[\033[94m" + descr + "\033[0m]"
	}
	fmt.Println(printText)
}

func printDesc(keyword, descr string) {
	printText := ""
	if keyword != "" {
		printText = "\033[91m" + keyword + "\033[0m: "
	}
	printText += "\033[94m" + descr + "\033[0m"
	fmt.Println("  " + printText)
}

func showBanner() {
	fmt.Println("DNS Tool: A CLI tool to change DNS settings quickly")
	fmt.Println("Gitlab: https://gitlab.com/parrotsec/packages/anonsurf/")
	fmt.Println("License: GPL3")
}

func ShowHelp() {
	programName := "dnstool"
	programNameWithSudo := "sudo dnstool"
	showBanner()
	printHelpLine(programName, "help | -h | --help", "", "Show help banner")
	printHelpLine(programName, "status", "", "Show current system DNS")
	printHelpLine(programNameWithSudo, "address", "<DNS servers>", "Set DNS servers")
	printHelpLine(programNameWithSudo, "create-backup", "", "Make backup for current /etc/resolv.conf")
	printHelpLine(programNameWithSudo, "restore-backup", "", "Restore backup of /etc/resolv.conf")
	fmt.Println("\nAddress could be:")
	printDesc("dhcp", "Address[es] of current DHCP client.")
	printDesc("", "Any IPv4 or IPv6 address[es]")
}
