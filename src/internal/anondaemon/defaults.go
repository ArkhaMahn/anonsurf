package anondaemon

import (
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
	IPTablesRestorePath       = "/usr/sbin/iptables-restore"
	IPTablesRulesPath         = "/etc/network/iptables.rules"
	TorControlCookie          = tor.DefaultCookieAuthFile
	KillallPath               = "killall"
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

func (c Controller) iptablesRestorePath() string {
	if c.IPTablesRestorePath != "" {
		return c.IPTablesRestorePath
	}
	return IPTablesRestorePath
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

func (c Controller) killallPath() string {
	if c.KillallPath != "" {
		return c.KillallPath
	}
	return KillallPath
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
