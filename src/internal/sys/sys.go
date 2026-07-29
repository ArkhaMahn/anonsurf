package sys

import (
	"os"
)

const SudoPath = "/usr/bin/sudo"

type Sudo struct {
	SudoPath string
	IsAdmin  func() bool
}

func IsAdmin() bool {
	return os.Geteuid() == 0
}

func (s Sudo) Command(command string, args ...string) (string, []string) {
	if s.isAdmin() {
		return command, args
	}
	return s.sudoPath(), append([]string{command}, args...)
}

func (s Sudo) sudoPath() string {
	if s.SudoPath != "" {
		return s.SudoPath
	}
	return SudoPath
}

func (s Sudo) isAdmin() bool {
	if s.IsAdmin != nil {
		return s.IsAdmin()
	}
	return IsAdmin()
}
