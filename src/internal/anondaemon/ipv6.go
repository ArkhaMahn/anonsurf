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

func parseIPv6State(content string) (ipv6State, error) {
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "{") {
		var state ipv6State
		if err := json.Unmarshal([]byte(content), &state); err != nil {
			return ipv6State{}, fmt.Errorf("invalid IPv6 state: %w", err)
		}
		if err := validateIPv6State(state); err != nil {
			return ipv6State{}, err
		}
		return state, nil
	}

	values := strings.Fields(content)
	if len(values) != 2 {
		return ipv6State{}, fmt.Errorf("invalid IPv6 state")
	}
	for _, value := range values {
		if value != "0" && value != "1" {
			return ipv6State{}, fmt.Errorf("invalid IPv6 state value %q", value)
		}
	}
	return ipv6State{Settings: []ipv6Setting{
		{Key: ipv6AllKey, Value: values[0]},
		{Key: ipv6DefaultKey, Value: values[1]},
	}}, nil
}

func validateIPv6State(state ipv6State) error {
	seen := make(map[string]bool)
	required := map[string]bool{
		ipv6AllKey:     false,
		ipv6DefaultKey: false,
	}
	for _, setting := range state.Settings {
		if _, ok := ipv6InterfaceName(setting.Key); !ok {
			return fmt.Errorf("invalid IPv6 state key %q", setting.Key)
		}
		if setting.Value != "0" && setting.Value != "1" {
			return fmt.Errorf("invalid IPv6 state value %q", setting.Value)
		}
		if seen[setting.Key] {
			return fmt.Errorf("duplicate IPv6 state key %q", setting.Key)
		}
		seen[setting.Key] = true
		if _, ok := required[setting.Key]; ok {
			required[setting.Key] = true
		}
	}
	for key, present := range required {
		if !present {
			return fmt.Errorf("missing IPv6 state key %q", key)
		}
	}
	return nil
}

func (c Controller) restoreExistingIPv6State(state ipv6State) error {
	var failures errorCollector
	settings := append([]ipv6Setting{}, state.Settings...)
	sort.SliceStable(settings, func(i, j int) bool {
		return ipv6SettingOrder(settings[i].Key) < ipv6SettingOrder(settings[j].Key)
	})
	for _, setting := range settings {
		if setting.Key != ipv6AllKey && setting.Key != ipv6DefaultKey {
			exists, err := c.ipv6SettingExists(setting.Key)
			if err != nil {
				failures.add(setting.Key, err)
				continue
			}
			if !exists {
				continue
			}
		}
		if err := c.run(SysctlPath, "-w", setting.Key+"="+setting.Value); err != nil {
			failures.add(setting.Key, err)
		}
	}
	return failures.err()
}

func ipv6SettingOrder(key string) string {
	switch key {
	case ipv6AllKey:
		return "0"
	case ipv6DefaultKey:
		return "1"
	default:
		return "2" + key
	}
}

func (c Controller) ipv6SettingExists(key string) (bool, error) {
	name, ok := ipv6InterfaceName(key)
	if !ok {
		return false, fmt.Errorf("invalid IPv6 state key %q", key)
	}
	return fileExists(filepath.Join(IPv6ConfPath, name, "disable_ipv6"))
}

func ipv6InterfaceName(key string) (string, bool) {
	const prefix = "net.ipv6.conf."
	const suffix = ".disable_ipv6"
	if !strings.HasPrefix(key, prefix) || !strings.HasSuffix(key, suffix) {
		return "", false
	}
	name := strings.TrimSuffix(strings.TrimPrefix(key, prefix), suffix)
	if name == "" || strings.ContainsAny(name, "/= \t\r\n") {
		return "", false
	}
	return name, true
}

func (c Controller) setIPv6State(state ipv6State) error {
	var failures errorCollector
	for _, setting := range state.Settings {
		if err := c.run(SysctlPath, "-w", setting.Key+"="+setting.Value); err != nil {
			failures.add(setting.Key, err)
		}
	}
	return failures.err()
}
