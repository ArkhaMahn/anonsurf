package tor

import (
	"fmt"
	"strconv"
	"strings"
)

func renderBridgeSettings(options TorrcOptions) (string, error) {
	address := strings.TrimSpace(options.BridgeAddress)
	if options.AutoBridge && address == "" {
		return "", fmt.Errorf("auto bridge requires a selected bridge address")
	}
	if address == "" {
		return "# Bridge setting: No bridge", nil
	}

	transport, err := bridgeTransport(address)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	builder.WriteString("# Enable bridge mode\nUseBridges 1\n")
	switch transport {
	case "obfs4":
		if strings.TrimSpace(options.Obfs4ProxyPath) == "" {
			return "", fmt.Errorf("obfs4 bridge requires obfs4proxy path")
		}
		builder.WriteString("ClientTransportPlugin obfs4 exec ")
		builder.WriteString(strings.TrimSpace(options.Obfs4ProxyPath))
		builder.WriteString(" managed\n")
	case "snowflake":
		if strings.TrimSpace(options.SnowflakeClientPath) == "" {
			return "", fmt.Errorf("snowflake bridge requires snowflake-client path")
		}
		builder.WriteString("ClientTransportPlugin snowflake exec ")
		builder.WriteString(strings.TrimSpace(options.SnowflakeClientPath))
		builder.WriteString("\n")
	case "meek":
		return "", fmt.Errorf("meek bridges are not supported")
	}

	builder.WriteString("\nBridge ")
	builder.WriteString(address)
	return builder.String(), nil
}

func bridgeTransport(address string) (string, error) {
	fields := strings.Fields(address)
	if len(fields) == 0 {
		return "", fmt.Errorf("bridge address is empty")
	}

	switch fields[0] {
	case "obfs4", "snowflake", "meek":
		return fields[0], nil
	default:
		return "", nil
	}
}

func renderProxySettings(options TorrcOptions) (string, error) {
	address := strings.TrimSpace(options.HTTPProxyAddress)
	credential := strings.TrimSpace(options.HTTPProxyCredential)

	if address == "" {
		if credential != "" {
			return "", fmt.Errorf("proxy credential requires proxy address")
		}
		return "", nil
	}
	if err := validateTorrcSingleLine("proxy address", address); err != nil {
		return "", err
	}
	if credential != "" {
		if err := validateTorrcSingleLine("proxy credential", credential); err != nil {
			return "", err
		}
	}

	var builder strings.Builder
	builder.WriteString("# Set proxy address\n")
	if options.ProxyIsHTTPS {
		builder.WriteString("HTTPSProxy ")
		builder.WriteString(address)
		builder.WriteString("\n")
		if credential != "" {
			builder.WriteString("HTTPSProxyAuthenticator ")
			builder.WriteString(credential)
			builder.WriteString("\n")
		}
		return strings.TrimRight(builder.String(), "\n"), nil
	}

	if credential != "" {
		return "", fmt.Errorf("HTTP proxy credentials require proxy_is_https=true")
	}
	builder.WriteString("HTTPProxy ")
	builder.WriteString(address)
	return builder.String(), nil
}

func validateTorrcSingleLine(name, value string) error {
	if strings.ContainsAny(value, "\r\n") {
		return fmt.Errorf("%s must be a single line", name)
	}
	return nil
}

func torBool(value bool) string {
	if value {
		return "1"
	}
	return "0"
}

func normalizePlaintextPorts(portList string) (string, error) {
	parts := strings.Split(portList, ",")
	ports := make([]string, 0, len(parts))
	seen := make(map[int]bool)

	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			return "", fmt.Errorf("empty plaintext port")
		}

		port, err := strconv.Atoi(value)
		if err != nil || port < 1 || port > 65535 {
			return "", fmt.Errorf("invalid plaintext port %q", value)
		}
		if seen[port] {
			continue
		}
		seen[port] = true
		ports = append(ports, strconv.Itoa(port))
	}
	if len(ports) == 0 {
		return "", fmt.Errorf("no plaintext ports configured")
	}

	return strings.Join(ports, ","), nil
}
