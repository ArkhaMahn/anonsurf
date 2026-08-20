package anondaemon

import (
	"fmt"

	"gitlab.com/parrotsec/packages/anonsurf/internal/sys"
	"gitlab.com/parrotsec/packages/anonsurf/internal/systemd"
)

const BinaryPath = "/usr/lib/anonsurf/anondaemon"

type IdentityClient struct{}

type MonitorClient struct{}

func (IdentityClient) Change() error {
	command, args := sys.PrivilegedCommand(BinaryPath, "changeid")
	if err := (systemd.ExecRunner{}).Run(command, args...); err != nil {
		return fmt.Errorf("request privileged identity change: %w", err)
	}
	return nil
}

func (MonitorClient) Run() error {
	command, args := sys.PrivilegedCommand(BinaryPath, "monitor")
	if err := (systemd.ExecRunner{}).Run(command, args...); err != nil {
		return fmt.Errorf("open privileged Tor monitor: %w", err)
	}
	return nil
}
