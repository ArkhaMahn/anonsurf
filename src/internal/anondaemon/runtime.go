package anondaemon

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
)

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
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	tempFile, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	if _, err := tempFile.Write(data); err != nil {
		_ = tempFile.Close()
		return err
	}
	if err := tempFile.Chmod(perm); err != nil {
		_ = tempFile.Close()
		return err
	}
	if err := tempFile.Sync(); err != nil {
		_ = tempFile.Close()
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}
	return os.Rename(tempPath, path)
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
