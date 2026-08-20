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

type CommandRunner interface {
	Run(command string, args ...string) error
}

type ProcessKiller interface {
	Kill(names []string) error
}

type ExecRunner struct{}

func (ExecRunner) Run(command string, args ...string) error {
	return exec.Command(command, args...).Run()
}

type ProcProcessKiller struct {
	ProcRoot string
	EUID     func() int
	Signal   func(pid int) error
}

func NewProcProcessKiller() ProcProcessKiller {
	return ProcProcessKiller{
		ProcRoot: procRootPath,
		EUID:     os.Geteuid,
		Signal: func(pid int) error {
			return syscall.Kill(pid, syscall.SIGTERM)
		},
	}
}

func (k ProcProcessKiller) Kill(names []string) error {
	procRoot := k.ProcRoot
	if procRoot == "" {
		procRoot = procRootPath
	}
	euid := os.Geteuid()
	if k.EUID != nil {
		euid = k.EUID()
	}
	signal := k.Signal
	if signal == nil {
		signal = func(pid int) error {
			return syscall.Kill(pid, syscall.SIGTERM)
		}
	}

	nameTargets := make(map[string]bool)
	pathTargets := make(map[string]bool)
	for _, target := range names {
		if strings.ContainsRune(target, filepath.Separator) {
			pathTargets[target] = true
		} else {
			nameTargets[target] = true
		}
	}

	entries, err := os.ReadDir(procRoot)
	if err != nil {
		return fmt.Errorf("read process table: %w", err)
	}

	var firstFailure error
	for _, entry := range entries {
		pid, err := strconv.Atoi(entry.Name())
		if err != nil || pid <= 0 || pid == os.Getpid() {
			continue
		}

		processDir := filepath.Join(procRoot, entry.Name())
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

		executable := ""
		if len(pathTargets) > 0 {
			executable, err = os.Readlink(filepath.Join(processDir, "exe"))
			if err != nil && !os.IsNotExist(err) {
				if firstFailure == nil {
					firstFailure = fmt.Errorf("read process %d executable: %w", pid, err)
				}
				continue
			}
			executable = strings.TrimSuffix(executable, " (deleted)")
		}
		if !nameTargets[processName] && !pathTargets[executable] {
			continue
		}

		if err := signal(pid); err != nil && !errors.Is(err, syscall.ESRCH) {
			if firstFailure == nil {
				firstFailure = fmt.Errorf("terminate process %d: %w", pid, err)
			}
		}
	}
	return firstFailure
}

type Runner struct {
	Executor      CommandRunner
	Processes     ProcessKiller
	BleachBitPath string
}

func NewRunner() Runner {
	return Runner{
		Executor:      ExecRunner{},
		Processes:     NewProcProcessKiller(),
		BleachBitPath: BleachBitPath,
	}
}

func (r Runner) Run() error {
	processes := r.Processes
	if processes == nil {
		processes = NewProcProcessKiller()
	}
	if err := processes.Kill(processNames); err != nil {
		return fmt.Errorf("kill networked applications: %w", err)
	}

	executor := r.Executor
	if executor == nil {
		executor = ExecRunner{}
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
