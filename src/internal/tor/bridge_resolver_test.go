package tor

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseBridgeEntriesIgnoresCommentsAndEmptyLines(t *testing.T) {
	input := `
# default bridges

obfs4 192.0.2.1:443 cert=one
  snowflake 192.0.2.2:443
`

	got, err := ParseBridgeEntries(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		"obfs4 192.0.2.1:443 cert=one",
		"snowflake 192.0.2.2:443",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("entries = %#v, want %#v", got, want)
	}
}

func TestParseBridgeEntriesRejectsEmptyList(t *testing.T) {
	_, err := ParseBridgeEntries(strings.NewReader("# comments only\n\n"))
	if !errors.Is(err, ErrNoBridgeEntries) {
		t.Fatalf("error = %v, want %v", err, ErrNoBridgeEntries)
	}
}

func TestLoadBridgeEntriesReadsFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bridges.txt")
	if err := os.WriteFile(path, []byte("obfs4 192.0.2.1:443 cert=one\n"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := LoadBridgeEntries(path)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"obfs4 192.0.2.1:443 cert=one"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("entries = %#v, want %#v", got, want)
	}
}

func TestBridgeResolverLeavesDisabledBridgeUntouched(t *testing.T) {
	resolver := BridgeResolver{
		LookupExecutable: func(name string) (string, error) {
			t.Fatalf("lookup should not be called for disabled bridge")
			return "", nil
		},
		SelectBridge: func(entries []string) (string, error) {
			t.Fatalf("selector should not be called for disabled bridge")
			return "", nil
		},
	}
	options := TorrcOptions{PlaintextPorts: "23"}

	got, err := resolver.Resolve(options)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, options) {
		t.Fatalf("options = %#v, want %#v", got, options)
	}
}

func TestBridgeResolverResolvesManualObfs4Transport(t *testing.T) {
	resolver := BridgeResolver{
		LookupExecutable: func(name string) (string, error) {
			if name != "obfs4proxy" {
				t.Fatalf("lookup name = %q, want obfs4proxy", name)
			}
			return "/usr/bin/obfs4proxy", nil
		},
	}

	got, err := resolver.Resolve(TorrcOptions{
		BridgeAddress: "obfs4 192.0.2.1:443 cert=one",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Obfs4ProxyPath != "/usr/bin/obfs4proxy" {
		t.Fatalf("Obfs4ProxyPath = %q", got.Obfs4ProxyPath)
	}
}

func TestBridgeResolverKeepsProvidedTransportPath(t *testing.T) {
	resolver := BridgeResolver{
		LookupExecutable: func(name string) (string, error) {
			t.Fatalf("lookup should not be called when transport path is provided")
			return "", nil
		},
	}

	got, err := resolver.Resolve(TorrcOptions{
		BridgeAddress:  "obfs4 192.0.2.1:443 cert=one",
		Obfs4ProxyPath: "/custom/obfs4proxy",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Obfs4ProxyPath != "/custom/obfs4proxy" {
		t.Fatalf("Obfs4ProxyPath = %q", got.Obfs4ProxyPath)
	}
}

func TestBridgeResolverResolvesManualSnowflakeTransport(t *testing.T) {
	resolver := BridgeResolver{
		LookupExecutable: func(name string) (string, error) {
			if name != "snowflake-client" {
				t.Fatalf("lookup name = %q, want snowflake-client", name)
			}
			return "/usr/bin/snowflake-client", nil
		},
	}

	got, err := resolver.Resolve(TorrcOptions{
		BridgeAddress: "snowflake 192.0.2.1:443",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.SnowflakeClientPath != "/usr/bin/snowflake-client" {
		t.Fatalf("SnowflakeClientPath = %q", got.SnowflakeClientPath)
	}
}

func TestBridgeResolverSelectsAutoBridge(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bridges.txt")
	if err := os.WriteFile(path, []byte("obfs4 192.0.2.1:443 cert=one\nsnowflake 192.0.2.2:443\n"), 0644); err != nil {
		t.Fatal(err)
	}
	resolver := BridgeResolver{
		BridgeListPath: path,
		LookupExecutable: func(name string) (string, error) {
			if name != "snowflake-client" {
				t.Fatalf("lookup name = %q, want snowflake-client", name)
			}
			return "/usr/bin/snowflake-client", nil
		},
		SelectBridge: func(entries []string) (string, error) {
			if len(entries) != 2 {
				t.Fatalf("entries = %#v, want 2 entries", entries)
			}
			return entries[1], nil
		},
	}

	got, err := resolver.Resolve(TorrcOptions{
		AutoBridge:    true,
		BridgeAddress: "obfs4 manual.example:443 cert=manual",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.BridgeAddress != "snowflake 192.0.2.2:443" {
		t.Fatalf("BridgeAddress = %q", got.BridgeAddress)
	}
	if got.SnowflakeClientPath != "/usr/bin/snowflake-client" {
		t.Fatalf("SnowflakeClientPath = %q", got.SnowflakeClientPath)
	}
}

func TestBridgeResolverReturnsLookupError(t *testing.T) {
	wantErr := errors.New("missing binary")
	resolver := BridgeResolver{
		LookupExecutable: func(name string) (string, error) {
			return "", wantErr
		},
	}

	_, err := resolver.Resolve(TorrcOptions{BridgeAddress: "obfs4 192.0.2.1:443 cert=one"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want wrapping %v", err, wantErr)
	}
	if !strings.Contains(err.Error(), "find obfs4proxy") {
		t.Fatalf("error = %v, want lookup context", err)
	}
}

func TestBridgeResolverRejectsEmptyLookupPath(t *testing.T) {
	resolver := BridgeResolver{
		LookupExecutable: func(name string) (string, error) {
			return " ", nil
		},
	}

	_, err := resolver.Resolve(TorrcOptions{BridgeAddress: "obfs4 192.0.2.1:443 cert=one"})
	if err == nil || !strings.Contains(err.Error(), "obfs4proxy path is empty") {
		t.Fatalf("error = %v, want empty path error", err)
	}
}

func TestBridgeResolverReturnsSelectorError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bridges.txt")
	if err := os.WriteFile(path, []byte("obfs4 192.0.2.1:443 cert=one\n"), 0644); err != nil {
		t.Fatal(err)
	}
	wantErr := fmt.Errorf("selector failed")
	resolver := BridgeResolver{
		BridgeListPath: path,
		SelectBridge: func(entries []string) (string, error) {
			return "", wantErr
		},
	}

	_, err := resolver.Resolve(TorrcOptions{AutoBridge: true})
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want wrapping %v", err, wantErr)
	}
}
