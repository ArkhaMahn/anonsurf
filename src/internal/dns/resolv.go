package dns

import (
	"os"

	"gitlab.com/parrotsec/packages/anonsurf/internal/ui"
)

const (
	ResolvConfBackup = "/etc/anonsurf/resolv.conf.bak"
	ResolvConfPath   = "/etc/resolv.conf"
)

func tryRemoveFile(path string) bool {
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return false
	}
	return true
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func moveFile(src, dst string) error {
	return os.Rename(src, dst)
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func ResolvCreateDHCPOnly() {
	if !tryRemoveFile(ResolvConfPath) {
		ui.PrintError("Failed to remove current DNS settings")
		return
	}
	NMCreateHookDHCP()
}

func ResolvCreateAddrs(listAddr []string) {
	if len(listAddr) == 0 {
		ui.PrintError("No valid addresses is provided")
		return
	}

	if !tryRemoveFile(ResolvConfPath) {
		ui.PrintError("Failed to remove " + ResolvConfPath + " to create new one.")
		return
	}

	WriteDNSToSystem(listAddr, ResolvConfPath)
	NMCreateHookUnmanage()
}

func ResolvRestoreFromBackup() {
	if !tryRemoveFile(ResolvConfPath) {
		ui.PrintError("Restore backup: Failed to remove " + ResolvConfPath + " to restore backup.")
	}

	if fileExists(ResolvConfBackup) {
		err := moveFile(ResolvConfBackup, ResolvConfPath)
		if err != nil {
			ui.PrintError("Restore backup: Failed to restore backup")
		}
	}
}

func ResolvMakeBackup() {
	if !fileExists(ResolvConfPath) {
		ui.PrintError("Create backup: Skip creating backup file. Missing " + ResolvConfPath)
		return
	}

	if SystemHasOnlyLocalhost(ResolvConfPath) {
		ui.PrintError("Create backup: Skip creating backup file. Only found localhost nameserver.")
		return
	}

	if fileExists(ResolvConfBackup) {
		if !tryRemoveFile(ResolvConfBackup) {
			ui.PrintError("Create backup: Failed to remove " + ResolvConfBackup + " to create new one.")
			return
		}
	}

	if SystemDNSFileIsSymlink(ResolvConfPath) {
		ui.PrintError("Create backup: Skip creating backup file because " + ResolvConfPath + " is a symlink.")
		return
	}

	err := copyFile(ResolvConfPath, ResolvConfBackup)
	if err != nil {
		ui.PrintError("Create backup: Failed to create backup file")
	}
}
