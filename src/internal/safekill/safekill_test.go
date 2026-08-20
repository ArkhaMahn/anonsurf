package safekill

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"syscall"
	"testing"
)

func writeTestProcess(t *testing.T, root string, pid int, name, executable string) {
	t.Helper()
	processDir := filepath.Join(root, stringPID(pid))
	if err := os.MkdirAll(processDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(processDir, "comm"), []byte(name+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if executable != "" {
		if err := os.Symlink(executable, filepath.Join(processDir, "exe")); err != nil {
			t.Fatal(err)
		}
	}
}

func stringPID(pid int) string {
	return strconv.Itoa(pid)
}

func TestProcProcessKillerMatchesNamesAndPaths(t *testing.T) {
	root := t.TempDir()
	writeTestProcess(t, root, 101, "firefox-esr", "/usr/lib/firefox-esr/firefox-esr")
	writeTestProcess(t, root, 102, "unrelated", "/usr/bin/unrelated")

	var signaled []int
	killer := ProcProcessKiller{
		ProcRoot: root,
		EUID:     os.Geteuid,
		Signal: func(pid int) error {
			signaled = append(signaled, pid)
			return nil
		},
	}

	if err := killer.Kill([]string{"firefox-esr", "/usr/lib/firefox-esr/firefox-esr"}); err != nil {
		t.Fatal(err)
	}
	if want := []int{101}; !reflect.DeepEqual(signaled, want) {
		t.Fatalf("signaled = %#v, want %#v", signaled, want)
	}
}

func TestProcProcessKillerReturnsSignalFailure(t *testing.T) {
	root := t.TempDir()
	writeTestProcess(t, root, 101, "firefox", "/usr/bin/firefox")
	wantErr := errors.New("permission denied")
	killer := ProcProcessKiller{
		ProcRoot: root,
		EUID:     os.Geteuid,
		Signal: func(pid int) error {
			return wantErr
		},
	}

	err := killer.Kill([]string{"firefox"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
}

func TestProcProcessKillerAllowsNoMatches(t *testing.T) {
	root := t.TempDir()
	writeTestProcess(t, root, 101, "unrelated", "/usr/bin/unrelated")
	killer := ProcProcessKiller{
		ProcRoot: root,
		EUID:     os.Geteuid,
		Signal: func(pid int) error {
			t.Fatalf("unexpected signal for pid %d", pid)
			return nil
		},
	}

	if err := killer.Kill([]string{"firefox"}); err != nil {
		t.Fatal(err)
	}
}

func TestProcProcessKillerIgnoresExitedProcess(t *testing.T) {
	root := t.TempDir()
	writeTestProcess(t, root, 101, "firefox", "/usr/bin/firefox")
	killer := ProcProcessKiller{
		ProcRoot: root,
		EUID:     os.Geteuid,
		Signal: func(pid int) error {
			return syscall.ESRCH
		},
	}

	if err := killer.Kill([]string{"firefox"}); err != nil {
		t.Fatal(err)
	}
}

type fakeProcessKiller struct {
	err error
}

func (k fakeProcessKiller) Kill([]string) error {
	return k.err
}

type recordingCommandRunner struct {
	called bool
	err    error
}

func (r *recordingCommandRunner) Run(command string, args ...string) error {
	r.called = true
	return r.err
}

func TestRunnerStopsWhenProcessTerminationFails(t *testing.T) {
	wantErr := errors.New("terminate failed")
	executor := &recordingCommandRunner{}
	runner := Runner{
		Executor:  executor,
		Processes: fakeProcessKiller{err: wantErr},
	}

	err := runner.Run()
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
	if executor.called {
		t.Fatal("BleachBit ran after process termination failed")
	}
}

func TestRunnerRunsBleachBitAfterProcessScan(t *testing.T) {
	executor := &recordingCommandRunner{}
	runner := Runner{
		Executor:  executor,
		Processes: fakeProcessKiller{},
	}

	if err := runner.Run(); err != nil {
		t.Fatal(err)
	}
	if !executor.called {
		t.Fatal("BleachBit was not run")
	}
}
