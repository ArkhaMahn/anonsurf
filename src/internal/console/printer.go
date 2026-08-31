package console

import (
	"fmt"
	"io"
)

const (
	blueText   = "\x1b[38;2;0;124;166m"
	greenText  = "\x1b[32m"
	yellowText = "\x1b[33m"
	redText    = "\x1b[31m"
	resetText  = "\x1b[0m"
)

type Printer struct {
	stdout      io.Writer
	stderr      io.Writer
	stdoutColor bool
	stderrColor bool
}

func New(stdout, stderr io.Writer) Printer {
	return Printer{
		stdout:      stdout,
		stderr:      stderr,
		stdoutColor: SupportsColor(stdout),
		stderrColor: SupportsColor(stderr),
	}
}

func (p Printer) Result(message string) {
	fmt.Fprintln(p.stdout, message)
}

func (p Printer) Resultf(format string, args ...any) {
	fmt.Fprintf(p.stdout, format, args...)
}

func (p Printer) Running(message string) {
	p.writeStatus(p.stdout, p.stdoutColor, "....", blueText, message)
}

func (p Printer) Success(message string) {
	p.writeStatus(p.stdout, p.stdoutColor, " OK ", greenText, message)
}

func (p Printer) Warning(message string) {
	p.writeStatus(p.stdout, p.stdoutColor, "WARN", yellowText, message)
}

func (p Printer) Error(message string) {
	p.writeStatus(p.stderr, p.stderrColor, "FAIL", redText, message)
}

func (p Printer) Info(message string) {
	p.writeStatus(p.stderr, p.stderrColor, "INFO", blueText, message)
}

func (p Printer) writeStatus(output io.Writer, useColor bool, label, color, message string) {
	if useColor {
		fmt.Fprintf(output, "[%s%s%s] %s\n", color, label, resetText, message)
		return
	}
	fmt.Fprintf(output, "[%s] %s\n", label, message)
}
