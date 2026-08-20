package dns

import (
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
