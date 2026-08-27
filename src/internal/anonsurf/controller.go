package anonsurf

import (
	"errors"
)

var ErrTorrcNotFound = errors.New("torrc not found")

type Controller struct{}
