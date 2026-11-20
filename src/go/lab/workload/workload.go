package workload

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"polystore_database/src/go/data/migrate"
	"sort"
	"strings"
	"sync"
)

// workloadDef is the signature of workload definition functions.
type workloadDef func(migrate.MigrationMode, bool) (string, map[string]string, []migrate.MigrationConfig)

// queryFS holds the workloads' Cypher (queries/<dataset>/[<kind>/]<workload>.cypher), embedded in the binary. Kind
// directories (ldbc_snb's interactive-complex / interactive-short / bi / custom) only organize; names are unique per
// dataset. queries/ldbc_snb/official/ copies the official implementations as they are (do not edit). The workloads in
// interactive-complex / interactive-short / bi are those texts without comments; lines changed for the dump's types
// are listed in officialExceptions in queries_test.go. OfficialQuery reads the official version.
//
//go:embed queries
var queryFS embed.FS

var (
	queryOnce     sync.Once
	queryIndex    map[string]string // "<ds>/<name>" -> path under queries
	officialIndex map[string]string // the same key under official/
)

// indexQueries indexes the .cypher files under queries by "<ds>/<name>" and panics on a duplicate name within a dataset.
func indexQueries() {
	queryIndex = map[string]string{}
	officialIndex = map[string]string{}
	err := fs.WalkDir(queryFS, "queries", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || path.Ext(p) != ".cypher" {
			return err
		}
		parts := strings.Split(strings.TrimPrefix(p, "queries/"), "/")
		key := parts[0] + "/" + strings.TrimSuffix(path.Base(p), ".cypher")
		idx := queryIndex
		if len(parts) > 2 && parts[1] == "official" {
			idx = officialIndex
		}
		if prev, ok := idx[key]; ok {
			return fmt.Errorf("%s and %s have the same name", prev, p)
		}
		idx[key] = p
		return nil
	})
	if err != nil {
		panic(fmt.Sprintf("workload: indexing queries: %v", err))
	}
}

// query reads dataset ds's workload name and returns it without its final newline; it panics if missing (a definition error).
func query(ds, name string) string {
	queryOnce.Do(indexQueries)
	p, ok := queryIndex[ds+"/"+name]
	if !ok {
		panic(fmt.Sprintf("workload: no query %s/%s", ds, name))
	}
	return readQuery(p)
}

// OfficialQuery returns the official version (under official/) of dataset ds's workload name, or false.
func OfficialQuery(ds, name string) (string, bool) {
	queryOnce.Do(indexQueries)
	p, ok := officialIndex[ds+"/"+name]
	if !ok {
		return "", false
	}
	return readQuery(p), true
}

func readQuery(p string) string {
	b, err := queryFS.ReadFile(p)
	if err != nil {
		panic(fmt.Sprintf("workload: %s: %v", p, err))
	}
	return strings.TrimSuffix(string(b), "\n")
}

// Def returns a workload's definition (nil if unknown).
func Def(name string) workloadDef {
	d, _, _ := Lookup(name)
	return d
}

// Available lists every workload name, sorted, for messages.
func Available() string {
	var names []string
	for _, ds := range DatasetNames() {
		names = append(names, Names(ds)...)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}
