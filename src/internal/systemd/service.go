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
	TorUnit       = "tor.service"
)

type Runner interface {
	Run(command string, args ...string) error
	Output(command string, args ...string) ([]byte, error)
}

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
	Runner        Runner
	SystemctlPath string
	Unit          string
	UseSudo       bool
	Sudo          sys.Sudo
}

func NewAnonsurfdController() ServiceController {
	return ServiceController{
		Runner:        ExecRunner{},
		SystemctlPath: SystemctlPath,
		Unit:          AnonsurfdUnit,
		UseSudo:       true,
	}
}

func NewTorController() ServiceController {
	return ServiceController{
		Runner:        ExecRunner{},
		SystemctlPath: SystemctlPath,
		Unit:          TorUnit,
	}
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
	runner := c.runner()
	systemctlPath := c.systemctlPath()
	unit := c.unit()

	output, err := runner.Output(systemctlPath, action, unit)
	state := strings.TrimSpace(string(output))
	if value, ok := knownStates[state]; ok {
		return value, nil
	}
	if err != nil {
		return false, fmt.Errorf("systemctl %s %s: %w", action, unit, err)
	}
	return false, fmt.Errorf("unexpected systemctl %s state %q", action, state)
}

func (c ServiceController) run(action string) error {
	runner := c.runner()
	systemctlPath := c.systemctlPath()
	unit := c.unit()
	command, args := c.runCommand(systemctlPath, action, unit)

	if err := runner.Run(command, args...); err != nil {
		return fmt.Errorf("systemctl %s %s: %w", action, unit, err)
	}
	return nil
}

func (c ServiceController) runCommand(systemctlPath, action, unit string) (string, []string) {
	args := []string{action, unit}
	if c.UseSudo {
		return c.Sudo.Command(systemctlPath, args...)
	}
	return systemctlPath, args
}

func (c ServiceController) runner() Runner {
	if c.Runner != nil {
		return c.Runner
	}
	return ExecRunner{}
}

func (c ServiceController) systemctlPath() string {
	if c.SystemctlPath != "" {
		return c.SystemctlPath
	}
	return SystemctlPath
}

func (c ServiceController) unit() string {
	if c.Unit != "" {
		return c.Unit
	}
	return AnonsurfdUnit
}
