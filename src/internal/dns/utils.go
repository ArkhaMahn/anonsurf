package dns

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"gitlab.com/parrotsec/packages/anonsurf/internal/ui"
)

type AddrFromParams struct {
	HasDHCPFlag bool
	ListAddr    []string
}

func ConvertSeqAddrToString(addresses []string) string {
	var dnsAddrInText string
	for _, address := range addresses {
		dnsAddrInText += "nameserver " + address + "\n"
	}
	return dnsAddrInText
}

func HasOnlyLocalhost(listAddr []string) bool {
	if len(listAddr) == 1 && (listAddr[0] == "localhost" || listAddr[0] == "127.0.0.1") {
		return true
	}
	return false
}

func SystemDNSFileIsSymlink(resolvConfPath string) (bool, error) {
	fileInfo, err := os.Lstat(resolvConfPath)
	if err != nil {
		return false, err
	}
	return fileInfo.Mode()&os.ModeSymlink != 0, nil
}

func ParseDNSAddresses(resolvConfPath string) []string {
	var result []string
	file, err := os.Open(resolvConfPath)
	if err != nil {
		return result
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("Error closing file:", err)
		}
	}(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "nameserver" {
			result = append(result, fields[1])
		}
	}
	return result
}

func SystemHasOnlyLocalhost(resolvConfPath string) bool {
	addresses := ParseDNSAddresses(resolvConfPath)
	return HasOnlyLocalhost(addresses)
}

func WriteDNSToSystem(listDNSAddr []string, resolvConfPath string) {
	seen := make(map[string]bool)
	var deduplicatedListAddr []string
	for _, addr := range listDNSAddr {
		if !seen[addr] {
			seen[addr] = true
			deduplicatedListAddr = append(deduplicatedListAddr, addr)
		}
	}

	if len(deduplicatedListAddr) == 0 {
		ui.PrintError("No valid DNS address is found.")
		return
	}

	dnsAddr := ConvertSeqAddrToString(deduplicatedListAddr)
	content := "# Written by DNSTool\n" + dnsAddr

	err := os.WriteFile(resolvConfPath, []byte(content), 0644)
	if err != nil {
		ui.PrintError("Error while writing DNS address to " + resolvConfPath)
	}
}

func GetListValidAddresses(listAddr []string) []string {
	var result []string
	for _, eachAddr := range listAddr {
		if net.ParseIP(eachAddr) != nil {
			result = append(result, eachAddr)
		}
	}
	return result
}

func ParseAddrFromParams(params []string) AddrFromParams {
	paramResult := AddrFromParams{
		HasDHCPFlag: false,
		ListAddr:    []string{},
	}

	for _, value := range params {
		if value == "dhcp" {
			paramResult.HasDHCPFlag = true
		} else if net.ParseIP(value) != nil {
			paramResult.ListAddr = append(paramResult.ListAddr, value)
		}
	}

	return paramResult
}
