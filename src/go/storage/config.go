package storage

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	MappingPath string           `json:"mapping_path,omitempty"`
	Neo4j       *Neo4jConfig     `json:"neo4j,omitempty"`
	Mongo       *MongoConfig     `json:"mongo,omitempty"`
	LevelDB     *LevelDBConfig   `json:"leveldb,omitempty"`
	MySQL       *MySQLConfig     `json:"mysql,omitempty"`
	Cassandra   *CassandraConfig `json:"cassandra,omitempty"`
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
type LevelDBConfig struct {
	Path string `json:"path"`
}
type MySQLConfig struct {
	DSN string `json:"dsn"`
}
type CassandraConfig struct {
	Hosts    []string `json:"hosts"`
	Keyspace string   `json:"keyspace"`
}

func DefaultConfig() Config {
	neo4j := &Neo4jConfig{
		URI:      "bolt://localhost:7687",
		User:     "neo4j",
		Password: "password123",
	}
	mongo := &MongoConfig{
		URI:    "mongodb://localhost:27017",
		DBName: "polystore_doc",
	}
	level := &LevelDBConfig{
		Path: "../datastore/data/polystore_kvs",
	}
	mysql := &MySQLConfig{
		DSN: "user:pass@tcp(127.0.0.1:3306)/polystore_relational",
	}
	cassandra := &CassandraConfig{
		Hosts:    []string{"127.0.0.1"},
		Keyspace: "polystore_columnar",
	}

	return Config{
		MappingPath: "../catalog/mapping.json",
		Neo4j:       neo4j,
		Mongo:       mongo,
		LevelDB:     level,
		MySQL:       mysql,
		Cassandra:   cassandra,
	}
}

func LoadConfig(path string) (Config, error) {
	// If config json file don't exist, default config
	if path == "" {
		return DefaultConfig(), nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}
