package main

import (
	"fmt"
	"io"
	"os"
)

type palette struct {
	red, green, blue, reset string
}

func newPalette(output io.Writer) palette {
	if !supportsColor(output) {
		return palette{}
	}
	return palette{
		red:   "\033[91m",
		green: "\033[32m",
		blue:  "\033[94m",
		reset: "\033[0m",
	}
}

func supportsColor(output io.Writer) bool {
	if _, disabled := os.LookupEnv("NO_COLOR"); disabled || os.Getenv("TERM") == "dumb" {
		return false
	}
	file, ok := output.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

var (
	out    = newPalette(os.Stdout)
	errOut = newPalette(os.Stderr)
)

func printError(msg string) {
	fmt.Fprintln(os.Stderr, "["+errOut.red+"!"+errOut.reset+"] "+errOut.red+msg+errOut.reset)
}

func printUnderTorDNS() {
	fmt.Println(" " + out.green + "Using Tor's DNS" + out.reset)
}

func printUnderTorDNSRedirect() {
	fmt.Println(" " + out.green + "Using Tor DNS redirection with transport DNS preserved" + out.reset)
}

func printErrorResolvNotFound(resolvConfPath string) {
	fmt.Println("[" + out.red + "DNS error" + out.reset + "] " + out.blue + resolvConfPath + out.reset + " not found")
}

func printErrorResolvEmpty(resolvConfPath string) {
	fmt.Println("[" + out.red + "DNS error" + out.reset + "] " + out.blue + resolvConfPath + out.reset + " is empty")
}

func printErrorLocalHost() {
	fmt.Println(" " + out.red + "LocalHost only. This may cause no internet access" + out.reset)
}

func printInfoAddresses() {
	fmt.Println("- " + out.red + "Address[es]" + out.reset + ":")
}

func printDNSAddresses(listAddr []string) {
	printInfoAddresses()
	for _, address := range listAddr {
		fmt.Println("  " + address)
	}
}
