package safekill

import (
	"fmt"
	"os/exec"
)

const Path = "/usr/lib/anonsurf/safekill"

type Runner struct {
	Path string
}

func NewRunner() Runner {
	return Runner{Path: Path}
}

func (r Runner) Run() error {
	path := r.Path
	if path == "" {
		path = Path
	}
	if err := exec.Command(path).Run(); err != nil {
		return fmt.Errorf("run safekill: %w", err)
	}
	return nil
}
