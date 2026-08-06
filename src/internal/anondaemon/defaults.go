package anondaemon

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"gitlab.com/parrotsec/packages/anonsurf/internal/dns"
	"gitlab.com/parrotsec/packages/anonsurf/internal/systemd"
	"gitlab.com/parrotsec/packages/anonsurf/internal/tor"
)

const (
	MakeTorrcPath             = "/usr/lib/anonsurf/make-torrc"
	SysctlPath                = "/sbin/sysctl"
	IPTablesPath              = "/usr/sbin/iptables"
	IPTablesSavePath          = "/usr/sbin/iptables-save"
	IPTablesRulesPath         = "/etc/network/iptables.rules"
	TorControlCookie          = tor.DefaultCookieAuthFile
	defaultTorUser            = "debian-tor"
	defaultCookieTimeout      = time.Second
	defaultCookiePollInterval = 100 * time.Millisecond
)

func (c Controller) generateTorrc() error {
	if c.GenerateTorrc != nil {
		return c.GenerateTorrc()
	}
	return c.run(c.makeTorrcPath())
}

func (c Controller) run(command string, args ...string) error {
	return c.runner().Run(command, args...)
}

func (c Controller) runner() Runner {
	if c.Runner != nil {
		return c.Runner
	}
	return systemd.ExecRunner{}
}

func (c Controller) dnsManager() DNSManager {
	if c.DNS != nil {
		return c.DNS
	}
	return dns.NewManager()
}

func (c Controller) torService() TorService {
	if c.TorService != nil {
		return c.TorService
	}
	return systemd.NewTorController()
}

func (c Controller) makeTorrcPath() string {
	if c.MakeTorrcPath != "" {
		return c.MakeTorrcPath
	}
	return MakeTorrcPath
}

func (c Controller) sysctlPath() string {
	if c.SysctlPath != "" {
		return c.SysctlPath
	}
	return SysctlPath
}

func (c Controller) iptablesPath() string {
	if c.IPTablesPath != "" {
		return c.IPTablesPath
	}
	return IPTablesPath
}

func (c Controller) iptablesSavePath() string {
	if c.IPTablesSavePath != "" {
		return c.IPTablesSavePath
	}
	return IPTablesSavePath
}

func (c Controller) iptablesRulesPath() string {
	if c.IPTablesRulesPath != "" {
		return c.IPTablesRulesPath
	}
	return IPTablesRulesPath
}

func (c Controller) torrcPath() string {
	if c.TorrcPath != "" {
		return c.TorrcPath
	}
	return tor.TorrcPath
}

func (c Controller) torCookiePath() string {
	if c.TorCookiePath != "" {
		return c.TorCookiePath
	}
	return TorControlCookie
}

func (c Controller) torUser() string {
	if c.TorUser != "" {
		return c.TorUser
	}
	return defaultTorUser
}

func (c Controller) torExclude() []string {
	if len(c.TorExclude) > 0 {
		return c.TorExclude
	}
	return []string{"192.168.0.0/16", "172.16.0.0/12", "10.0.0.0/8", "127.0.0.0/9", "127.128.0.0/10"}
}

func (c Controller) cookieTimeout() time.Duration {
	if c.CookieTimeout != 0 {
		return c.CookieTimeout
	}
	return defaultCookieTimeout
}

func (c Controller) cookiePollInterval() time.Duration {
	if c.CookiePollInterval > 0 {
		return c.CookiePollInterval
	}
	return defaultCookiePollInterval
}

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

func (c Controller) sleep(duration time.Duration) {
	if c.Sleep != nil {
		c.Sleep(duration)
		return
	}
	time.Sleep(duration)
}
