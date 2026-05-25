package ui

import (
	"fmt"
)

const (
	Red   = "\033[91m"
	Green = "\033[32m"
	Blue  = "\033[94m"
	Reset = "\033[0m"
)

func PrintError(msg string) {
	fmt.Println("[" + Red + "!" + Reset + "] " + Red + msg + Reset)
}

func PrintUnderTorDNS() {
	fmt.Println(" " + Green + "Using Tor's DNS" + Reset)
}

func PrintErrorResolvNotFound(resolvConfPath string) {
	fmt.Println("[" + Red + "DNS error" + Reset + "] " + Blue + resolvConfPath + Reset + " not found")
}

func PrintErrorResolvEmpty(resolvConfPath string) {
	fmt.Println("[" + Red + "DNS error" + Reset + "] " + Blue + resolvConfPath + Reset + " is empty")
}

func PrintErrorLocalHost() {
	fmt.Println(" " + Red + "LocalHost only. This may cause no internet access" + Reset)
}

func PrintErrorDNSLeak() {
	fmt.Println(Red + "\nDetected Non-Tor address[es]. This may cause information leaks." + Reset)
}

func PrintAnonsurfRunning() {
	fmt.Println("- " + Red + "Address" + Reset + ": AnonSurf is running")
}

func PrintInfoAddresses() {
	fmt.Println("- " + Red + "Address[es]" + Reset + ":")
}

func PrintDNSAddresses(listAddr []string) {
	PrintInfoAddresses()
	for _, address := range listAddr {
		fmt.Println("  " + address)
	}
}
