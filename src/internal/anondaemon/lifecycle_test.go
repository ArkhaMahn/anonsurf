package anondaemon

import (
	"errors"
	"os"
	"strings"
	"testing"

	"gitlab.com/parrotsec/packages/anonsurf/internal/tor"
)

func TestStartRollsBackFirewallFailure(t *testing.T) {
	controller, runner, dnsManager, torService := testController(t)
	torService.active = true
	applyErr := errors.New("iptables failed")
	runner.failRun = func(command string, args []string) error {
		if command == controller.IPTablesPath && len(args) == 1 && args[0] == "-F" {
			return applyErr
		}
		return nil
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
}

func TestStopDoesNotFlushFirewallWithoutBackup(t *testing.T) {
	controller, runner, _, _ := testController(t)

	if err := controller.Stop(); err != nil {
		t.Fatal(err)
	}

	if runner.called(controller.IPTablesPath, "-F") {
		t.Fatal("firewall was flushed without an AnonSurf backup")
	}
}

func TestIPTablesRejectRuleIsLast(t *testing.T) {
	controller := Controller{}
	rules := controller.iptablesRules(torRoutingConfig())
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
