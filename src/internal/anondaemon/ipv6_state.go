package anondaemon

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

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
