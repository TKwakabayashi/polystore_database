package migrator

import (
	"context"
	"fmt"
	"polystore_database/src/go/codec"
	"polystore_database/src/go/id"
	"polystore_database/src/go/plan"
	"polystore_database/src/go/storage"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/syndtr/goleveldb/leveldb/util"
	"go.mongodb.org/mongo-driver/bson"
)

func fetchDataStream(ctx context.Context, cfg MigrationConfig, dbKind storage.StoreKind, reg *storage.Registry, typeMap map[string]string, dataCh chan<- DataRowStream) error {
	switch dbKind {
	case storage.Graph:
		return fetchGraph(ctx, cfg, reg, typeMap, dataCh)
	case storage.Kvs:
		return fetchKvs(ctx, cfg, reg, typeMap, dataCh)
	case storage.Document:
		return fetchDoc(ctx, cfg, reg, typeMap, dataCh)
	case storage.Columnar:
		return fetchCol(ctx, cfg, reg, typeMap, dataCh)
	case storage.Relational:
		return fetchRdb(ctx, cfg, reg, typeMap, dataCh)
	default:
		return fmt.Errorf("streaming fetch not implemented for: %s", dbKind.String())
	}
}

func fetchGraph(ctx context.Context, cfg MigrationConfig, reg *storage.Registry, typeMap map[string]string, dataCh chan<- DataRowStream) error {
	drv, ok := reg.Neo4j()
	if !ok {
		return fmt.Errorf("graph store not available")
	}
	session := drv.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	var label, query string
	if cfg.ObjType == plan.Relationship {
		label = "r"
		query = fmt.Sprintf("MATCH ()-[%s:%s]-() ", label, cfg.Entity)
	} else {
		label = "n"
		query = fmt.Sprintf("MATCH (%s:%s) ", label, cfg.Entity)
	}

	propList := []string{"uuid: " + label + ".uuid"}
	for _, p := range cfg.Properties {
		if p != "uuid" {
			propList = append(propList, fmt.Sprintf("%s: %s.%s", p, label, p))
		}
	}
	query += fmt.Sprintf("RETURN { %s } AS data", strings.Join(propList, ", "))

	res, err := session.Run(ctx, query, nil)
	if err != nil {
		return err
	}

	for res.Next(ctx) {
		record := res.Record()
		dataInterface, _ := record.Get("data")
		dataMap, ok := dataInterface.(map[string]interface{})
		if !ok {
			// data object error
			continue
		}

		uuid, _ := dataMap["uuid"].(id.UUID)
		if uuid == "" {
			// uuid empty
			continue
		}

		payload := make(map[string]interface{}, len(cfg.Properties))
		for _, p := range cfg.Properties {
			// ParseToNative に置換：Neo4jのfloat64をint32/int64へ
			nativeVal, err := codec.ParseToNative(dataMap[p], typeMap[p])
			if err != nil {
				return fmt.Errorf("parse error [graph] UUID %s, Prop %s: %w", uuid.String(), p, err)
			}
			payload[p] = nativeVal
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case dataCh <- DataRowStream{UUID: uuid, Payload: payload}:
		}
	}

	return nil
}

func fetchKvs(ctx context.Context, cfg MigrationConfig, reg *storage.Registry, typeMap map[string]string, dataCh chan<- DataRowStream) error {
	db, ok := reg.LevelDB()
	if !ok {
		return fmt.Errorf("kvs store not available")
	}
	prefix := cfg.Entity + codec.Sep
	iter := db.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	defer iter.Release()

	var currentUUID id.UUID
	var currentPayload map[string]interface{}

	flush := func() error {
		if currentUUID.Empty() {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case dataCh <- DataRowStream{UUID: currentUUID, Payload: currentPayload}:
		}
		return nil
	}

	for iter.Next() {
		parts := strings.Split(string(iter.Key()), codec.Sep)
		if len(parts) < 3 {
			continue
		}
		uuid, _ := id.ParseUUID(parts[1])
		propName := parts[2]

		targetType, ok := typeMap[propName]
		if !ok {
			continue
		}

		if uuid != currentUUID {
			if err := flush(); err != nil {
				return err
			}
			currentUUID = uuid
			currentPayload = make(map[string]interface{}, len(cfg.Properties))
		}

		rawVal := codec.DecodeValue(iter.Value(), targetType)
		// ParseToNative に置換
		nativeVal, err := codec.ParseToNative(rawVal, targetType)
		if err != nil {
			return fmt.Errorf("parse error [kvs] UUID %s, Prop %s: %w", uuid, propName, err)
		}
		currentPayload[propName] = nativeVal
	}
	return flush()
}

func fetchDoc(ctx context.Context, cfg MigrationConfig, reg *storage.Registry, typeMap map[string]string, dataCh chan<- DataRowStream) error {
	db, ok := reg.Mongo()
	if !ok {
		return fmt.Errorf("document store not available")
	}
	cur, err := db.Collection(cfg.Entity).Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var doc bson.M
		if err := cur.Decode(&doc); err != nil {
			return err
		}
		uuid, _ := doc["uuid"].(id.UUID)
		if uuid == "" {
			continue
		}

		payload := make(map[string]interface{}, len(cfg.Properties))
		for _, p := range cfg.Properties {
			// ParseToNative に置換
			nativeVal, err := codec.ParseToNative(doc[p], typeMap[p])
			if err != nil {
				return fmt.Errorf("parse error [document] UUID %s, Prop %s: %w", uuid, p, err)
			}
			payload[p] = nativeVal
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case dataCh <- DataRowStream{UUID: uuid, Payload: payload}:
		}
	}

	return nil
}

func fetchCol(ctx context.Context, cfg MigrationConfig, reg *storage.Registry, typeMap map[string]string, dataCh chan<- DataRowStream) error {
	session, ok := reg.Cassandra()
	if !ok {
		return fmt.Errorf("columnar store not available")
	}
	quotedProps := make([]string, len(cfg.Properties))
	for i, p := range cfg.Properties {
		quotedProps[i] = fmt.Sprintf("\"%s\"", p)
	}
	query := fmt.Sprintf("SELECT \"uuid\", %s FROM \"%s\"", strings.Join(quotedProps, ", "), cfg.Entity)
	iter := session.Query(query).WithContext(ctx).Iter()

	for {
		row := make(map[string]interface{})
		if !iter.MapScan(row) {
			break
		}

		uuid, _ := row["uuid"].(id.UUID)
		if uuid.Empty() {
			continue
		}

		payload := make(map[string]interface{}, len(cfg.Properties))
		for _, p := range cfg.Properties {
			var val interface{}
			if v, ok := row[p]; ok {
				val = v
			} else {
				for k, v := range row {
					if strings.EqualFold(k, p) {
						val = v
						break
					}
				}
			}
			// ParseToNative に置換
			nativeVal, err := codec.ParseToNative(val, typeMap[p])
			if err != nil {
				return fmt.Errorf("parse error [columnar] UUID %s, Prop %s: %w", uuid, p, err)
			}
			payload[p] = nativeVal
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case dataCh <- DataRowStream{UUID: uuid, Payload: payload}:
		}
	}

	return nil
}

func fetchRdb(ctx context.Context, cfg MigrationConfig, reg *storage.Registry, typeMap map[string]string, dataCh chan<- DataRowStream) error {
	db, ok := reg.MySQL()
	if !ok {
		return fmt.Errorf("relational store not available")
	}
	query := fmt.Sprintf("SELECT uuid, %s FROM %s", strings.Join(cfg.Properties, ", "), cfg.Entity)
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return err
		}

		payload := make(map[string]interface{}, len(cfg.Properties))
		var uuid id.UUID
		for i, colName := range cols {
			val := columns[i]
			if b, ok := val.([]byte); ok {
				val = string(b)
			}

			if colName == "uuid" {
				uuid, _ = val.(id.UUID)
				continue
			}
			// ParseToNative に置換
			nativeVal, err := codec.ParseToNative(val, typeMap[colName])
			if err != nil {
				return fmt.Errorf("parse error [relational] UUID %s, Prop %s: %w", uuid.String(), colName, err)
			}
			payload[colName] = nativeVal
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case dataCh <- DataRowStream{UUID: uuid, Payload: payload}:
		}
	}
	return nil
}
