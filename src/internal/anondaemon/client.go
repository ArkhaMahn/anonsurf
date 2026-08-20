package anondaemon

import (
	"fmt"

	"gitlab.com/parrotsec/packages/anonsurf/internal/sys"
	"gitlab.com/parrotsec/packages/anonsurf/internal/systemd"
)

const BinaryPath = "/usr/lib/anonsurf/anondaemon"

type ClientRunner interface {
	Run(command string, args ...string) error
}

type IdentityClient struct {
	Runner ClientRunner
	Path   string
	Sudo   sys.Sudo
}

type MonitorClient struct {
	Runner ClientRunner
	Path   string
	Sudo   sys.Sudo
}

func NewIdentityClient() IdentityClient {
	return IdentityClient{
		Runner: systemd.ExecRunner{},
		Path:   BinaryPath,
	}
}

func NewMonitorClient() MonitorClient {
	return MonitorClient{
		Runner: systemd.ExecRunner{},
		Path:   BinaryPath,
	}
}

func (c IdentityClient) Change() error {
	runner := c.Runner
	if runner == nil {
		runner = systemd.ExecRunner{}
	}
	path := c.Path
	if path == "" {
		path = BinaryPath
	}
	command, args := c.Sudo.Command(path, "changeid")
	if err := runner.Run(command, args...); err != nil {
		return fmt.Errorf("request privileged identity change: %w", err)
	}
	return nil
}

func (c MonitorClient) Run() error {
	runner := c.Runner
	if runner == nil {
		runner = systemd.ExecRunner{}
	}
	path := c.Path
	if path == "" {
		path = BinaryPath
	}
	command, args := c.Sudo.Command(path, "monitor")
	if err := runner.Run(command, args...); err != nil {
		return fmt.Errorf("open privileged Tor monitor: %w", err)
	}
	return nil
}
