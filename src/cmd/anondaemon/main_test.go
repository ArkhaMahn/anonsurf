package main

import (
	"io"
	"testing"
)

type fakeDaemonController struct{}

func (fakeDaemonController) Start() error   { return nil }
func (fakeDaemonController) Stop() error    { return nil }
func (fakeDaemonController) Restart() error { return nil }

type fakeIdentityChanger struct{}

func (fakeIdentityChanger) Change() error { return nil }

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
		daemon:   fakeDaemonController{},
		identity: fakeIdentityChanger{},
		monitor:  monitor,
	}

	if err := run([]string{"monitor"}, io.Discard, io.Discard, deps); err != nil {
		t.Fatal(err)
	}
	if !monitor.called {
		t.Fatal("monitor backend was not called")
	}
}
