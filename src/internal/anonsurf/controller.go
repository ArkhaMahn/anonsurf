package anonsurf

import (
	"errors"
	"fmt"
)

var ErrTorrcNotFound = errors.New("torrc not found")

type Controller struct{}

func (c Controller) Restart() error {
	if err := c.Stop(); err != nil {
		return fmt.Errorf("stop before restart: %w", err)
	}
	if err := c.Start(); err != nil {
		return fmt.Errorf("start after restart: %w", err)
	}
	return nil
}
