package anondaemon

import (
	"errors"
	"io"
	"os"
	"time"

	"gitlab.com/parrotsec/packages/anonsurf/internal/dns"
	"gitlab.com/parrotsec/packages/anonsurf/internal/systemd"
)

var (
	ErrCommandNotImplemented = errors.New("daemon command is not implemented")
	ErrTorrcNotFound         = errors.New("torrc not found")
)

type Runner interface {
	Run(command string, args ...string) error
	Output(command string, args ...string) ([]byte, error)
}

type DNSManager interface {
	MakeBackup() error
	CreateAddrs(addresses []string) error
}

type TorService interface {
	Start() error
	Stop() error
	IsActive() (bool, error)
}

type Controller struct {
	Runner     Runner
	DNS        DNSManager
	TorService TorService

	MakeTorrcPath      string
	SysctlPath         string
	IPTablesPath       string
	IPTablesSavePath   string
	IPTablesRulesPath  string
	TorrcPath          string
	TorCookiePath      string
	TorUser            string
	TorExclude         []string
	CookieTimeout      time.Duration
	CookiePollInterval time.Duration
	WarningWriter      io.Writer

	GenerateTorrc func() error
	ReadFile      func(path string) ([]byte, error)
	FileExists    func(path string) (bool, error)
	WriteFile     func(path string, data []byte, perm os.FileMode) error
	ChmodReadable func(path string) error
	Sleep         func(duration time.Duration)
}

func NewController() Controller {
	return Controller{
		Runner:     systemd.ExecRunner{},
		DNS:        dns.NewManager(),
		TorService: systemd.NewTorController(),
	}
}

func (c Controller) Stop() error {
	return ErrCommandNotImplemented
}

func (c Controller) Restart() error {
	return ErrCommandNotImplemented
}
