package anondaemon

import (
	"errors"
	"fmt"
	"os"

	"gitlab.com/parrotsec/packages/anonsurf/internal/dns"
	"gitlab.com/parrotsec/packages/anonsurf/internal/systemd"
)

var ErrTorrcNotFound = errors.New("torrc not found")

type Runner interface {
	Run(command string, args ...string) error
	Output(command string, args ...string) ([]byte, error)
}

type DNSManager interface {
	MakeBackup() error
	CreateAddrs(addresses []string) error
	RestoreBackupOrDHCP() error
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

	MakeTorrcPath          string
	SysctlPath             string
	IPTablesPath           string
	IPTablesSavePath       string
	IPTablesRestorePath    string
	IPTablesRulesPath      string
	DNSStatePath           string
	IPv6StatePath          string
	TorrcPath              string
	TorrcBackupPath        string
	TorrcMissingMarkerPath string
	TorUser                string
	TorExclude             []string

	GenerateTorrc func() error
	ReadFile      func(path string) ([]byte, error)
	FileExists    func(path string) (bool, error)
	WriteFile     func(path string, data []byte, perm os.FileMode) error
	RemoveFile    func(path string) error
	RunWithInput  func(command string, input []byte, args ...string) error
}

func NewController() Controller {
	return Controller{
		Runner:     systemd.ExecRunner{},
		DNS:        dns.NewManager(),
		TorService: systemd.NewTorController(),
	}
}

func (c Controller) Restart() error {
	if err := c.Stop(); err != nil {
		return fmt.Errorf("stop before restart: %w", err)
	}
	if err := c.Start(); err != nil {
		return fmt.Errorf("start after restart: %w", err)
	}
	return nil
}
