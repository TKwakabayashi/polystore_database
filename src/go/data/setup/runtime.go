package setup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"polystore_database/src/go/base/id"
	"polystore_database/src/go/catalog"
	"polystore_database/src/go/stores/conn"
)

// PrepareRuntime rebuilds the datastore's runtime/ for dataset: it removes the previous state (the record,
// statistics and embedded stores of whichever dataset ran before), copies the dataset's mapping.json and
// stats.json as the new record (every property on Neo4j), sets the record's uuid storage form to what
// Neo4j holds, and records the dataset name. RunDataset clears the other stores.
func PrepareRuntime(ctx context.Context, l conn.Layout, cfg conn.Config, dataset string) error {
	src := l.DatasetDir(dataset)
	if _, err := os.Stat(filepath.Join(src, "mapping.json")); err != nil {
		return fmt.Errorf("dataset %s has no mapping.json: %w", dataset, err)
	}
	rt := l.RuntimeDir()
	for _, name := range []string{"dataset", "mapping.json", "stats.json", "leveldb", "rocksdb"} {
		if err := os.RemoveAll(filepath.Join(rt, name)); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(rt, 0o755); err != nil {
		return err
	}
	if err := copyFile(filepath.Join(src, "mapping.json"), l.MappingPath()); err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(src, "stats.json")); err == nil {
		if err := copyFile(filepath.Join(src, "stats.json"), filepath.Join(rt, "stats.json")); err != nil {
			return err
		}
	}
	form, err := neo4jUUIDStorage(ctx, cfg)
	if err != nil {
		return err
	}
	md, err := catalog.LoadMappingDictionary(l.MappingPath())
	if err != nil {
		return err
	}
	md.UUIDStorage = form.String()
	if err := md.SaveMappingDictionary(l.MappingPath()); err != nil {
		return err
	}
	return l.SetRuntimeDataset(dataset)
}

// neo4jUUIDStorage reads the storage form of Neo4j's uuids from one value (text when none has a uuid yet).
func neo4jUUIDStorage(ctx context.Context, cfg conn.Config) (id.Storage, error) {
	if cfg.Neo4j == nil {
		return id.StorageText, fmt.Errorf("no neo4j in the config")
	}
	driver, err := neo4j.NewDriverWithContext(cfg.Neo4j.URI, neo4j.BasicAuth(cfg.Neo4j.User, cfg.Neo4j.Password, ""))
	if err != nil {
		return id.StorageText, err
	}
	defer driver.Close(ctx)
	res, err := neo4j.ExecuteQuery(ctx, driver, "MATCH (n) WHERE n.uuid IS NOT NULL RETURN n.uuid AS u LIMIT 1",
		nil, neo4j.EagerResultTransformer)
	if err != nil {
		return id.StorageText, fmt.Errorf("read a uuid from neo4j: %w", err)
	}
	if len(res.Records) > 0 {
		if _, ok := res.Records[0].Values[0].([]byte); ok {
			return id.StorageBinary, nil
		}
	}
	return id.StorageText, nil
}

func copyFile(src, dst string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, b, 0o644)
}
