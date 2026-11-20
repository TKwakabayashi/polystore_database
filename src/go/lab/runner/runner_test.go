package runner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"polystore_database/src/go/query/physical"
)

func TestCanonRows(t *testing.T) {
	a := []map[string]any{{"x": 1, "y": "a"}, {"x": 2, "y": "b"}}
	b := []map[string]any{{"y": "b", "x": 2}, {"y": "a", "x": 1}}
	if canonRows(a) != canonRows(b) {
		t.Error("row order must not matter")
	}
	dup := []map[string]any{{"x": 1, "y": "a"}, {"x": 1, "y": "a"}}
	if canonRows(a) == canonRows(dup) {
		t.Error("multiplicity must matter")
	}
}

func TestCanonRowsFloat(t *testing.T) {
	a := []map[string]any{{"avg": 32.60430442511862}}
	b := []map[string]any{{"avg": 32.60430442511958}}
	if canonRows(a) != canonRows(b) {
		t.Error("rounding noise must compare equal")
	}
	c := []map[string]any{{"avg": 32.6044}}
	if canonRows(a) == canonRows(c) {
		t.Error("a real difference must not compare equal")
	}
}

func TestCanonRowsTime(t *testing.T) {
	at := time.Date(2012, 11, 1, 9, 0, 0, 0, time.FixedZone("x", 9*3600))
	if canonRows([]map[string]any{{"t": at}}) != canonRows([]map[string]any{{"t": at.UTC()}}) {
		t.Error("the same instant must compare equal across zones")
	}
}

// A Neo4j Date equals the executor's date (UTC midnight).
func TestCanonRowsDate(t *testing.T) {
	d := time.Date(1985, 11, 15, 0, 0, 0, 0, time.UTC)
	if canonRows([]map[string]any{{"d": neo4j.DateOf(d)}}) != canonRows([]map[string]any{{"d": d}}) {
		t.Error("a Date and its UTC midnight must compare equal")
	}
}

func TestRunStatus(t *testing.T) {
	for want, err := range map[string]error{
		"ok":      nil,
		"timeout": fmt.Errorf("wrapped: %w", context.DeadlineExceeded),
		"error":   errors.New("boom"),
	} {
		if got := runStatus(err); got != want {
			t.Errorf("%v: got %s, want %s", err, got, want)
		}
	}
}

func TestMergeParams(t *testing.T) {
	got := mergeParams(map[string]string{"a": "1", "b": "2"}, map[string]string{"b": "3"})
	if got["a"] != "1" || got["b"] != "3" {
		t.Errorf("%v", got)
	}
}

func TestToValuedParamsDate(t *testing.T) {
	if _, ok := toValuedParams(map[string]string{"d": "1985-01-01"})["d"].(neo4j.Date); !ok {
		t.Error("a date-only string must become neo4j.Date")
	}
}

// Integers in a list stay integers (also above 2^53); strings and floats stay as they are.
func TestToValuedParamsList(t *testing.T) {
	got, ok := toValuedParams(map[string]string{"ids": `[14, 9007199254740993, 1.5, "a"]`})["ids"].([]any)
	if !ok || len(got) != 4 {
		t.Fatalf("not a list: %#v", got)
	}
	if got[0] != int64(14) || got[1] != int64(9007199254740993) || got[2] != 1.5 || got[3] != "a" {
		t.Errorf("values differ: %#v", got)
	}
}

// Every spec under lab/specs loads.
func TestLoadSpecFiles(t *testing.T) {
	files, _ := filepath.Glob("../specs/*.json")
	if len(files) == 0 {
		t.Fatal("no specs")
	}
	for _, f := range files {
		s, err := LoadSpec(f)
		if err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		if *s.Trials < 1 || len(s.Tunings) == 0 || len(s.Cases) == 0 || len(s.Queries) == 0 {
			t.Errorf("%s: defaults not filled: %+v", f, s)
		}
	}
}

func TestLoadSpecErrors(t *testing.T) {
	for name, body := range map[string]string{
		"no_name":       `{"workloads":["IS4"]}`,
		"bad_mode":      `{"name":"x","mode":"fast","workloads":["IS4"]}`,
		"bad_workload":  `{"name":"x","workloads":["NOPE"]}`,
		"bad_placement": `{"name":"x","workloads":["IS4"],"placements":["oracle"]}`,
		"bad_model":     `{"name":"x","workloads":["IS4"],"models":["x"]}`,
		"bad_timeout":   `{"name":"x","workloads":["IS4"],"timeout":"z"}`,
		"unknown_field": `{"name":"x","workloads":["IS4"],"nope":1}`,
		"dup_case":      `{"name":"x","workloads":["IS4"],"cases":[{"label":"a"},{"label":"a"}]}`,
		"bad_query":     `{"name":"x","workloads":["IS4"],"queries":["latest"]}`,
		"bad_migrate":   `{"name":"x","workloads":["IS4"],"migrate":[{"kind":"x","entity":"Person","props":["id"]}]}`,
		"no_official":   `{"name":"x","workloads":["AGG1"],"queries":["official"]}`,
		"bad_path":      `{"name":"x","workloads":["IS4"],"tunings":[{"path_preds":"eager"}]}`,
	} {
		p := filepath.Join(t.TempDir(), name+".json")
		_ = os.WriteFile(p, []byte(body), 0o644)
		if _, err := LoadSpec(p); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}

// A tuning sets the execution and planning options without touching settings.
func TestTuningOptions(t *testing.T) {
	off := false
	tu := Tuning{Vector: 7, Chunk: 3, Chunks: map[string]int{"mysql": 9}, Defer: &off}
	p, e := tu.options("IS4", "mongo", "vecstream")
	if e.VectorSize != 7 || e.MaterializeChunkSize != 3 || p.Logical.DeferPredicates || p.Physical.DeferPredicates {
		t.Errorf("options: %+v %+v", p, e)
	}
	if _, e := tu.options("IS4", "mysql", "bulk"); e.MaterializeChunkSize != 9 {
		t.Errorf("per-placement chunk: %d", e.MaterializeChunkSize)
	}
	if p, _ := (Tuning{PathPreds: "stepwise"}).options("IS4", "mongo", "bulk"); p.Physical.PathPredicates != physical.PathStepwise {
		t.Errorf("path_preds: %v", p.Physical.PathPredicates)
	}
}
