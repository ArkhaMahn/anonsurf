package tor

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"

	"gitlab.com/parrotsec/packages/anonsurf/internal/config"
)

//go:embed assets/torrc.base
var defaultTorrcBase string

// TorrcOptions contains the config values that can be rendered without system side effects.
type TorrcOptions struct {
	UseSandbox       bool
	UseSafeSocks     bool
	RejectPlainPorts bool
	PlaintextPorts   string
}

// TorrcOptionsFromConfig keeps config parsing separate from Tor-specific rendering rules.
func TorrcOptionsFromConfig(cfg config.Config) TorrcOptions {
	return TorrcOptions{
		UseSandbox:       cfg.UseSandbox,
		UseSafeSocks:     cfg.UseSafeSocks,
		RejectPlainPorts: cfg.RejectPlainPorts,
		PlaintextPorts:   cfg.ListPlainPorts,
	}
}

// RenderTorrc renders the embedded Parrot torrc base with the supplied options.
func RenderTorrc(options TorrcOptions) (string, error) {
	return RenderTorrcWithBase(defaultTorrcBase, options)
}

// RenderTorrcWithBase keeps rendering testable without reading or writing system files.
func RenderTorrcWithBase(base string, options TorrcOptions) (string, error) {
	base = strings.TrimRight(base, "\n")
	if strings.TrimSpace(base) == "" {
		return "", fmt.Errorf("torrc base is empty")
	}

	plaintextPorts, err := normalizePlaintextPorts(options.PlaintextPorts)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	builder.WriteString(base)
	builder.WriteString("\n\n# Use sandbox\n")
	builder.WriteString("Sandbox ")
	builder.WriteString(torBool(options.UseSandbox))
	builder.WriteString("\n\n# Use SafeSocks\n")
	builder.WriteString("SafeSocks ")
	builder.WriteString(torBool(options.UseSafeSocks))
	builder.WriteString("\n\n")
	if options.RejectPlainPorts {
		builder.WriteString("# Set RejectPlaintextPorts\nRejectPlaintextPorts ")
	} else {
		builder.WriteString("# Set WarnPlaintextPorts\nWarnPlaintextPorts ")
	}
	builder.WriteString(plaintextPorts)
	builder.WriteString("\n")

	return builder.String(), nil
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
