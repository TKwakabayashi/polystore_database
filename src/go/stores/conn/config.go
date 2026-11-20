package conn

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config is config/config.json: the running dataset, where the datastore directory is, and the database
// connections. Paths are relative to the config file. The record mapping, statistics, embedded stores and
// the uuid lease ledger follow from the datastore directory (Layout).
type Config struct {
	Dataset    string            `json:"dataset"`   // a folder in datastore/datasets (the running dataset)
	Datastore  string            `json:"datastore"` // the datastore directory (default ../datastore)
	Neo4j      *Neo4jConfig      `json:"neo4j,omitempty"`
	Mongo      *MongoConfig      `json:"mongo,omitempty"`
	MySQL      *MySQLConfig      `json:"mysql,omitempty"`
	ClickHouse *ClickHouseConfig `json:"clickhouse,omitempty"` // columnar
	Cassandra  *CassandraConfig  `json:"cassandra,omitempty"`  // widecolumn

	dir string // the config file's directory (relative paths start here)
}

type Neo4jConfig struct {
	URI      string `json:"uri"`
	User     string `json:"user"`
	Password string `json:"password"`
}
type MongoConfig struct {
	URI    string `json:"uri"`
	DBName string `json:"dbname"`
}
type ClickHouseConfig struct {
	Addr     []string `json:"addr"` // native protocol (host:9000)
	Database string   `json:"database"`
	User     string   `json:"user"`
	Password string   `json:"password"`
}
type MySQLConfig struct {
	DSN string `json:"dsn"`
}
type CassandraConfig struct {
	Hosts    []string `json:"hosts"`
	Keyspace string   `json:"keyspace"`
}

// defaultDatastore is the datastore directory relative to config/.
const defaultDatastore = "../datastore"

// Layout returns where the datastore directory keeps the dataset files and the running dataset's state.
func (c Config) Layout() Layout {
	d := c.Datastore
	if d == "" {
		d = defaultDatastore
	}
	if !filepath.IsAbs(d) && c.dir != "" {
		d = filepath.Join(c.dir, d)
	}
	return Layout{Dir: d}
}

// MappingPath is the running dataset's record mapping (datastore/runtime/mapping.json).
func (c Config) MappingPath() string { return c.Layout().MappingPath() }

// DefaultConfig is the local docker stack's default (used when there is no config file).
func DefaultConfig() Config {
	return Config{
		Dataset:   "ldbc_snb_sf1",
		Datastore: defaultDatastore,
		Neo4j:     &Neo4jConfig{URI: "bolt://localhost:7687", User: "neo4j", Password: "password123"},
		Mongo:     &MongoConfig{URI: "mongodb://localhost:27018", DBName: "polystore_document"},
		MySQL:     &MySQLConfig{DSN: "user:pass@tcp(127.0.0.1:3306)/polystore_relational"},
		ClickHouse: &ClickHouseConfig{
			Addr:     []string{"127.0.0.1:9000"},
			Database: "polystore_columnar",
			User:     "default",
		},
		Cassandra: &CassandraConfig{Hosts: []string{"127.0.0.1"}, Keyspace: "polystore_widecolumn"},
	}
}

// LoadConfig reads the config file at path (the default config when path is empty or the file is missing).
func LoadConfig(path string) (Config, error) {
	cfg := DefaultConfig()
	if path == "" {
		return cfg, nil
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		if os.IsNotExist(err) {
			cfg.dir = filepath.Dir(abs)
			return cfg, nil
		}
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	cfg = Config{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	cfg.dir = filepath.Dir(abs)
	return cfg, nil
}
