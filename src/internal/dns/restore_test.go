package dns

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestRestoreBackupRestoresNetworkManagerOwnership(t *testing.T) {
	manager := testManager(t)
	if err := os.MkdirAll(filepath.Dir(manager.ResolvConfBackup), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manager.ResolvConfPath, []byte("nameserver 127.0.0.1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manager.ResolvConfBackup, []byte("nameserver 1.1.1.1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	manager.ReloadNetworkManager = func() error {
		content, err := os.ReadFile(manager.ResolvConfPath)
		if err != nil {
			t.Fatal(err)
		}
		if string(content) != "nameserver 1.1.1.1\n" {
			t.Fatalf("resolv.conf = %q", string(content))
		}
		return nil
	}

	if err := manager.RestoreBackupOrDHCP(); err != nil {
		t.Fatal(err)
	}
	hook, err := os.ReadFile(manager.NetworkManagerHookPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(hook) != NMHookHandleResolvConf {
		t.Fatalf("NetworkManager hook = %q", string(hook))
	}
}

func TestRestoreWithoutBackupRecreatesDHCPResolver(t *testing.T) {
	manager := testManager(t)
	if err := os.WriteFile(manager.ResolvConfPath, []byte("nameserver 127.0.0.1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	manager.ReloadNetworkManager = func() error {
		if _, err := os.Stat(manager.ResolvConfPath); !os.IsNotExist(err) {
			t.Fatalf("resolv.conf still exists during reload: %v", err)
		}
		return nil
	}

	if err := manager.RestoreBackupOrDHCP(); err != nil {
		t.Fatal(err)
	}
}

func TestRestoreBackupRestoresOriginalNetworkManagerHook(t *testing.T) {
	manager := testManager(t)
	originalHook := []byte("#[main]\n#dns=none\n")
	if err := os.MkdirAll(filepath.Dir(manager.NetworkManagerHookPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manager.NetworkManagerHookPath, originalHook, 0640); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manager.ResolvConfPath, []byte("nameserver 9.9.9.9\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := manager.MakeBackup(); err != nil {
		t.Fatal(err)
	}
	if err := manager.CreateAddrs([]string{"127.0.0.1"}); err != nil {
		t.Fatal(err)
	}

	if err := manager.RestoreBackupOrDHCP(); err != nil {
		t.Fatal(err)
	}
	hook, err := os.ReadFile(manager.NetworkManagerHookPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(hook) != string(originalHook) {
		t.Fatalf("NetworkManager hook = %q, want %q", string(hook), string(originalHook))
	}
	info, err := os.Stat(manager.NetworkManagerHookPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0640 {
		t.Fatalf("NetworkManager hook mode = %v, want 0640", info.Mode().Perm())
	}
	for _, path := range []string{manager.NetworkManagerHookBackup, manager.NetworkManagerHookMissingMarker} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("NetworkManager state %s still exists: %v", path, err)
		}
	}
}

func TestRestoreRemovesHookThatWasOriginallyMissing(t *testing.T) {
	manager := testManager(t)
	if err := os.WriteFile(manager.ResolvConfPath, []byte("nameserver 9.9.9.9\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := manager.MakeBackup(); err != nil {
		t.Fatal(err)
	}
	if err := manager.CreateAddrs([]string{"127.0.0.1"}); err != nil {
		t.Fatal(err)
	}

	if err := manager.RestoreBackupOrDHCP(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(manager.NetworkManagerHookPath); !os.IsNotExist(err) {
		t.Fatalf("NetworkManager hook still exists: %v", err)
	}
}

func TestRestoreKeepsBackupsWhenNetworkManagerReloadFails(t *testing.T) {
	manager := testManager(t)
	originalHook := []byte("# original hook\n")
	if err := os.MkdirAll(filepath.Dir(manager.NetworkManagerHookPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manager.NetworkManagerHookPath, originalHook, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manager.ResolvConfPath, []byte("nameserver 9.9.9.9\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := manager.MakeBackup(); err != nil {
		t.Fatal(err)
	}
	if err := manager.CreateAddrs([]string{"127.0.0.1"}); err != nil {
		t.Fatal(err)
	}

	reloadErr := errors.New("reload failed")
	manager.ReloadNetworkManager = func() error {
		return reloadErr
	}
	err := manager.RestoreBackupOrDHCP()
	if !errors.Is(err, reloadErr) {
		t.Fatalf("restore error = %v, want %v", err, reloadErr)
	}
	for _, path := range []string{manager.ResolvConfBackup, manager.NetworkManagerHookBackup} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("backup %s was removed: %v", path, err)
		}
	}
	manager.ReloadNetworkManager = func() error {
		return nil
	}
	if err := manager.RestoreBackupOrDHCP(); err != nil {
		t.Fatal(err)
	}
	resolv, err := os.ReadFile(manager.ResolvConfPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(resolv) != "nameserver 9.9.9.9\n" {
		t.Fatalf("resolv.conf = %q", string(resolv))
	}
	hook, err := os.ReadFile(manager.NetworkManagerHookPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(hook) != string(originalHook) {
		t.Fatalf("NetworkManager hook = %q", string(hook))
	}
}
