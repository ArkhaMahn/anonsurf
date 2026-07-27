package tor

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testNyxrcManager(t *testing.T) NyxrcManager {
	t.Helper()

	return NyxrcManager{
		Path:            filepath.Join(t.TempDir(), "etc", "anonsurf", "nyxrc"),
		ExpectedContent: "tor_chroot\ndata_directory disabled\n",
	}
}

func TestEmbeddedNyxrcMatchesRepositoryConfig(t *testing.T) {
	repositoryNyxrc, err := os.ReadFile(filepath.Join("..", "..", "..", "configs", "nyxrc"))
	if err != nil {
		t.Fatal(err)
	}

	if strings.TrimRight(defaultNyxrc, "\n") != string(repositoryNyxrc) {
		t.Fatal("embedded nyxrc differs from configs/nyxrc")
	}
}

func TestNewNyxrcManagerUsesParrotPath(t *testing.T) {
	manager := NewNyxrcManager()

	if manager.Path != NyxrcPath {
		t.Fatalf("Path = %q, want %q", manager.Path, NyxrcPath)
	}
}

func TestNyxrcManagerEnsureWritesMissingConfig(t *testing.T) {
	manager := testNyxrcManager(t)

	if err := manager.Ensure(); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(manager.Path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "tor_chroot\ndata_directory disabled\n" {
		t.Fatalf("nyxrc = %q", string(content))
	}
}

func TestNyxrcManagerEnsureKeepsMatchingConfig(t *testing.T) {
	manager := testNyxrcManager(t)
	if err := os.MkdirAll(filepath.Dir(manager.Path), 0755); err != nil {
		t.Fatal(err)
	}
	original := "custom header\ntor_chroot\ndata_directory disabled\ncustom footer\n"
	if err := os.WriteFile(manager.Path, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	if err := manager.Ensure(); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(manager.Path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != original {
		t.Fatalf("nyxrc changed to %q", string(content))
	}
}

func TestNyxrcManagerEnsureReplacesOutdatedConfig(t *testing.T) {
	manager := testNyxrcManager(t)
	if err := os.MkdirAll(filepath.Dir(manager.Path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manager.Path, []byte("outdated config\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := manager.Ensure(); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(manager.Path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "tor_chroot\ndata_directory disabled\n" {
		t.Fatalf("nyxrc = %q", string(content))
	}
}

func TestNyxrcManagerEnsureRejectsEmptyExpectedContent(t *testing.T) {
	manager := testNyxrcManager(t)
	manager.ExpectedContent = " \n"

	err := manager.Ensure()
	if !errors.Is(err, ErrEmptyNyxrc) {
		t.Fatalf("error = %v, want %v", err, ErrEmptyNyxrc)
	}
}
