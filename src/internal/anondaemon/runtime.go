package anondaemon

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func (c Controller) warningWriter() io.Writer {
	if c.WarningWriter != nil {
		return c.WarningWriter
	}
	return os.Stderr
}

func (c Controller) warnf(format string, args ...any) {
	fmt.Fprintf(c.warningWriter(), format+"\n", args...)
}

func (c Controller) readFile(path string) ([]byte, error) {
	if c.ReadFile != nil {
		return c.ReadFile(path)
	}
	return os.ReadFile(path)
}

func (c Controller) fileExists(path string) (bool, error) {
	if c.FileExists != nil {
		return c.FileExists(path)
	}
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (c Controller) writeFile(path string, data []byte, perm os.FileMode) error {
	if c.WriteFile != nil {
		return c.WriteFile(path, data, perm)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, perm)
}

func (c Controller) removeFile(path string) error {
	if c.RemoveFile != nil {
		return c.RemoveFile(path)
	}
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (c Controller) chmodReadable(path string) error {
	if c.ChmodReadable != nil {
		return c.ChmodReadable(path)
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	return os.Chmod(path, info.Mode().Perm()|0444)
}

func (c Controller) runWithInput(command string, input []byte, args ...string) error {
	if c.RunWithInput != nil {
		return c.RunWithInput(command, input, args...)
	}
	cmd := exec.Command(command, args...)
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (c Controller) sleep(duration time.Duration) {
	if c.Sleep != nil {
		c.Sleep(duration)
		return
	}
	time.Sleep(duration)
}
