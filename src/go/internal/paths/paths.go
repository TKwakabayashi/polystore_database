// Package paths resolves the repository's relative default paths independently of the working
// directory. The defaults are written relative to src/go (where the module lives), so a command
// run from elsewhere does not create stray directories.
package paths

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	once  sync.Once
	goDir string
)

// GoDir returns the module directory (src/go): the nearest ancestor of the working directory, or
// of the executable, that contains go.mod with this module, or "" if none is found.
func GoDir() string {
	once.Do(func() {
		var starts []string
		if wd, err := os.Getwd(); err == nil {
			starts = append(starts, wd)
		}
		if exe, err := os.Executable(); err == nil {
			starts = append(starts, filepath.Dir(exe))
		}
		for _, s := range starts {
			for d := s; ; d = filepath.Dir(d) {
				if isModule(d) {
					goDir = d
					return
				}
				if sub := filepath.Join(d, "src", "go"); isModule(sub) {
					goDir = sub
					return
				}
				if filepath.Dir(d) == d {
					break
				}
			}
		}
	})
	return goDir
}

func isModule(dir string) bool {
	b, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	return err == nil && containsModule(string(b))
}

func containsModule(mod string) bool {
	return strings.Contains(mod, "module polystore_database/src/go")
}

// Resolve returns p as is when it is absolute or the module directory is unknown, and otherwise
// p taken relative to the module directory (src/go).
func Resolve(p string) string {
	if p == "" || filepath.IsAbs(p) {
		return p
	}
	if d := GoDir(); d != "" {
		return filepath.Clean(filepath.Join(d, p))
	}
	return p
}
