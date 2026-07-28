package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"gitlab.com/parrotsec/packages/anonsurf/internal/status"
)

type fakeService struct {
	startErr   error
	stopErr    error
	restartErr error
	enableErr  error
	disableErr error
	statusErr  error
	enabled    bool
	calls      []string
}

func (s *fakeService) Start() error {
	s.calls = append(s.calls, "start")
	return s.startErr
}

func (s *fakeService) Stop() error {
	s.calls = append(s.calls, "stop")
	return s.stopErr
}

func (s *fakeService) Restart() error {
	s.calls = append(s.calls, "restart")
	return s.restartErr
}

func (s *fakeService) Enable() error {
	s.calls = append(s.calls, "enable")
	return s.enableErr
}

func (s *fakeService) Disable() error {
	s.calls = append(s.calls, "disable")
	return s.disableErr
}

func (s *fakeService) IsEnabled() (bool, error) {
	s.calls = append(s.calls, "is-enabled")
	return s.enabled, s.statusErr
}

func testDependencies(service *fakeService) dependencies {
	return dependencies{
		service:  service,
		status:   fakeStatus{snapshot: status.Snapshot{}},
		identity: &fakeIdentity{},
	}
}

type fakeStatus struct {
	snapshot status.Snapshot
	err      error
}

func (s fakeStatus) Snapshot() (status.Snapshot, error) {
	return s.snapshot, s.err
}

type fakeIdentity struct {
	err    error
	called bool
}

func (i *fakeIdentity) Change() error {
	i.called = true
	return i.err
}

func TestRootHelpWithNoArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer

	if err := execute(nil, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}

	output := stdout.String()
	for _, expected := range []string{
		"AnonSurf manages Parrot OS system-wide Tor routing",
		"Commands:",
		"start",
		"enable-boot",
		"Examples:",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("help output missing %q:\n%s", expected, output)
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRootHelpFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer

	if err := execute([]string{"--help"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(stdout.String(), "status") {
		t.Fatalf("help output missing status command:\n%s", stdout.String())
	}
}

func TestVersionFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer

	if err := execute([]string{"--version"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}

	if got := stdout.String(); got != "AnonSurf 6.0.0\n" {
		t.Fatalf("version output = %q", got)
	}
}

func TestServiceCommandsRunController(t *testing.T) {
	tests := []struct {
		command string
		output  string
		call    string
	}{
		{command: "start", output: "AnonSurf started.\n", call: "start"},
		{command: "stop", output: "AnonSurf stopped.\n", call: "stop"},
		{command: "restart", output: "AnonSurf restarted.\n", call: "restart"},
		{command: "enable-boot", output: "AnonSurf enabled at startup.\n", call: "enable"},
		{command: "disable-boot", output: "AnonSurf disabled at startup.\n", call: "disable"},
	}

	for _, test := range tests {
		t.Run(test.command, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			service := &fakeService{}
			err := executeWithDependencies([]string{test.command}, &stdout, &stderr, testDependencies(service))
			if err != nil {
				t.Fatal(err)
			}
			if got := stdout.String(); got != test.output {
				t.Fatalf("stdout = %q, want %q", got, test.output)
			}
			if len(service.calls) != 1 || service.calls[0] != test.call {
				t.Fatalf("service calls = %#v, want %q", service.calls, test.call)
			}
		})
	}
}

func TestStatusBootCommandReportsEnabledState(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
		output  string
	}{
		{name: "enabled", enabled: true, output: "AnonSurf is enabled at startup.\n"},
		{name: "disabled", enabled: false, output: "AnonSurf is disabled at startup.\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			service := &fakeService{enabled: test.enabled}
			err := executeWithDependencies([]string{"status-boot"}, &stdout, &stderr, testDependencies(service))
			if err != nil {
				t.Fatal(err)
			}
			if got := stdout.String(); got != test.output {
				t.Fatalf("stdout = %q, want %q", got, test.output)
			}
			if len(service.calls) != 1 || service.calls[0] != "is-enabled" {
				t.Fatalf("service calls = %#v, want is-enabled", service.calls)
			}
		})
	}
}

func TestServiceCommandReturnsContextualError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	wantErr := errors.New("systemctl failed")
	service := &fakeService{enableErr: wantErr}

	err := executeWithDependencies([]string{"enable-boot"}, &stdout, &stderr, testDependencies(service))
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want wrapping %v", err, wantErr)
	}
	if !strings.Contains(err.Error(), "enable-boot:") {
		t.Fatalf("error = %v, want command context", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestStatusBootCommandReturnsContextualError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	wantErr := errors.New("systemctl failed")
	service := &fakeService{statusErr: wantErr}

	err := executeWithDependencies([]string{"status-boot"}, &stdout, &stderr, testDependencies(service))
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want wrapping %v", err, wantErr)
	}
	if !strings.Contains(err.Error(), "status-boot:") {
		t.Fatalf("error = %v, want command context", err)
	}
}

func TestStatusCommandPrintsSnapshot(t *testing.T) {
	var stdout, stderr bytes.Buffer
	deps := testDependencies(&fakeService{})
	deps.status = fakeStatus{
		snapshot: status.Snapshot{
			AnonSurfActive: true,
			TorActive:      true,
			DNSState:       status.DNSTor,
			DNSAddresses:   []string{"127.0.0.1"},
		},
	}

	if err := executeWithDependencies([]string{"status"}, &stdout, &stderr, deps); err != nil {
		t.Fatal(err)
	}

	output := stdout.String()
	for _, expected := range []string{
		"AnonSurf status",
		"AnonSurf: active",
		"Tor:      active",
		"DNS:      using Tor DNS",
		"Servers:  127.0.0.1",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("status output missing %q:\n%s", expected, output)
		}
	}
}

func TestStatusCommandReturnsContextualError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	wantErr := errors.New("status failed")
	deps := testDependencies(&fakeService{})
	deps.status = fakeStatus{err: wantErr}

	err := executeWithDependencies([]string{"status"}, &stdout, &stderr, deps)
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want wrapping %v", err, wantErr)
	}
	if !strings.Contains(err.Error(), "status:") {
		t.Fatalf("error = %v, want command context", err)
	}
}

func TestChangeIDCommandRequestsNewIdentity(t *testing.T) {
	var stdout, stderr bytes.Buffer
	deps := testDependencies(&fakeService{})
	identity := &fakeIdentity{}
	deps.identity = identity

	if err := executeWithDependencies([]string{"changeid"}, &stdout, &stderr, deps); err != nil {
		t.Fatal(err)
	}
	if !identity.called {
		t.Fatal("expected identity changer to be called")
	}
	if got := stdout.String(); got != "Tor identity change requested.\n" {
		t.Fatalf("stdout = %q", got)
	}
}

func TestChangeIDAliasRequestsNewIdentity(t *testing.T) {
	var stdout, stderr bytes.Buffer
	deps := testDependencies(&fakeService{})
	identity := &fakeIdentity{}
	deps.identity = identity

	if err := executeWithDependencies([]string{"newnym"}, &stdout, &stderr, deps); err != nil {
		t.Fatal(err)
	}
	if !identity.called {
		t.Fatal("expected identity changer to be called")
	}
}

func TestChangeIDCommandReturnsContextualError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	wantErr := errors.New("tor control failed")
	deps := testDependencies(&fakeService{})
	deps.identity = &fakeIdentity{err: wantErr}

	err := executeWithDependencies([]string{"changeid"}, &stdout, &stderr, deps)
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want wrapping %v", err, wantErr)
	}
	if !strings.Contains(err.Error(), "changeid:") {
		t.Fatalf("error = %v, want command context", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestPlannedCommandsReturnNotImplemented(t *testing.T) {
	commands := []string{
		"myip",
	}

	for _, command := range commands {
		t.Run(command, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := executeWithDependencies([]string{command}, &stdout, &stderr, testDependencies(&fakeService{}))
			if err == nil {
				t.Fatal("expected not implemented error")
			}
			if !strings.Contains(err.Error(), "not implemented yet") {
				t.Fatalf("error = %v, want not implemented message", err)
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
		})
	}
}

func TestCommandAliasesReturnNotImplemented(t *testing.T) {
	tests := map[string]string{
		"ip": "myip",
	}

	for alias, command := range tests {
		t.Run(alias, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := executeWithDependencies([]string{alias}, &stdout, &stderr, testDependencies(&fakeService{}))
			if err == nil {
				t.Fatal("expected not implemented error")
			}
			if !strings.Contains(err.Error(), command) {
				t.Fatalf("error = %v, want command %q", err, command)
			}
		})
	}
}

func TestCommandHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	if err := execute([]string{"start", "--help"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}

	output := stdout.String()
	for _, expected := range []string{
		"Start system-wide Tor transparent routing",
		"Usage:",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("command help missing %q:\n%s", expected, output)
		}
	}
}

func TestUnknownCommandReturnsError(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := execute([]string{"unknown"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected unknown command error")
	}
	if !strings.Contains(err.Error(), `unknown command "unknown"`) {
		t.Fatalf("error = %v", err)
	}
}
