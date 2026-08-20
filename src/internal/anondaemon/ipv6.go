package anondaemon

import (
	"fmt"
	"os"
	"strings"
)

const (
	ipv6AllKey     = "net.ipv6.conf.all.disable_ipv6"
	ipv6DefaultKey = "net.ipv6.conf.default.disable_ipv6"
)

type ipv6State struct {
	all          string
	defaultValue string
}

func (c Controller) captureIPv6State() error {
	path := c.ipv6StatePath()
	exists, err := c.fileExists(path)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	all, err := c.readSysctlValue(ipv6AllKey)
	if err != nil {
		return err
	}
	defaultValue, err := c.readSysctlValue(ipv6DefaultKey)
	if err != nil {
		return err
	}
	content := all + "\n" + defaultValue + "\n"
	return c.writeFile(path, []byte(content), 0600)
}

func (c Controller) readSysctlValue(key string) (string, error) {
	output, err := c.runner().Output(c.sysctlPath(), "-n", key)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", key, err)
	}
	value := strings.TrimSpace(string(output))
	if value != "0" && value != "1" {
		return "", fmt.Errorf("read %s: unexpected value %q", key, value)
	}
	return value, nil
}

func (c Controller) disableIPv6() error {
	return c.setIPv6State(ipv6State{all: "1", defaultValue: "1"})
}

func (c Controller) restoreIPv6State() error {
	path := c.ipv6StatePath()
	data, err := c.readFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("IPv6 state is missing")
		}
		return fmt.Errorf("read IPv6 state: %w", err)
	}
	state, err := parseIPv6State(string(data))
	if err != nil {
		return err
	}
	if err := c.setIPv6State(state); err != nil {
		return err
	}
	if err := c.removeFile(path); err != nil {
		return fmt.Errorf("remove IPv6 state: %w", err)
	}
	return nil
}

func parseIPv6State(content string) (ipv6State, error) {
	values := strings.Fields(content)
	if len(values) != 2 {
		return ipv6State{}, fmt.Errorf("invalid IPv6 state")
	}
	for _, value := range values {
		if value != "0" && value != "1" {
			return ipv6State{}, fmt.Errorf("invalid IPv6 state value %q", value)
		}
	}
	return ipv6State{all: values[0], defaultValue: values[1]}, nil
}

func (c Controller) setIPv6State(state ipv6State) error {
	var failures errorCollector
	settings := []struct {
		key   string
		value string
	}{
		{key: ipv6AllKey, value: state.all},
		{key: ipv6DefaultKey, value: state.defaultValue},
	}
	for _, setting := range settings {
		if err := c.run(c.sysctlPath(), "-w", setting.key+"="+setting.value); err != nil {
			failures.add(setting.key, err)
		}
	}
	return failures.err()
}
