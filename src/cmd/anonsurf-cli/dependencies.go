package main

import (
	"context"

	"gitlab.com/parrotsec/packages/anonsurf/internal/anondaemon"
	"gitlab.com/parrotsec/packages/anonsurf/internal/netcheck"
	"gitlab.com/parrotsec/packages/anonsurf/internal/safekill"
	"gitlab.com/parrotsec/packages/anonsurf/internal/status"
	"gitlab.com/parrotsec/packages/anonsurf/internal/systemd"
)

type serviceController interface {
	Start() error
	Stop() error
	Restart() error
	Enable() error
	Disable() error
	IsEnabled() (bool, error)
}

type statusReader interface {
	Snapshot() (status.Snapshot, error)
}

type identityChanger interface {
	Change() error
}

type ipChecker interface {
	Check(ctx context.Context) (netcheck.Result, error)
}

type monitorRunner interface {
	Run() error
}

type safeKillRunner interface {
	Run() error
}

type dependencies struct {
	service  serviceController
	status   statusReader
	identity identityChanger
	ipCheck  ipChecker
	monitor  monitorRunner
	safeKill safeKillRunner
}

func defaultDependencies() dependencies {
	return dependencies{
		service:  systemd.NewAnonsurfdController(),
		status:   status.Reader{},
		identity: anondaemon.IdentityClient{},
		ipCheck:  netcheck.Checker{},
		monitor:  anondaemon.MonitorClient{},
		safeKill: safekill.Runner{},
	}
}
