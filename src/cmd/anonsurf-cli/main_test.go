package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootHelpWithNoArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer

	if err := execute(nil, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}

	output := stdout.String()
	for _, expected := range []string{
		"AnonSurf manages Parrot OS system-wide Tor routing",
		"Commands:",
		"start",
		"enable-boot",
		"Examples:",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("help output missing %q:\n%s", expected, output)
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRootHelpFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer

	if err := execute([]string{"--help"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(stdout.String(), "status") {
		t.Fatalf("help output missing status command:\n%s", stdout.String())
	}
}

func TestVersionFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer

	if err := execute([]string{"--version"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}

	if got := stdout.String(); got != "AnonSurf 5.0.0\n" {
		t.Fatalf("version output = %q", got)
	}
}

func TestRegisteredCommandsReturnNotImplemented(t *testing.T) {
	commands := []string{
		"start",
		"stop",
		"restart",
		"status",
		"changeid",
		"myip",
		"enable-boot",
		"disable-boot",
		"status-boot",
	}

	for _, command := range commands {
		t.Run(command, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := execute([]string{command}, &stdout, &stderr)
			if err == nil {
				t.Fatal("expected not implemented error")
			}
			if !strings.Contains(err.Error(), "not implemented yet") {
				t.Fatalf("error = %v, want not implemented message", err)
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
		})
	}
}

func TestCommandAliasesReturnNotImplemented(t *testing.T) {
	tests := map[string]string{
		"newnym": "changeid",
		"ip":     "myip",
	}

	for alias, command := range tests {
		t.Run(alias, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := execute([]string{alias}, &stdout, &stderr)
			if err == nil {
				t.Fatal("expected not implemented error")
			}
			if !strings.Contains(err.Error(), command) {
				t.Fatalf("error = %v, want command %q", err, command)
			}
		})
	}
}

func TestCommandHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	if err := execute([]string{"start", "--help"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}

	output := stdout.String()
	for _, expected := range []string{
		"Start system-wide Tor transparent routing",
		"dedicated migration step",
		"Usage:",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("command help missing %q:\n%s", expected, output)
		}
	}
}

func TestUnknownCommandReturnsError(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := execute([]string{"unknown"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected unknown command error")
	}
	if !strings.Contains(err.Error(), `unknown command "unknown"`) {
		t.Fatalf("error = %v", err)
	}
}
