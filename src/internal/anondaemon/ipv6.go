package anondaemon

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	ipv6AllKey     = "net.ipv6.conf.all.disable_ipv6"
	ipv6DefaultKey = "net.ipv6.conf.default.disable_ipv6"
)

type ipv6Setting struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ipv6State struct {
	Settings []ipv6Setting `json:"settings"`
}

func (c Controller) captureIPv6State() error {
	path := IPv6StatePath
	exists, err := fileExists(path)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	state, err := c.currentIPv6State()
	if err != nil {
		return err
	}
	content, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return writeFile(path, append(content, '\n'), 0600)
}

func (c Controller) currentIPv6State() (ipv6State, error) {
	entries, err := os.ReadDir(IPv6ConfPath)
	if err != nil {
		return ipv6State{}, fmt.Errorf("read IPv6 interfaces: %w", err)
	}

	var interfaces []string
	for _, entry := range entries {
		if entry.IsDir() {
			interfaces = append(interfaces, entry.Name())
		}
	}
	sort.SliceStable(interfaces, func(i, j int) bool {
		return ipv6InterfaceOrder(interfaces[i]) < ipv6InterfaceOrder(interfaces[j])
	})

	state := ipv6State{Settings: make([]ipv6Setting, 0, len(interfaces))}
	for _, name := range interfaces {
		path := filepath.Join(IPv6ConfPath, name, "disable_ipv6")
		data, err := readFile(path)
		if err != nil {
			return ipv6State{}, fmt.Errorf("read IPv6 state for %s: %w", name, err)
		}
		value := strings.TrimSpace(string(data))
		if value != "0" && value != "1" {
			return ipv6State{}, fmt.Errorf("read IPv6 state for %s: unexpected value %q", name, value)
		}
		state.Settings = append(state.Settings, ipv6Setting{
			Key:   "net.ipv6.conf." + name + ".disable_ipv6",
			Value: value,
		})
	}
	if err := validateIPv6State(state); err != nil {
		return ipv6State{}, err
	}
	return state, nil
}

func ipv6InterfaceOrder(name string) string {
	switch name {
	case "all":
		return "0"
	case "default":
		return "1"
	default:
		return "2" + name
	}
}

func (c Controller) disableIPv6() error {
	return c.setIPv6State(ipv6State{Settings: []ipv6Setting{
		{Key: ipv6AllKey, Value: "1"},
		{Key: ipv6DefaultKey, Value: "1"},
	}})
}

func (c Controller) restoreIPv6State() error {
	path := IPv6StatePath
	data, err := readFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read IPv6 state: %w", err)
	}
	state, err := parseIPv6State(string(data))
	if err != nil {
		return err
	}
	if err := c.restoreExistingIPv6State(state); err != nil {
		return err
	}
	if err := removeFile(path); err != nil {
		return fmt.Errorf("remove IPv6 state: %w", err)
	}
	return nil
}
