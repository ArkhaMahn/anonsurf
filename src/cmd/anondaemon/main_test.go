package main

import (
	"testing"
)

type fakeMonitorRunner struct {
	called bool
}

func (m *fakeMonitorRunner) Run() error {
	m.called = true
	return nil
}

func TestRunMonitorUsesMonitorBackend(t *testing.T) {
	monitor := &fakeMonitorRunner{}
	deps := dependencies{
		monitor: monitor,
	}

	if err := run([]string{"monitor"}, deps); err != nil {
		t.Fatal(err)
	}
	if !monitor.called {
		t.Fatal("monitor backend was not called")
	}
}
