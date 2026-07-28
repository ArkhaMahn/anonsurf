package systemd

import (
	"errors"
	"reflect"
	"testing"
)

type recordedCall struct {
	command string
	args    []string
}

type recordingRunner struct {
	err    error
	output []byte
	calls  []recordedCall
}

func (r *recordingRunner) Run(command string, args ...string) error {
	r.calls = append(r.calls, recordedCall{command: command, args: append([]string{}, args...)})
	return r.err
}

func (r *recordingRunner) Output(command string, args ...string) ([]byte, error) {
	r.calls = append(r.calls, recordedCall{command: command, args: append([]string{}, args...)})
	return r.output, r.err
}

func TestNewAnonsurfdControllerDefaults(t *testing.T) {
	controller := NewAnonsurfdController()

	if controller.SystemctlPath != SystemctlPath {
		t.Fatalf("SystemctlPath = %q, want %q", controller.SystemctlPath, SystemctlPath)
	}
	if controller.Unit != AnonsurfdUnit {
		t.Fatalf("Unit = %q, want %q", controller.Unit, AnonsurfdUnit)
	}
	if controller.Runner == nil {
		t.Fatal("Runner is nil")
	}
}

func TestServiceControllerRunsSystemctlActions(t *testing.T) {
	tests := []struct {
		name   string
		run    func(ServiceController) error
		action string
	}{
		{name: "start", run: ServiceController.Start, action: "start"},
		{name: "stop", run: ServiceController.Stop, action: "stop"},
		{name: "restart", run: ServiceController.Restart, action: "restart"},
		{name: "enable", run: ServiceController.Enable, action: "enable"},
		{name: "disable", run: ServiceController.Disable, action: "disable"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := &recordingRunner{}
			controller := ServiceController{
				Runner:        runner,
				SystemctlPath: "/bin/systemctl",
				Unit:          "anonsurfd.service",
			}

			if err := test.run(controller); err != nil {
				t.Fatal(err)
			}
			want := []recordedCall{{
				command: "/bin/systemctl",
				args:    []string{test.action, "anonsurfd.service"},
			}}
			if !reflect.DeepEqual(runner.calls, want) {
				t.Fatalf("calls = %#v, want %#v", runner.calls, want)
			}
		})
	}
}

func TestServiceControllerUsesDefaultsForZeroValue(t *testing.T) {
	runner := &recordingRunner{}
	controller := ServiceController{Runner: runner}

	if err := controller.Start(); err != nil {
		t.Fatal(err)
	}

	want := []recordedCall{{
		command: SystemctlPath,
		args:    []string{"start", AnonsurfdUnit},
	}}
	if !reflect.DeepEqual(runner.calls, want) {
		t.Fatalf("calls = %#v, want %#v", runner.calls, want)
	}
}

func TestServiceControllerWrapsRunnerError(t *testing.T) {
	wantErr := errors.New("failed")
	controller := ServiceController{
		Runner: &recordingRunner{err: wantErr},
	}

	err := controller.Stop()
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want wrapping %v", err, wantErr)
	}
}

func TestServiceControllerIsEnabledStates(t *testing.T) {
	tests := []struct {
		name    string
		output  []byte
		enabled bool
	}{
		{name: "enabled", output: []byte("enabled\n"), enabled: true},
		{name: "disabled", output: []byte("disabled\n"), enabled: false},
		{name: "masked", output: []byte("masked\n"), enabled: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := &recordingRunner{output: test.output}
			controller := ServiceController{
				Runner:        runner,
				SystemctlPath: "/bin/systemctl",
				Unit:          "anonsurfd.service",
			}

			enabled, err := controller.IsEnabled()
			if err != nil {
				t.Fatal(err)
			}
			if enabled != test.enabled {
				t.Fatalf("enabled = %v, want %v", enabled, test.enabled)
			}
			want := []recordedCall{{
				command: "/bin/systemctl",
				args:    []string{"is-enabled", "anonsurfd.service"},
			}}
			if !reflect.DeepEqual(runner.calls, want) {
				t.Fatalf("calls = %#v, want %#v", runner.calls, want)
			}
		})
	}
}

func TestServiceControllerIsActiveStates(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		active bool
	}{
		{name: "active", output: []byte("active\n"), active: true},
		{name: "inactive", output: []byte("inactive\n"), active: false},
		{name: "failed", output: []byte("failed\n"), active: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := &recordingRunner{output: test.output}
			controller := ServiceController{
				Runner:        runner,
				SystemctlPath: "/bin/systemctl",
				Unit:          "tor.service",
			}

			active, err := controller.IsActive()
			if err != nil {
				t.Fatal(err)
			}
			if active != test.active {
				t.Fatalf("active = %v, want %v", active, test.active)
			}
			want := []recordedCall{{
				command: "/bin/systemctl",
				args:    []string{"is-active", "tor.service"},
			}}
			if !reflect.DeepEqual(runner.calls, want) {
				t.Fatalf("calls = %#v, want %#v", runner.calls, want)
			}
		})
	}
}

func TestServiceControllerIsEnabledWrapsUnexpectedError(t *testing.T) {
	wantErr := errors.New("failed")
	controller := ServiceController{
		Runner: &recordingRunner{err: wantErr},
	}

	_, err := controller.IsEnabled()
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want wrapping %v", err, wantErr)
	}
}

func TestServiceControllerIsEnabledRejectsUnexpectedState(t *testing.T) {
	controller := ServiceController{
		Runner: &recordingRunner{output: []byte("unknown\n")},
	}

	_, err := controller.IsEnabled()
	if err == nil {
		t.Fatal("expected unexpected state error")
	}
}
