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
	stdout io.Writer
	stderr io.Writer
	color  bool
}

func New(stdout, stderr io.Writer) Printer {
	return Printer{
		stdout: stdout,
		stderr: stderr,
		color:  SupportsColor(stderr),
	}
}

func (p Printer) Result(message string) {
	fmt.Fprintln(p.stdout, message)
}

func (p Printer) Resultf(format string, args ...any) {
	fmt.Fprintf(p.stdout, format, args...)
}

func (p Printer) Running(message string) {
	p.writeStatus("....", blueText, message)
}

func (p Printer) Success(message string) {
	p.writeStatus(" OK ", greenText, message)
}

func (p Printer) Warning(message string) {
	p.writeStatus("WARN", yellowText, message)
}

func (p Printer) Error(message string) {
	p.writeStatus("FAIL", redText, message)
}

func (p Printer) Info(message string) {
	p.writeStatus("INFO", blueText, message)
}

func (p Printer) writeStatus(label, color, message string) {
	if p.color {
		fmt.Fprintf(p.stderr, "[%s%s%s] %s\n", color, label, resetText, message)
		return
	}
	fmt.Fprintf(p.stderr, "[%s] %s\n", label, message)
}
