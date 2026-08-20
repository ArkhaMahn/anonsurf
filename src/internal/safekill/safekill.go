package safekill

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

const (
	BleachBitPath = "/usr/bin/bleachbit"
	procRootPath  = "/proc"
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

func killProcesses(names []string) error {
	nameTargets := make(map[string]bool)
	for _, target := range names {
		nameTargets[target] = true
	}

	entries, err := os.ReadDir(procRootPath)
	if err != nil {
		return fmt.Errorf("read process table: %w", err)
	}

	euid := os.Geteuid()
	var firstFailure error
	for _, entry := range entries {
		pid, err := strconv.Atoi(entry.Name())
		if err != nil || pid <= 0 || pid == os.Getpid() {
			continue
		}

		processDir := filepath.Join(procRootPath, entry.Name())
		info, err := os.Stat(processDir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			if firstFailure == nil {
				firstFailure = fmt.Errorf("inspect process %d: %w", pid, err)
			}
			continue
		}
		stat, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			if firstFailure == nil {
				firstFailure = fmt.Errorf("inspect process %d owner", pid)
			}
			continue
		}
		if euid != 0 && int(stat.Uid) != euid {
			continue
		}

		comm, err := os.ReadFile(filepath.Join(processDir, "comm"))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			if firstFailure == nil {
				firstFailure = fmt.Errorf("read process %d name: %w", pid, err)
			}
			continue
		}
		processName := strings.TrimSpace(string(comm))

		if !nameTargets[processName] {
			continue
		}

		if err := syscall.Kill(pid, syscall.SIGTERM); err != nil && !errors.Is(err, syscall.ESRCH) {
			if firstFailure == nil {
				firstFailure = fmt.Errorf("terminate process %d: %w", pid, err)
			}
		}
	}
	return firstFailure
}

type Runner struct{}

func (Runner) Run() error {
	if err := killProcesses(processNames); err != nil {
		return fmt.Errorf("kill networked applications: %w", err)
	}

	cleanArgs := append([]string{"-c"}, cacheTargets...)
	if err := exec.Command(BleachBitPath, cleanArgs...).Run(); err != nil {
		return fmt.Errorf("clear application caches: %w", err)
	}
	return nil
}
