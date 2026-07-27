package systemd

import (
	"fmt"
	"os/exec"
	"strings"
)

const (
	SystemctlPath = "/usr/bin/systemctl"
	AnonsurfdUnit = "anonsurfd.service"
)

type Runner interface {
	Run(command string, args ...string) error
	Output(command string, args ...string) ([]byte, error)
}

type ExecRunner struct{}

func (ExecRunner) Run(command string, args ...string) error {
	return exec.Command(command, args...).Run()
}

func (ExecRunner) Output(command string, args ...string) ([]byte, error) {
	return exec.Command(command, args...).CombinedOutput()
}

type ServiceController struct {
	Runner        Runner
	SystemctlPath string
	Unit          string
}

func NewAnonsurfdController() ServiceController {
	return ServiceController{
		Runner:        ExecRunner{},
		SystemctlPath: SystemctlPath,
		Unit:          AnonsurfdUnit,
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
	runner := c.runner()
	systemctlPath := c.systemctlPath()
	unit := c.unit()

	output, err := runner.Output(systemctlPath, "is-enabled", unit)
	state := strings.TrimSpace(string(output))
	switch state {
	case "enabled":
		return true, nil
	case "disabled", "masked", "static", "indirect", "generated", "transient":
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("systemctl is-enabled %s: %w", unit, err)
	}
	return false, fmt.Errorf("unexpected systemctl is-enabled state %q", state)
}

func (c ServiceController) run(action string) error {
	runner := c.runner()
	systemctlPath := c.systemctlPath()
	unit := c.unit()

	if err := runner.Run(systemctlPath, action, unit); err != nil {
		return fmt.Errorf("systemctl %s %s: %w", action, unit, err)
	}
	return nil
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
