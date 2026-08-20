package tor

import (
	"os"
	"path/filepath"
	"runtime"
)

func repositoryConfigPath(name string) string {
	if root := os.Getenv("ANONSURF_SOURCE_ROOT"); root != "" {
		return filepath.Join(root, "configs", name)
	}

	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Join("..", "..", "..", "configs", name)
	}
	return filepath.Join(filepath.Dir(sourceFile), "..", "..", "..", "configs", name)
}
