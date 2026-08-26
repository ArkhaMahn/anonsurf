package main

import (
	"fmt"
)

const (
	red   = "\033[91m"
	green = "\033[32m"
	blue  = "\033[94m"
	reset = "\033[0m"
)

func printError(msg string) {
	fmt.Println("[" + red + "!" + reset + "] " + red + msg + reset)
}

func printUnderTorDNS() {
	fmt.Println(" " + green + "Using Tor's DNS" + reset)
}

func printUnderTorDNSRedirect() {
	fmt.Println(" " + green + "Using Tor DNS redirection with transport DNS preserved" + reset)
}

func printErrorResolvNotFound(resolvConfPath string) {
	fmt.Println("[" + red + "DNS error" + reset + "] " + blue + resolvConfPath + reset + " not found")
}

func printErrorResolvEmpty(resolvConfPath string) {
	fmt.Println("[" + red + "DNS error" + reset + "] " + blue + resolvConfPath + reset + " is empty")
}

func printErrorLocalHost() {
	fmt.Println(" " + red + "LocalHost only. This may cause no internet access" + reset)
}

func printInfoAddresses() {
	fmt.Println("- " + red + "Address[es]" + reset + ":")
}

func printDNSAddresses(listAddr []string) {
	printInfoAddresses()
	for _, address := range listAddr {
		fmt.Println("  " + address)
	}
}
