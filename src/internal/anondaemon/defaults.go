package anondaemon

import (
	"gitlab.com/parrotsec/packages/anonsurf/internal/systemd"
	"gitlab.com/parrotsec/packages/anonsurf/internal/tor"
)

const (
	MakeTorrcPath       = "/usr/lib/anonsurf/make-torrc"
	SysctlPath          = "/sbin/sysctl"
	IPTablesSavePath    = "/usr/sbin/iptables-save"
	IPTablesRestorePath = "/usr/sbin/iptables-restore"
	IPTablesRulesPath   = "/var/lib/anonsurf/iptables.rules"
	ActiveStatePath     = "/var/lib/anonsurf/active"
	DNSStatePath        = "/var/lib/anonsurf/dns.changed"
	IPv6StatePath       = "/var/lib/anonsurf/ipv6.state"
	TorStatePath        = "/var/lib/anonsurf/tor.was-active"
	IPv6ConfPath        = "/proc/sys/net/ipv6/conf"
	TorrcPath           = tor.TorrcPath
	TorrcBackupPath     = tor.TorrcBackupPath
	TorrcMissingPath    = tor.TorrcMissingMarkerPath
	defaultTorUser      = "debian-tor"
)

func (c Controller) generateTorrc() error {
	return c.run(MakeTorrcPath)
}

func (c Controller) run(command string, args ...string) error {
	return (systemd.ExecRunner{}).Run(command, args...)
}

var torExcludedNetworks = [...]string{
	"192.168.0.0/16",
	"172.16.0.0/12",
	"10.0.0.0/8",
	"127.0.0.0/8",
}
