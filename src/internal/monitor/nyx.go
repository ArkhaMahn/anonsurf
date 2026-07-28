package monitor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

const (
	NyxPath       = "/usr/bin/nyx"
	NyxConfigPath = "/etc/anonsurf/nyxrc"
)

type Runner interface {
	Run(command string, args []string, stdin io.Reader, stdout, stderr io.Writer) error
}

type ExecRunner struct{}

func (ExecRunner) Run(command string, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	cmd := exec.Command(command, args...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

type NyxMonitor struct {
	Runner  Runner
	Command string
	Config  string
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
}

func NewNyxMonitor(stdin io.Reader, stdout, stderr io.Writer) NyxMonitor {
	return NyxMonitor{
		Runner:  ExecRunner{},
		Command: NyxPath,
		Config:  NyxConfigPath,
		Stdin:   stdin,
		Stdout:  stdout,
		Stderr:  stderr,
	}
}

func (m NyxMonitor) Run() error {
	runner := m.Runner
	if runner == nil {
		runner = ExecRunner{}
	}
	command := m.Command
	if command == "" {
		command = NyxPath
	}
	config := m.Config
	if config == "" {
		config = NyxConfigPath
	}
	stdin := m.Stdin
	if stdin == nil {
		stdin = os.Stdin
	}
	stdout := m.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := m.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	if err := runner.Run(command, []string{"--config", config}, stdin, stdout, stderr); err != nil {
		return fmt.Errorf("run nyx monitor: %w", err)
	}
	return nil
}
