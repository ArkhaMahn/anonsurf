package systemd

import (
	"fmt"
	"os/exec"
)

const (
	SystemctlPath = "/usr/bin/systemctl"
	AnonsurfdUnit = "anonsurfd.service"
)

type Runner interface {
	Run(command string, args ...string) error
}

type ExecRunner struct{}

func (ExecRunner) Run(command string, args ...string) error {
	return exec.Command(command, args...).Run()
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

func (c ServiceController) run(action string) error {
	runner := c.Runner
	if runner == nil {
		runner = ExecRunner{}
	}
	systemctlPath := c.SystemctlPath
	if systemctlPath == "" {
		systemctlPath = SystemctlPath
	}
	unit := c.Unit
	if unit == "" {
		unit = AnonsurfdUnit
	}

	if err := runner.Run(systemctlPath, action, unit); err != nil {
		return fmt.Errorf("systemctl %s %s: %w", action, unit, err)
	}
	return nil
}
