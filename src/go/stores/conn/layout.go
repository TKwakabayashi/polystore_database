package conn

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"polystore_database/src/go/base/kind"
	"sort"
	"strings"
)

// Layout is where the datastore directory (datastore/ in the repository) keeps things:
//
//	docker/                  compose.yaml (every dataset)
//	datasets/<dataset>/      static: dataset.env, mapping.json (schema template), stats.json, dumps/
//	shared/uuid_leases.json  the uuid lease ledger of every dataset
//	runtime/                 the running dataset only, rebuilt by setup: dataset, mapping.json (the record),
//	                         stats.json, leveldb/, rocksdb/
type Layout struct {
	Dir string // the datastore directory
}

// DatasetDir is a dataset's static folder.
func (l Layout) DatasetDir(dataset string) string { return filepath.Join(l.Dir, "datasets", dataset) }

// Datasets lists the dataset folders.
func (l Layout) Datasets() ([]string, error) {
	es, err := os.ReadDir(filepath.Join(l.Dir, "datasets"))
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range es {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out, nil
}

// RuntimeDir holds the running dataset's state.
func (l Layout) RuntimeDir() string { return filepath.Join(l.Dir, "runtime") }

// MappingPath is the record mapping of the running dataset (stats.json sits next to it).
func (l Layout) MappingPath() string { return filepath.Join(l.RuntimeDir(), "mapping.json") }

// KVDir is the data directory of embedded store k ("" for other kinds).
func (l Layout) KVDir(k kind.Kind) string {
	switch k {
	case kind.LevelDB:
		return filepath.Join(l.RuntimeDir(), "leveldb")
	case kind.RocksDB:
		return filepath.Join(l.RuntimeDir(), "rocksdb")
	}
	return ""
}

// LeaseLedger is the uuid lease ledger shared by every dataset.
func (l Layout) LeaseLedger() string { return filepath.Join(l.Dir, "shared", "uuid_leases.json") }

// runtimeDatasetFile names the dataset whose state runtime/ holds.
const runtimeDatasetFile = "dataset"

// RuntimeDataset returns the dataset runtime/ was built for ("" when it was never built).
func (l Layout) RuntimeDataset() (string, error) {
	b, err := os.ReadFile(filepath.Join(l.RuntimeDir(), runtimeDatasetFile))
	if os.IsNotExist(err) {
		return "", nil
	}
	return strings.TrimSpace(string(b)), err
}

// SetRuntimeDataset records the dataset runtime/ holds.
func (l Layout) SetRuntimeDataset(dataset string) error {
	return os.WriteFile(filepath.Join(l.RuntimeDir(), runtimeDatasetFile), []byte(dataset+"\n"), 0o644)
}

// DatasetOfProject returns the dataset whose dataset.env names compose project suffix name (VOLUMES when
// set, else DATASET; containers made before a rename keep the older name in VOLUMES).
func (l Layout) DatasetOfProject(name string) (string, error) {
	ds, err := l.Datasets()
	if err != nil {
		return "", err
	}
	for _, d := range ds {
		env, err := readEnv(filepath.Join(l.DatasetDir(d), "dataset.env"))
		if err != nil {
			continue
		}
		project := env["VOLUMES"]
		if project == "" {
			project = env["DATASET"]
		}
		if project == name {
			return d, nil
		}
	}
	return "", fmt.Errorf("no dataset in %s uses compose project %q", filepath.Join(l.Dir, "datasets"), name)
}

// readEnv reads KEY=VALUE lines (comments and blank lines skipped).
func readEnv(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if k, v, ok := strings.Cut(line, "="); ok {
			out[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return out, sc.Err()
}

// CheckRuntime reports whether runtime/ was built for the config's dataset (admin setup builds it).
func (c Config) CheckRuntime() error {
	got, err := c.Layout().RuntimeDataset()
	if err != nil {
		return err
	}
	switch {
	case got == "":
		return fmt.Errorf("%s was never built; run admin setup", c.Layout().RuntimeDir())
	case c.Dataset != "" && got != c.Dataset:
		return fmt.Errorf("the config's dataset is %s but %s holds %s; run admin setup", c.Dataset, c.Layout().RuntimeDir(), got)
	}
	return nil
}
