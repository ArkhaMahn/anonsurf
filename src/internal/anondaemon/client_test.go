package anondaemon

import (
	"reflect"
	"testing"

	"gitlab.com/parrotsec/packages/anonsurf/internal/sys"
)

type recordedClientCall struct {
	command string
	args    []string
}

type recordingClientRunner struct {
	calls []recordedClientCall
}

func (r *recordingClientRunner) Run(command string, args ...string) error {
	r.calls = append(r.calls, recordedClientCall{command: command, args: append([]string{}, args...)})
	return nil
}

func TestMonitorClientRequestsPrivilegedBackend(t *testing.T) {
	runner := &recordingClientRunner{}
	client := MonitorClient{
		Runner: runner,
		Path:   "/test/anondaemon",
		Sudo: sys.Sudo{
			SudoPath: "/test/sudo",
			IsAdmin:  func() bool { return false },
		},
	}

	if err := client.Run(); err != nil {
		t.Fatal(err)
	}

	want := []recordedClientCall{{
		command: "/test/sudo",
		args:    []string{"/test/anondaemon", "monitor"},
	}}
	if !reflect.DeepEqual(runner.calls, want) {
		t.Fatalf("calls = %#v, want %#v", runner.calls, want)
	}
}
