package tor

import (
	"reflect"
	"strings"
	"testing"

	"gitlab.com/parrotsec/packages/anonsurf/internal/config"
)

func TestTorrcOptionsFromConfig(t *testing.T) {
	cfg := config.Config{
		UseSandbox:          true,
		UseSafeSocks:        true,
		RejectPlainPorts:    true,
		ListPlainPorts:      "23,80",
		AutoBridge:          true,
		BridgeAddress:       "obfs4 127.0.0.1:443 cert=abc",
		HTTPProxyAddress:    "127.0.0.1:8080",
		HTTPProxyCredential: "user:pass",
		ProxyIsHTTPS:        true,
	}

	got := TorrcOptionsFromConfig(cfg)
	want := TorrcOptions{
		UseSandbox:          true,
		UseSafeSocks:        true,
		RejectPlainPorts:    true,
		PlaintextPorts:      "23,80",
		AutoBridge:          true,
		BridgeAddress:       "obfs4 127.0.0.1:443 cert=abc",
		HTTPProxyAddress:    "127.0.0.1:8080",
		HTTPProxyCredential: "user:pass",
		ProxyIsHTTPS:        true,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("options = %#v, want %#v", got, want)
	}
}

func TestRenderTorrcWithWarnPlaintextPorts(t *testing.T) {
	got, err := RenderTorrcWithBase("Base\n", TorrcOptions{
		PlaintextPorts: "23,109,109, 110",
	})
	if err != nil {
		t.Fatal(err)
	}

	want := "Base\n\n# Use sandbox\nSandbox 0\n\n# Bridge setting: No bridge\n\n# Use SafeSocks\nSafeSocks 0\n\n# Set WarnPlaintextPorts\nWarnPlaintextPorts 23,109,110\n"
	if got != want {
		t.Fatalf("torrc = %q, want %q", got, want)
	}
}

func TestRenderTorrcWithRejectPlaintextPorts(t *testing.T) {
	got, err := RenderTorrcWithBase("Base", TorrcOptions{
		UseSandbox:       true,
		UseSafeSocks:     true,
		RejectPlainPorts: true,
		PlaintextPorts:   "23,80",
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, expected := range []string{
		"Sandbox 1",
		"SafeSocks 1",
		"RejectPlaintextPorts 23,80",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("torrc missing %q: %s", expected, got)
		}
	}
	if strings.Contains(got, "WarnPlaintextPorts") {
		t.Fatalf("torrc should not include WarnPlaintextPorts: %s", got)
	}
}

func TestRenderTorrcWithPlainBridge(t *testing.T) {
	got, err := RenderTorrcWithBase("Base", TorrcOptions{
		PlaintextPorts: "23",
		BridgeAddress:  "192.0.2.10:443 ABCDEF",
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, expected := range []string{
		"UseBridges 1",
		"Bridge 192.0.2.10:443 ABCDEF",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("torrc missing %q: %s", expected, got)
		}
	}
	if strings.Contains(got, "ClientTransportPlugin") {
		t.Fatalf("plain bridge should not include a pluggable transport: %s", got)
	}
}

func TestRenderTorrcWithObfs4Bridge(t *testing.T) {
	got, err := RenderTorrcWithBase("Base", TorrcOptions{
		PlaintextPorts: "23",
		BridgeAddress:  "obfs4 192.0.2.10:443 cert=abc",
		Obfs4ProxyPath: "/usr/bin/obfs4proxy",
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, expected := range []string{
		"UseBridges 1",
		"ClientTransportPlugin obfs4 exec /usr/bin/obfs4proxy managed",
		"Bridge obfs4 192.0.2.10:443 cert=abc",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("torrc missing %q: %s", expected, got)
		}
	}
}

func TestRenderTorrcWithSnowflakeBridge(t *testing.T) {
	got, err := RenderTorrcWithBase("Base", TorrcOptions{
		PlaintextPorts:      "23",
		BridgeAddress:       "snowflake 192.0.2.10:443",
		SnowflakeClientPath: "/usr/bin/snowflake-client",
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, expected := range []string{
		"UseBridges 1",
		"ClientTransportPlugin snowflake exec /usr/bin/snowflake-client",
		"Bridge snowflake 192.0.2.10:443",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("torrc missing %q: %s", expected, got)
		}
	}
}

func TestRenderTorrcWithHTTPProxy(t *testing.T) {
	got, err := RenderTorrcWithBase("Base", TorrcOptions{
		PlaintextPorts:   "23",
		HTTPProxyAddress: "127.0.0.1:8080",
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, expected := range []string{
		"# Set proxy address",
		"HTTPProxy 127.0.0.1:8080",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("torrc missing %q: %s", expected, got)
		}
	}
	if strings.Contains(got, "HTTPSProxy") {
		t.Fatalf("HTTP proxy should not include HTTPS proxy directives: %s", got)
	}
}

func TestRenderTorrcWithHTTPSProxyCredential(t *testing.T) {
	got, err := RenderTorrcWithBase("Base", TorrcOptions{
		PlaintextPorts:      "23",
		HTTPProxyAddress:    "127.0.0.1:8443",
		HTTPProxyCredential: "user:pass",
		ProxyIsHTTPS:        true,
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, expected := range []string{
		"# Set proxy address",
		"HTTPSProxy 127.0.0.1:8443",
		"HTTPSProxyAuthenticator user:pass",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("torrc missing %q: %s", expected, got)
		}
	}
}

func TestRenderTorrcRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		base    string
		options TorrcOptions
		want    string
	}{
		{
			name:    "empty base",
			base:    "",
			options: TorrcOptions{PlaintextPorts: "23"},
			want:    "torrc base is empty",
		},
		{
			name:    "empty port",
			base:    "Base",
			options: TorrcOptions{PlaintextPorts: "23,"},
			want:    "empty plaintext port",
		},
		{
			name:    "invalid port",
			base:    "Base",
			options: TorrcOptions{PlaintextPorts: "23,abc"},
			want:    `invalid plaintext port "abc"`,
		},
		{
			name:    "out of range port",
			base:    "Base",
			options: TorrcOptions{PlaintextPorts: "65536"},
			want:    `invalid plaintext port "65536"`,
		},
		{
			name:    "auto bridge without selected address",
			base:    "Base",
			options: TorrcOptions{PlaintextPorts: "23", AutoBridge: true},
			want:    "auto bridge requires a selected bridge address",
		},
		{
			name:    "obfs4 without transport binary",
			base:    "Base",
			options: TorrcOptions{PlaintextPorts: "23", BridgeAddress: "obfs4 192.0.2.10:443 cert=abc"},
			want:    "obfs4 bridge requires obfs4proxy path",
		},
		{
			name:    "snowflake without transport binary",
			base:    "Base",
			options: TorrcOptions{PlaintextPorts: "23", BridgeAddress: "snowflake 192.0.2.10:443"},
			want:    "snowflake bridge requires snowflake-client path",
		},
		{
			name:    "unsupported meek transport",
			base:    "Base",
			options: TorrcOptions{PlaintextPorts: "23", BridgeAddress: "meek 192.0.2.10:443"},
			want:    "meek bridges are not supported",
		},
		{
			name:    "proxy credential without address",
			base:    "Base",
			options: TorrcOptions{PlaintextPorts: "23", HTTPProxyCredential: "user:pass"},
			want:    "proxy credential requires proxy address",
		},
		{
			name:    "HTTP proxy with credentials",
			base:    "Base",
			options: TorrcOptions{PlaintextPorts: "23", HTTPProxyAddress: "127.0.0.1:8080", HTTPProxyCredential: "user:pass"},
			want:    "HTTP proxy credentials require proxy_is_https=true",
		},
		{
			name:    "proxy address with newline",
			base:    "Base",
			options: TorrcOptions{PlaintextPorts: "23", HTTPProxyAddress: "127.0.0.1:8080\nSocksPort 0"},
			want:    "proxy address must be a single line",
		},
		{
			name: "proxy credential with newline",
			base: "Base",
			options: TorrcOptions{
				PlaintextPorts:      "23",
				HTTPProxyAddress:    "127.0.0.1:8443",
				HTTPProxyCredential: "user:pass\nSocksPort 0",
				ProxyIsHTTPS:        true,
			},
			want: "proxy credential must be a single line",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := RenderTorrcWithBase(test.base, test.options)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want containing %q", err, test.want)
			}
		})
	}
}
