package anondaemon

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gitlab.com/parrotsec/packages/anonsurf/internal/tor"
)

func TestStartRollsBackFirewallFailure(t *testing.T) {
	controller, runner, dnsManager, torService := testController(t)
	torService.active = true
	applyErr := errors.New("iptables failed")
	restoreCalls := 0
	controller.RunWithInput = func(command string, _ []byte, args ...string) error {
		if command == controller.IPTablesRestorePath {
			restoreCalls++
		}
		if restoreCalls == 1 {
			runner.runs = append(runner.runs, commandKey(command, args...))
			return applyErr
		}
		return runner.Run(command, args...)
	}

	err := controller.Start()
	if !errors.Is(err, applyErr) {
		t.Fatalf("Start() error = %v, want %v", err, applyErr)
	}
	if got := strings.Join(dnsManager.calls, ","); got != "backup,create,restore" {
		t.Fatalf("DNS calls = %q", got)
	}
	if torService.stops != 1 || torService.starts != 1 {
		t.Fatalf("Tor calls = %d stops, %d starts", torService.stops, torService.starts)
	}
	if !runner.called(controller.IPTablesRestorePath) {
		t.Fatal("iptables backup was not restored")
	}
	if !runner.called(controller.MakeTorrcPath, "restore") {
		t.Fatal("torrc was not restored")
	}
	for _, path := range []string{
		controller.IPTablesRulesPath,
		controller.DNSStatePath,
		controller.IPv6StatePath,
	} {
		if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
			t.Fatalf("state file %s still exists: %v", path, statErr)
		}
	}
	if err := controller.Stop(); err != nil {
		t.Fatal(err)
	}
	if torService.stops != 1 {
		t.Fatalf("cleanup stopped restored Tor service %d times", torService.stops)
	}
}

func TestStartAndStopPersistActiveState(t *testing.T) {
	controller, _, _, _ := testController(t)

	if err := controller.Start(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(controller.ActiveStatePath); err != nil {
		t.Fatalf("active state was not created: %v", err)
	}
	if err := controller.Stop(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(controller.ActiveStatePath); !os.IsNotExist(err) {
		t.Fatalf("active state still exists: %v", err)
	}
}

func TestApplyIPTablesRulesUsesAtomicRestore(t *testing.T) {
	controller, runner, _, _ := testController(t)
	backup := "*filter\n" +
		":INPUT DROP [12:34]\n" +
		":FORWARD DROP [0:0]\n" +
		":OUTPUT DROP [0:0]\n" +
		":CUSTOM - [0:0]\n" +
		"COMMIT\n"
	if err := controller.writeFile(controller.IPTablesRulesPath, []byte(backup), 0600); err != nil {
		t.Fatal(err)
	}

	var restored string
	restoreCalls := 0
	controller.RunWithInput = func(command string, input []byte, args ...string) error {
		restoreCalls++
		if command != controller.IPTablesRestorePath {
			t.Fatalf("restore command = %q", command)
		}
		restored = string(input)
		return nil
	}

	if err := controller.applyIPTablesRules(torRoutingConfig()); err != nil {
		t.Fatal(err)
	}
	if restoreCalls != 1 {
		t.Fatalf("restore calls = %d, want 1", restoreCalls)
	}
	if len(runner.runs) != 0 {
		t.Fatalf("unexpected sequential iptables calls: %#v", runner.runs)
	}
	for _, expected := range []string{
		"*filter\n",
		":INPUT DROP [0:0]\n",
		":CUSTOM - [0:0]\n",
		"-A OUTPUT -m owner --uid-owner debian-tor -j ACCEPT\n",
		"-A OUTPUT -j REJECT\n",
		"*nat\n",
		":PREROUTING ACCEPT [0:0]\n",
		"-A OUTPUT -m owner --uid-owner debian-tor -j RETURN\n",
	} {
		if !strings.Contains(restored, expected) {
			t.Fatalf("restore input missing %q:\n%s", expected, restored)
		}
	}
}

func TestStopAttemptsEveryRestoreAfterFailure(t *testing.T) {
	controller, runner, dnsManager, torService := testController(t)
	stopErr := errors.New("stop failed")
	restoreErr := errors.New("torrc restore failed")
	torService.stopErr = stopErr
	runner.failRun = func(command string, args []string) error {
		if command == controller.MakeTorrcPath && len(args) == 1 && args[0] == "restore" {
			return restoreErr
		}
		return nil
	}
	for path, content := range map[string]string{
		controller.IPTablesRulesPath: "*filter\nCOMMIT\n",
		controller.ActiveStatePath:   stateMarkerContent,
		controller.DNSStatePath:      stateMarkerContent,
		controller.IPv6StatePath:     "1\n0\n",
	} {
		if err := controller.writeFile(path, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}

	err := controller.Stop()
	if !errors.Is(err, stopErr) {
		t.Fatalf("Stop() error = %v, want %v", err, stopErr)
	}
	if !strings.Contains(err.Error(), restoreErr.Error()) {
		t.Fatalf("Stop() error does not include torrc failure: %v", err)
	}
	if !runner.called(controller.IPTablesRestorePath) {
		t.Fatal("iptables restore was not attempted")
	}
	if got := strings.Join(dnsManager.calls, ","); got != "restore" {
		t.Fatalf("DNS calls = %q", got)
	}
	if !runner.called(controller.SysctlPath, "-w", ipv6AllKey+"=1") {
		t.Fatal("IPv6 all state was not restored")
	}
	if !runner.called(controller.SysctlPath, "-w", ipv6DefaultKey+"=0") {
		t.Fatal("IPv6 default state was not restored")
	}
	if _, statErr := os.Stat(controller.ActiveStatePath); statErr != nil {
		t.Fatalf("active state was removed after failed cleanup: %v", statErr)
	}
}

func TestIPv6StateRestoresPerInterfaceOverrides(t *testing.T) {
	controller, runner, _, _ := testController(t)
	for name, value := range map[string]string{
		"lo":    "1\n",
		"wlan0": "0\n",
	} {
		path := filepath.Join(controller.IPv6ConfPath, name, "disable_ipv6")
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(value), 0644); err != nil {
			t.Fatal(err)
		}
	}

	if err := controller.captureIPv6State(); err != nil {
		t.Fatal(err)
	}
	if err := controller.restoreIPv6State(); err != nil {
		t.Fatal(err)
	}

	expected := []string{
		commandKey(controller.SysctlPath, "-w", ipv6AllKey+"=0"),
		commandKey(controller.SysctlPath, "-w", ipv6DefaultKey+"=0"),
		commandKey(controller.SysctlPath, "-w", "net.ipv6.conf.lo.disable_ipv6=1"),
		commandKey(controller.SysctlPath, "-w", "net.ipv6.conf.wlan0.disable_ipv6=0"),
	}
	if !reflect.DeepEqual(runner.runs, expected) {
		t.Fatalf("IPv6 restore calls = %#v, want %#v", runner.runs, expected)
	}
}

func TestRestoreFirewallCompletesMissingManagedTables(t *testing.T) {
	tests := []struct {
		name   string
		backup string
	}{
		{name: "empty backup"},
		{
			name: "filter only",
			backup: "*filter\n" +
				":INPUT DROP [0:0]\n" +
				":FORWARD DROP [0:0]\n" +
				":OUTPUT DROP [0:0]\n" +
				"-A OUTPUT -d 203.0.113.0/24 -j ACCEPT\n" +
				"COMMIT\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			controller, _, _, _ := testController(t)
			if err := controller.writeFile(controller.IPTablesRulesPath, []byte(test.backup), 0600); err != nil {
				t.Fatal(err)
			}

			var restored string
			controller.RunWithInput = func(command string, input []byte, args ...string) error {
				if command != controller.IPTablesRestorePath {
					t.Fatalf("restore command = %q", command)
				}
				restored = string(input)
				return nil
			}

			if err := controller.restoreFirewall(); err != nil {
				t.Fatal(err)
			}
			if strings.Count(restored, "*filter\n") != 1 {
				t.Fatalf("filter table count = %d:\n%s", strings.Count(restored, "*filter\n"), restored)
			}
			if strings.Count(restored, "*nat\n") != 1 {
				t.Fatalf("nat table count = %d:\n%s", strings.Count(restored, "*nat\n"), restored)
			}
			if test.backup != "" && !strings.Contains(restored, "-A OUTPUT -d 203.0.113.0/24 -j ACCEPT") {
				t.Fatalf("original filter rule missing:\n%s", restored)
			}
			if _, err := os.Stat(controller.IPTablesRulesPath); !os.IsNotExist(err) {
				t.Fatalf("iptables backup still exists: %v", err)
			}
		})
	}
}

func TestIPTablesRejectRuleIsLast(t *testing.T) {
	controller := Controller{}
	rules := controller.filterIPTablesRules(torRoutingConfig())
	last := rules[len(rules)-1]
	if got := strings.Join(last, " "); got != "-A OUTPUT -j REJECT" {
		t.Fatalf("last iptables rule = %q", got)
	}
}

func torRoutingConfig() tor.RoutingConfig {
	return tor.RoutingConfig{
		TransPort:          "9040",
		DNSPort:            "53",
		VirtualAddrNetwork: "10.192.0.0/10",
	}
}
