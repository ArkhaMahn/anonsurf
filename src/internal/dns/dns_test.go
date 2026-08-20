package dns

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func testManager(t *testing.T) Manager {
	t.Helper()

	dir := t.TempDir()
	return Manager{
		ResolvConfPath:         filepath.Join(dir, "resolv.conf"),
		ResolvConfBackup:       filepath.Join(dir, "anonsurf", "resolv.conf.bak"),
		NetworkManagerHookPath: filepath.Join(dir, "NetworkManager", "conf.d", "90-dns-none.conf"),
		ReloadNetworkManager: func() error {
			return nil
		},
		NetworkManagerDNSServers: func() ([]string, error) {
			return []string{"192.168.1.1", "invalid", "1.1.1.1"}, nil
		},
	}
}

func TestParseDNSAddresses(t *testing.T) {
	path := filepath.Join(t.TempDir(), "resolv.conf")
	content := "# comment\nnameserver 1.1.1.1\nsearch local\nnameserver 2001:4860:4860::8888 # google\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := ParseDNSAddresses(path)
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"1.1.1.1", "2001:4860:4860::8888"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("addresses = %#v, want %#v", got, want)
	}
}

func TestRenderResolvConfDeduplicatesValidAddresses(t *testing.T) {
	got, err := RenderResolvConf([]string{"1.1.1.1", "bad", "1.1.1.1", " 8.8.8.8 "})
	if err != nil {
		t.Fatal(err)
	}

	want := "# Written by DNSTool\nnameserver 1.1.1.1\nnameserver 8.8.8.8\n"
	if got != want {
		t.Fatalf("rendered resolv.conf = %q, want %q", got, want)
	}
}

func TestRenderResolvConfRejectsEmptyValidAddresses(t *testing.T) {
	_, err := RenderResolvConf([]string{"bad"})
	if !errors.Is(err, ErrNoValidDNSAddress) {
		t.Fatalf("error = %v, want %v", err, ErrNoValidDNSAddress)
	}
}

func TestParseAddrFromParamsReportsInvalidValues(t *testing.T) {
	got := ParseAddrFromParams([]string{"dhcp", "1.1.1.1", "bad", "", "2001:4860:4860::8888"})

	if !got.HasDHCPFlag {
		t.Fatal("expected DHCP flag")
	}
	if want := []string{"1.1.1.1", "2001:4860:4860::8888"}; !reflect.DeepEqual(got.ListAddr, want) {
		t.Fatalf("addresses = %#v, want %#v", got.ListAddr, want)
	}
	if want := []string{"bad"}; !reflect.DeepEqual(got.Invalid, want) {
		t.Fatalf("invalid addresses = %#v, want %#v", got.Invalid, want)
	}
}

func TestParseNetworkManagerDNSOutput(t *testing.T) {
	output := "192.168.1.1\n\n2001:4860:4860::8888,1.1.1.1 invalid 1.1.1.1\n"

	got := ParseNetworkManagerDNSOutput(output)
	want := []string{"192.168.1.1", "2001:4860:4860::8888", "1.1.1.1"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DNS servers = %#v, want %#v", got, want)
	}
}

func TestManagerCreateAddrsWritesDNSAndNetworkManagerHook(t *testing.T) {
	manager := testManager(t)

	if err := manager.CreateAddrs([]string{"1.1.1.1", "1.1.1.1"}); err != nil {
		t.Fatal(err)
	}

	resolvContent, err := os.ReadFile(manager.ResolvConfPath)
	if err != nil {
		t.Fatal(err)
	}
	if want := "# Written by DNSTool\nnameserver 1.1.1.1\n"; string(resolvContent) != want {
		t.Fatalf("resolv.conf = %q, want %q", string(resolvContent), want)
	}

	hookContent, err := os.ReadFile(manager.NetworkManagerHookPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(hookContent) != NMHookDontManageResolvConf {
		t.Fatalf("NetworkManager hook = %q, want %q", string(hookContent), NMHookDontManageResolvConf)
	}
}

func TestManagerCreateAddrsRejectsInvalidWithoutReplacingExistingDNS(t *testing.T) {
	manager := testManager(t)
	oldContent := []byte("nameserver 9.9.9.9\n")
	if err := os.WriteFile(manager.ResolvConfPath, oldContent, 0644); err != nil {
		t.Fatal(err)
	}

	err := manager.CreateAddrs([]string{"bad"})
	if !errors.Is(err, ErrNoValidDNSAddress) {
		t.Fatalf("error = %v, want %v", err, ErrNoValidDNSAddress)
	}

	got, err := os.ReadFile(manager.ResolvConfPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(oldContent) {
		t.Fatalf("resolv.conf changed to %q, want %q", string(got), string(oldContent))
	}
}

func TestManagerCreateDHCPOnlyRemovesResolverBeforeReload(t *testing.T) {
	manager := testManager(t)
	if err := os.WriteFile(manager.ResolvConfPath, []byte("nameserver 1.1.1.1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	manager.ReloadNetworkManager = func() error {
		if _, err := os.Lstat(manager.ResolvConfPath); !os.IsNotExist(err) {
			t.Fatalf("resolv.conf still exists during reload: %v", err)
		}
		hook, err := os.ReadFile(manager.NetworkManagerHookPath)
		if err != nil {
			t.Fatal(err)
		}
		if string(hook) != NMHookHandleResolvConf {
			t.Fatalf("NetworkManager hook = %q, want %q", string(hook), NMHookHandleResolvConf)
		}
		return nil
	}

	if err := manager.CreateDHCPOnly(); err != nil {
		t.Fatal(err)
	}
}

func TestManagerMakeBackupSkipsLocalhostAndCopiesValidDNS(t *testing.T) {
	manager := testManager(t)
	if err := os.WriteFile(manager.ResolvConfPath, []byte("nameserver 127.0.0.1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	err := manager.MakeBackup()
	if !errors.Is(err, ErrBackupSkippedLocalhost) {
		t.Fatalf("error = %v, want %v", err, ErrBackupSkippedLocalhost)
	}

	if err := os.WriteFile(manager.ResolvConfPath, []byte("nameserver 1.1.1.1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := manager.MakeBackup(); err != nil {
		t.Fatal(err)
	}

	backup, err := os.ReadFile(manager.ResolvConfBackup)
	if err != nil {
		t.Fatal(err)
	}
	if want := "nameserver 1.1.1.1\n"; string(backup) != want {
		t.Fatalf("backup = %q, want %q", string(backup), want)
	}
}
