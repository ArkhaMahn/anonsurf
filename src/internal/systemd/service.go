package systemd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"gitlab.com/parrotsec/packages/anonsurf/internal/sys"
)

const (
	SystemctlPath = "/usr/bin/systemctl"
	AnonsurfdUnit = "anonsurfd.service"
	TorUnit       = "tor@default.service"
)

type ExecRunner struct{}

func (ExecRunner) Run(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (ExecRunner) Output(command string, args ...string) ([]byte, error) {
	return exec.Command(command, args...).CombinedOutput()
}

type ServiceController struct {
	unit       string
	privileged bool
}

func NewAnonsurfdController() ServiceController {
	return ServiceController{
		unit:       AnonsurfdUnit,
		privileged: true,
	}
}

func NewTorController() ServiceController {
	return ServiceController{unit: TorUnit}
}

func IsSystemStopping() bool {
	output, _ := (ExecRunner{}).Output(SystemctlPath, "is-system-running")
	return strings.TrimSpace(string(output)) == "stopping"
}

func (c ServiceController) Start() error {
	return c.run("start")
}

func (c ServiceController) Stop() error {
	return c.run("stop")
}

func (c ServiceController) Restart() error {
	return c.run("restart")
}

func (c ServiceController) Enable() error {
	return c.run("enable")
}

func (c ServiceController) Disable() error {
	return c.run("disable")
}

func (c ServiceController) IsEnabled() (bool, error) {
	return c.queryState("is-enabled", map[string]bool{
		"enabled":   true,
		"disabled":  false,
		"masked":    false,
		"static":    false,
		"indirect":  false,
		"generated": false,
		"transient": false,
	})
}

func (c ServiceController) IsActive() (bool, error) {
	return c.queryState("is-active", map[string]bool{
		"active":       true,
		"inactive":     false,
		"failed":       false,
		"activating":   false,
		"deactivating": false,
		"reloading":    false,
	})
}

func (c ServiceController) queryState(action string, knownStates map[string]bool) (bool, error) {
	output, err := (ExecRunner{}).Output(SystemctlPath, action, c.unit)
	state := strings.TrimSpace(string(output))
	if value, ok := knownStates[state]; ok {
		return value, nil
	}
	if err != nil {
		return false, fmt.Errorf("systemctl %s %s: %w", action, c.unit, err)
	}
	return false, fmt.Errorf("unexpected systemctl %s state %q", action, state)
}

func (c ServiceController) run(action string) error {
	command := SystemctlPath
	args := []string{action, c.unit}
	if c.privileged {
		command, args = sys.PrivilegedCommand(command, args...)
	}
	if err := (ExecRunner{}).Run(command, args...); err != nil {
		return fmt.Errorf("systemctl %s %s: %w", action, c.unit, err)
	}
	return nil
}
