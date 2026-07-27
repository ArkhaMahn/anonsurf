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
	err   error
	calls []recordedCall
}

func (r *recordingRunner) Run(command string, args ...string) error {
	r.calls = append(r.calls, recordedCall{command: command, args: append([]string{}, args...)})
	return r.err
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
