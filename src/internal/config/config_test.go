package config

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDefaultMatchesParrotConfig(t *testing.T) {
	got := Default()
	want := Config{
		UseSandbox:          false,
		BridgeAddress:       "",
		UseSafeSocks:        false,
		AutoBridge:          false,
		RejectPlainPorts:    false,
		ListPlainPorts:      "23,109,110,143",
		HTTPProxyAddress:    "",
		HTTPProxyCredential: "",
		ProxyIsHTTPS:        false,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("default config = %#v, want %#v", got, want)
	}
}

func TestParseOverridesKnownValues(t *testing.T) {
	input := `
[Ignored]
use_sandbox=false

[AnonSurfConfig]
use_sandbox=true
bridge_address="obfs4 127.0.0.1:443 cert=abc"
use_safe_sock=true
auto_bridge=true
reject_plain_port=true
list_plain_ports="23,80,110"
http_proxy_address=127.0.0.1:8080 # local proxy
http_proxy_credential="user:pass"
proxy_is_https=true
`

	got, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	want := Config{
		UseSandbox:          true,
		BridgeAddress:       "obfs4 127.0.0.1:443 cert=abc",
		UseSafeSocks:        true,
		AutoBridge:          true,
		RejectPlainPorts:    true,
		ListPlainPorts:      "23,80,110",
		HTTPProxyAddress:    "127.0.0.1:8080",
		HTTPProxyCredential: "user:pass",
		ProxyIsHTTPS:        true,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parsed config = %#v, want %#v", got, want)
	}
}

func TestParseKeepsDefaultsForEmptyValues(t *testing.T) {
	input := `
[AnonSurfConfig]
use_sandbox=
list_plain_ports=""
`

	got, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	if got != Default() {
		t.Fatalf("config = %#v, want defaults %#v", got, Default())
	}
}

func TestParseRejectsUnknownAnonSurfKey(t *testing.T) {
	_, err := Parse(strings.NewReader("[AnonSurfConfig]\nunknown=true\n"))
	if err == nil || !strings.Contains(err.Error(), `unknown config key "unknown"`) {
		t.Fatalf("error = %v, want unknown key error", err)
	}
}

func TestParseRejectsInvalidBoolean(t *testing.T) {
	_, err := Parse(strings.NewReader("[AnonSurfConfig]\nuse_sandbox=maybe\n"))
	if err == nil || !strings.Contains(err.Error(), "invalid boolean for use_sandbox") {
		t.Fatalf("error = %v, want invalid boolean error", err)
	}
}

func TestParseRequiresAnonSurfSection(t *testing.T) {
	_, err := Parse(strings.NewReader("[Other]\nuse_sandbox=true\n"))
	if !errors.Is(err, ErrSectionNotFound) {
		t.Fatalf("error = %v, want %v", err, ErrSectionNotFound)
	}
}

func TestLoadReturnsDefaultWithMissingFileError(t *testing.T) {
	got, err := Load(filepath.Join(t.TempDir(), "missing.cfg"))
	if err == nil {
		t.Fatal("expected missing file error")
	}
	if got != Default() {
		t.Fatalf("config = %#v, want defaults %#v", got, Default())
	}
}

func TestLoadParsesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "anonsurf.cfg")
	if err := os.WriteFile(path, []byte("[AnonSurfConfig]\nauto_bridge=true\n"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !got.AutoBridge {
		t.Fatalf("AutoBridge = false, want true")
	}
}
