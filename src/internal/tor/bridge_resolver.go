package tor

import (
	"bufio"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"os/exec"
	"strings"
)

// BridgesPath follows the Parrot/Debian AnonSurf package layout.
const BridgesPath = "/etc/anonsurf/bridges.txt"

var ErrNoBridgeEntries = errors.New("no bridge entries found")

// ExecutableLookup is injected so tests do not depend on host-installed transports.
type ExecutableLookup func(name string) (string, error)

// BridgeSelector is injected because auto-bridge selection is intentionally non-deterministic.
type BridgeSelector func(entries []string) (string, error)

// BridgeResolver turns config-derived bridge options into render-ready torrc options.
type BridgeResolver struct {
	BridgeListPath string

	LookupExecutable ExecutableLookup
	SelectBridge     BridgeSelector
}

func NewBridgeResolver() BridgeResolver {
	return BridgeResolver{
		BridgeListPath:   BridgesPath,
		LookupExecutable: exec.LookPath,
		SelectBridge:     SelectRandomBridge,
	}
}

func ParseBridgeEntries(reader io.Reader) ([]string, error) {
	var entries []string
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		entries = append(entries, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, ErrNoBridgeEntries
	}

	return entries, nil
}

func LoadBridgeEntries(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ParseBridgeEntries(file)
}

// SelectRandomBridge selects one bridge using crypto/rand to avoid predictable defaults.
func SelectRandomBridge(entries []string) (string, error) {
	if len(entries) == 0 {
		return "", ErrNoBridgeEntries
	}

	index, err := rand.Int(rand.Reader, big.NewInt(int64(len(entries))))
	if err != nil {
		return "", err
	}
	return entries[index.Int64()], nil
}

// Resolve selects auto bridges and fills transport plugin paths without rendering torrc.
func (r BridgeResolver) Resolve(options TorrcOptions) (TorrcOptions, error) {
	if !options.AutoBridge && strings.TrimSpace(options.BridgeAddress) == "" {
		return options, nil
	}

	if r.BridgeListPath == "" {
		r.BridgeListPath = BridgesPath
	}
	if r.LookupExecutable == nil {
		r.LookupExecutable = exec.LookPath
	}
	if r.SelectBridge == nil {
		r.SelectBridge = SelectRandomBridge
	}

	if options.AutoBridge {
		entries, err := LoadBridgeEntries(r.BridgeListPath)
		if err != nil {
			return options, fmt.Errorf("load bridge entries: %w", err)
		}
		selected, err := r.SelectBridge(entries)
		if err != nil {
			return options, fmt.Errorf("select bridge: %w", err)
		}
		options.BridgeAddress = selected
	}

	transport, err := bridgeTransport(options.BridgeAddress)
	if err != nil {
		return options, err
	}
	switch transport {
	case "obfs4":
		if strings.TrimSpace(options.Obfs4ProxyPath) == "" {
			path, err := lookupExecutablePath(r.LookupExecutable, "obfs4proxy")
			if err != nil {
				return options, fmt.Errorf("find obfs4proxy: %w", err)
			}
			options.Obfs4ProxyPath = path
		}
	case "snowflake":
		if strings.TrimSpace(options.SnowflakeClientPath) == "" {
			path, err := lookupExecutablePath(r.LookupExecutable, "snowflake-client")
			if err != nil {
				return options, fmt.Errorf("find snowflake-client: %w", err)
			}
			options.SnowflakeClientPath = path
		}
	}

	return options, nil
}

func lookupExecutablePath(lookup ExecutableLookup, name string) (string, error) {
	path, err := lookup(name)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("%s path is empty", name)
	}
	return path, nil
}
