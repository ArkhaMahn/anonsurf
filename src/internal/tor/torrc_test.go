package tor

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gitlab.com/parrotsec/packages/anonsurf/internal/config"
)

func TestEmbeddedTorrcBaseMatchesRepositoryConfig(t *testing.T) {
	repositoryBase, err := os.ReadFile(filepath.Join("..", "..", "..", "configs", "torrc.base"))
	if err != nil {
		t.Fatal(err)
	}

	if defaultTorrcBase != string(repositoryBase) {
		t.Fatal("embedded torrc.base differs from configs/torrc.base")
	}
}

func TestTorrcOptionsFromConfig(t *testing.T) {
	cfg := config.Config{
		UseSandbox:       true,
		UseSafeSocks:     true,
		RejectPlainPorts: true,
		ListPlainPorts:   "23,80",
	}

	got := TorrcOptionsFromConfig(cfg)
	want := TorrcOptions{
		UseSandbox:       true,
		UseSafeSocks:     true,
		RejectPlainPorts: true,
		PlaintextPorts:   "23,80",
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

	want := "Base\n\n# Use sandbox\nSandbox 0\n\n# Use SafeSocks\nSafeSocks 0\n\n# Set WarnPlaintextPorts\nWarnPlaintextPorts 23,109,110\n"
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
