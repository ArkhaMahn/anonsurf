package anondaemon

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type recordingRunner struct {
	runs     []string
	outputs  map[string][]byte
	failRun  func(command string, args []string) error
	failRead func(command string, args []string) error
}

func (r *recordingRunner) Run(command string, args ...string) error {
	r.runs = append(r.runs, commandKey(command, args...))
	if r.failRun != nil {
		return r.failRun(command, args)
	}
	return nil
}

func (r *recordingRunner) Output(command string, args ...string) ([]byte, error) {
	key := commandKey(command, args...)
	r.runs = append(r.runs, "output "+key)
	if r.failRead != nil {
		if err := r.failRead(command, args); err != nil {
			return nil, err
		}
	}
	output, ok := r.outputs[key]
	if !ok {
		return nil, fmt.Errorf("unexpected output command %s", key)
	}
	return output, nil
}

func (r *recordingRunner) called(command string, args ...string) bool {
	key := commandKey(command, args...)
	for _, call := range r.runs {
		if call == key {
			return true
		}
	}
	return false
}

func commandKey(command string, args ...string) string {
	return strings.Join(append([]string{command}, args...), " ")
}

type fakeDNSManager struct {
	calls      []string
	backupErr  error
	createErr  error
	restoreErr error
}

func (m *fakeDNSManager) MakeBackup() error {
	m.calls = append(m.calls, "backup")
	return m.backupErr
}

func (m *fakeDNSManager) CreateAddrs([]string) error {
	m.calls = append(m.calls, "create")
	return m.createErr
}

func (m *fakeDNSManager) RestoreBackupOrDHCP() error {
	m.calls = append(m.calls, "restore")
	return m.restoreErr
}

type fakeTorService struct {
	active      bool
	isActiveErr error
	startErr    error
	stopErr     error
	starts      int
	stops       int
}

func (s *fakeTorService) IsActive() (bool, error) {
	return s.active, s.isActiveErr
}

func (s *fakeTorService) Start() error {
	s.starts++
	if s.startErr == nil {
		s.active = true
	}
	return s.startErr
}

func (s *fakeTorService) Stop() error {
	s.stops++
	if s.stopErr == nil {
		s.active = false
	}
	return s.stopErr
}

func testController(t *testing.T) (Controller, *recordingRunner, *fakeDNSManager, *fakeTorService) {
	t.Helper()
	dir := t.TempDir()
	runner := &recordingRunner{outputs: make(map[string][]byte)}
	dnsManager := &fakeDNSManager{}
	torService := &fakeTorService{}
	controller := Controller{
		Runner:              runner,
		DNS:                 dnsManager,
		TorService:          torService,
		MakeTorrcPath:       "/test/make-torrc",
		SysctlPath:          "/test/sysctl",
		IPTablesPath:        "/test/iptables",
		IPTablesSavePath:    "/test/iptables-save",
		IPTablesRestorePath: "/test/iptables-restore",
		IPTablesRulesPath:   filepath.Join(dir, "state", "iptables.rules"),
		DNSStatePath:        filepath.Join(dir, "state", "dns.changed"),
		IPv6StatePath:       filepath.Join(dir, "state", "ipv6.state"),
		TorrcPath:           filepath.Join(dir, "etc", "tor", "torrc"),
	}
	controller.GenerateTorrc = func() error {
		if err := os.MkdirAll(filepath.Dir(controller.TorrcPath), 0755); err != nil {
			return err
		}
		content := "VirtualAddrNetwork 10.192.0.0/10\nTransPort 127.0.0.1:9040\nDNSPort 127.0.0.1:53\n"
		return os.WriteFile(controller.TorrcPath, []byte(content), 0600)
	}
	controller.RunWithInput = func(command string, _ []byte, args ...string) error {
		return runner.Run(command, args...)
	}
	runner.outputs[commandKey(controller.SysctlPath, "-n", ipv6AllKey)] = []byte("0\n")
	runner.outputs[commandKey(controller.SysctlPath, "-n", ipv6DefaultKey)] = []byte("0\n")
	runner.outputs[commandKey(controller.IPTablesSavePath)] = []byte("*filter\nCOMMIT\n")
	return controller, runner, dnsManager, torService
}
