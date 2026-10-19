// zz_rebuild は一時的な移行の作り直しツール（作業後に削除する）。
// 全ワークロードの移行設定の和集合を、graph から指定ストアへ写す。mapping.json は書き換えない。
package main

import (
	"flag"
	"fmt"
	"log"
	"sort"
	"time"

	"polystore_database/src/go/migrator"
	"polystore_database/src/go/plan"
	"polystore_database/src/go/schema"
	"polystore_database/src/go/storage"
	"polystore_database/src/go/workload"
)

func main() {
	target := flag.String("target", "", "relational | document | columnar | kvs")
	dry := flag.Bool("dry", false, "移行せず一覧だけ出す")
	flag.Parse()
	modes := map[string]migrator.MigrationMode{
		"relational": migrator.ModeGraphToRdb, "document": migrator.ModeGraphToDoc,
		"columnar": migrator.ModeGraphToCol, "kvs": migrator.ModeGraphToKvs,
	}
	mode, ok := modes[*target]
	if !ok {
		log.Fatalf("unknown target %q", *target)
	}
	cfg, err := storage.LoadConfig("../../config/config.json")
	if err != nil {
		log.Fatal(err)
	}
	base, err := schema.LoadMappingDictionary(cfg.MappingPath)
	if err != nil {
		log.Fatal(err)
	}

	type key struct {
		obj    plan.ObjectType
		entity string
	}
	union := map[key]map[string]bool{}
	for _, def := range workload.Registry {
		_, _, migs := def(mode, true)
		for _, m := range migs {
			k := key{m.ObjType, m.Entity}
			if union[k] == nil {
				union[k] = map[string]bool{}
			}
			for _, p := range m.Properties {
				if stores, err := base.GetDatastores(m.ObjType, m.Entity, p); err == nil && len(stores) > 0 {
					union[k][p] = true
				}
			}
		}
	}
	var keys []key
	for k := range union {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].entity < keys[j].entity })

	for _, k := range keys {
		var props []string
		for p := range union[k] {
			props = append(props, p)
		}
		sort.Strings(props)
		if len(props) == 0 {
			continue
		}
		fmt.Printf("%-14s %v\n", k.entity, props)
		if *dry {
			continue
		}
		md, _ := schema.LoadMappingDictionary(cfg.MappingPath) // 毎回 all-graph の写しから（保存しない）
		mc := migrator.MigrationConfig{ObjType: k.obj, Entity: k.entity, Properties: props, Mode: mode, MappingPath: cfg.MappingPath}
		if cfg.Mongo != nil {
			mc.MongoDbName = cfg.Mongo.DBName
		}
		_, dest, _ := migrator.ModeStores(mode)
		start := time.Now()
		n, err := migrator.ExecuteMigrationStream(mc, cfg, storage.Graph, dest, md)
		if err != nil {
			log.Fatalf("%s: %v", k.entity, err)
		}
		fmt.Printf("  -> %d 件 %v\n", n, time.Since(start).Round(time.Second))
	}
}
