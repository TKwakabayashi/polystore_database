package paths

import (
	"path/filepath"
	"testing"
)

func TestResolve(t *testing.T) {
	d := GoDir()
	if d == "" || filepath.Base(d) != "go" {
		t.Fatalf("GoDir = %q", d)
	}
	if got, want := Resolve("../../results"), filepath.Join(filepath.Dir(filepath.Dir(d)), "results"); got != want {
		t.Errorf("Resolve = %q, want %q", got, want)
	}
	if Resolve("/abs") != "/abs" {
		t.Error("absolute paths must be kept")
	}
}
