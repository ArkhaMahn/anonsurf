package tor

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed assets/nyxrc
var defaultNyxrc string

// NyxrcPath follows the Parrot/Debian AnonSurf package layout.
const NyxrcPath = "/etc/anonsurf/nyxrc"

var ErrEmptyNyxrc = errors.New("nyxrc content is empty")

type NyxrcManager struct {
	Path            string
	ExpectedContent string
}

func NewNyxrcManager() NyxrcManager {
	return NyxrcManager{
		Path:            NyxrcPath,
		ExpectedContent: defaultNyxrc,
	}
}

// Ensure writes nyxrc only when missing or when the expected config is not present.
func (m NyxrcManager) Ensure() error {
	expected := strings.TrimRight(m.ExpectedContent, "\n")
	if strings.TrimSpace(expected) == "" {
		return ErrEmptyNyxrc
	}

	current, err := os.ReadFile(m.Path)
	if err == nil && strings.Contains(string(current), expected) {
		return nil
	}
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read nyxrc: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(m.Path), 0755); err != nil {
		return fmt.Errorf("create nyxrc directory: %w", err)
	}
	if err := writeFileAtomically(m.Path, []byte(expected+"\n"), 0644); err != nil {
		return fmt.Errorf("write nyxrc: %w", err)
	}
	return nil
}
