package safekill

import (
	"errors"
	"fmt"
	"os/exec"
)

const (
	KillallPath   = "/usr/bin/killall"
	BleachBitPath = "/usr/bin/bleachbit"
)

var processNames = []string{
	"chrome",
	"dropbox",
	"skype",
	"icedove",
	"thunderbird",
	"firefox",
	"firefox-esr",
	"chromium",
	"xchat",
	"hexchat",
	"transmission",
	"steam",
	"firejail",
	"/usr/lib/firefox-esr/firefox-esr",
}

var cacheTargets = []string{
	"adobe_reader.cache",
	"chromium.cache",
	"chromium.session",
	"chromium.history",
	"chromium.form_history",
	"elinks.history",
	"emesene.cache",
	"epiphany.cache",
	"firefox.cache",
	"firefox.crash_reports",
	"firefox.url_history",
	"firefox.forms",
	"flash.cache",
	"flash.cookies",
	"google_chrome.cache",
	"google_chrome.history",
	"google_chrome.form_history",
	"google_chrome.search_engines",
	"google_chrome.session",
	"google_earth.temporary_files",
	"links2.history",
	"opera.cache",
	"opera.form_history",
	"opera.history",
}

type CommandRunner interface {
	Run(command string, args ...string) error
}

type ExecRunner struct{}

func (ExecRunner) Run(command string, args ...string) error {
	return exec.Command(command, args...).Run()
}

type Runner struct {
	Executor      CommandRunner
	KillallPath   string
	BleachBitPath string
}

func NewRunner() Runner {
	return Runner{
		Executor:      ExecRunner{},
		KillallPath:   KillallPath,
		BleachBitPath: BleachBitPath,
	}
}

func (r Runner) Run() error {
	executor := r.Executor
	if executor == nil {
		executor = ExecRunner{}
	}

	killallPath := r.KillallPath
	if killallPath == "" {
		killallPath = KillallPath
	}
	killArgs := append([]string{"-q"}, processNames...)
	if err := executor.Run(killallPath, killArgs...); err != nil {
		var exitError *exec.ExitError
		// killall returns non-zero when any listed process is not running.
		if !errors.As(err, &exitError) {
			return fmt.Errorf("kill networked applications: %w", err)
		}
	}

	bleachBitPath := r.BleachBitPath
	if bleachBitPath == "" {
		bleachBitPath = BleachBitPath
	}
	cleanArgs := append([]string{"-c"}, cacheTargets...)
	if err := executor.Run(bleachBitPath, cleanArgs...); err != nil {
		return fmt.Errorf("clear application caches: %w", err)
	}
	return nil
}
