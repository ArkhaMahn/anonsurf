package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

type fakeDaemon struct {
	err   error
	calls []string
}

func (d *fakeDaemon) Start() error {
	d.calls = append(d.calls, "start")
	return d.err
}

func (d *fakeDaemon) Stop() error {
	d.calls = append(d.calls, "stop")
	return d.err
}

func (d *fakeDaemon) Restart() error {
	d.calls = append(d.calls, "restart")
	return d.err
}

func TestRunDispatchesSupportedCommands(t *testing.T) {
	tests := []struct {
		command string
		call    string
	}{
		{command: "start", call: "start"},
		{command: "stop", call: "stop"},
		{command: "restart", call: "restart"},
	}

	for _, test := range tests {
		t.Run(test.command, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			daemon := &fakeDaemon{}

			err := run([]string{test.command}, &stdout, &stderr, dependencies{daemon: daemon})
			if err != nil {
				t.Fatal(err)
			}
			if len(daemon.calls) != 1 || daemon.calls[0] != test.call {
				t.Fatalf("calls = %#v, want %q", daemon.calls, test.call)
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
			if stderr.Len() != 0 {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}
		})
	}
}

func TestRunReturnsBackendError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	wantErr := errors.New("failed")
	daemon := &fakeDaemon{err: wantErr}

	err := run([]string{"start"}, &stdout, &stderr, dependencies{daemon: daemon})
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want wrapping %v", err, wantErr)
	}
	if len(daemon.calls) != 1 || daemon.calls[0] != "start" {
		t.Fatalf("calls = %#v, want start", daemon.calls)
	}
}

func TestRunRejectsInvalidArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "missing command", args: nil, want: "expected one command"},
		{name: "too many commands", args: []string{"start", "stop"}, want: "expected one command"},
		{name: "unknown command", args: []string{"status"}, want: `unknown command "status"`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			daemon := &fakeDaemon{}

			err := run(test.args, &stdout, &stderr, dependencies{daemon: daemon})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want containing %q", err, test.want)
			}
			if len(daemon.calls) != 0 {
				t.Fatalf("calls = %#v, want empty", daemon.calls)
			}
		})
	}
}

func TestRunRejectsMissingBackend(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := run([]string{"start"}, &stdout, &stderr, dependencies{})
	if err == nil || !strings.Contains(err.Error(), "daemon backend is missing") {
		t.Fatalf("error = %v, want missing backend", err)
	}
}
