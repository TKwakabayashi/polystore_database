package kv

import (
	"context"
	"fmt"
	"polystore_database/src/go/base/kind"

	"polystore_database/src/go/base/codec"
	"polystore_database/src/go/base/id"
	"polystore_database/src/go/stores"
	"polystore_database/src/go/stores/conn"
)

func fetchKvs(ctx context.Context, cfg stores.Target, kind kind.Kind, reg *conn.Registry, typeMap map[string]string, dataCh chan<- stores.Row, shardIdx, shardTotal int) error {
	db, ok := reg.KV(kind)
	if !ok {
		return fmt.Errorf("kvs store not available")
	}
	// Keys are Entity + Sep + uuid + Sep + prop; the uuid's leading byte bounds the range. All keys of one uuid
	// share its prefix, so shard bounds never split them.
	base := []byte(cfg.Entity + codec.Sep)
	start, limit := conn.PrefixRange(base)
	if lo, hi := id.ShardBounds(shardIdx, shardTotal); lo != nil || hi != nil {
		if lo != nil {
			start = append(append([]byte{}, base...), lo...)
		}
		if hi != nil {
			limit = append(append([]byte{}, base...), hi...)
		}
	}

	var currentUUID id.UUID
	var currentPayload map[string]interface{}

	flush := func() error {
		if currentUUID.Empty() {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case dataCh <- stores.Row{UUID: currentUUID, Values: currentPayload}:
		}
		return nil
	}

	err := db.Scan(start, limit, func(key, val []byte) error {
		_, uuid, propName, ok := codec.ParseEntityKey(key)
		if !ok {
			return nil
		}

		targetType, ok := typeMap[propName]
		if !ok {
			return nil
		}

		if uuid != currentUUID {
			if err := flush(); err != nil {
				return err
			}
			currentUUID = uuid
			currentPayload = make(map[string]interface{}, len(cfg.Props))
		}

		rawVal := codec.DecodeValue(append([]byte{}, val...), targetType)
		// ParseToNative
		nativeVal, err := codec.ParseToNative(rawVal, targetType)
		if err != nil {
			return fmt.Errorf("parse error [kvs] UUID %s, Prop %s: %w", uuid, propName, err)
		}
		currentPayload[propName] = nativeVal
		return nil
	})
	if err != nil {
		return err
	}
	return flush()
}

func upsertKvs(ctx context.Context, cfg stores.Target, kind kind.Kind, reg *conn.Registry, typeMap map[string]string, rows []stores.Row) error {
	db, ok := reg.KV(kind)
	if !ok {
		return fmt.Errorf("kvs store not available")
	}
	levelBatch := new(conn.KVBatch)
	for _, row := range rows {
		for p, val := range row.Values {
			if val == nil {
				continue
			}
			// KVS binary encoding
			valBytes, err := codec.EncodeForKVS(val, typeMap[p])
			if err != nil {
				return fmt.Errorf("kvs encode failed: %w", err)
			}

			// json (complex values) is stored only; no reverse index (no value search)
			complex := codec.IsComplexType(typeMap[p])

			entityKey := codec.BuildEntityKey(cfg.Entity, row.UUID, p)
			if !complex {
				if oldVal, found, err := db.Get(entityKey); err != nil {
					return fmt.Errorf("kvs read failed: %w", err)
				} else if found {
					levelBatch.Delete(codec.BuildIndexKey(cfg.Entity, p, oldVal, row.UUID))
				}
			}
			levelBatch.Put(entityKey, valBytes)
			if !complex {
				levelBatch.Put(codec.BuildIndexKey(cfg.Entity, p, valBytes, row.UUID), []byte{})
			}
		}
	}
	if err := db.Write(levelBatch); err != nil {
		return fmt.Errorf("kvs write failed: %w", err)
	}

	return nil
}

func deleteKvs(ctx context.Context, cfg stores.Target, kind kind.Kind, reg *conn.Registry, typeMap map[string]string, rows []stores.Row) error {
	db, ok := reg.KV(kind)
	if !ok {
		return fmt.Errorf("kvs store not available")
	}
	levelBatch := new(conn.KVBatch)
	for _, r := range rows {
		for _, p := range cfg.Props {
			entityKey := codec.BuildEntityKey(cfg.Entity, r.UUID, p)
			// json (complex values) has no reverse index, so only the entity key is deleted
			complex := codec.IsComplexType(typeMap[p])
			if val, found, err := db.Get(entityKey); err != nil {
				return fmt.Errorf("kvs read failed: %w", err)
			} else if found {
				levelBatch.Delete(entityKey)
				if !complex {
					levelBatch.Delete(codec.BuildIndexKey(cfg.Entity, p, val, r.UUID))
				}
			}
		}
	}
	// wait for the write to reach disk
	if err := db.Write(levelBatch); err != nil {
		return fmt.Errorf("kvs delete failed: %w", err)
	}

	return nil
}

func countKvs(ctx context.Context, cfg stores.Target, kind kind.Kind, reg *conn.Registry) (int64, error) {
	db, ok := reg.KV(kind)
	if !ok {
		return 0, fmt.Errorf("kvs store not available")
	}
	// Keys are Entity + Sep + uuid + Sep + prop; a uuid has several keys, so count distinct uuids.
	prefix := cfg.Entity + codec.Sep

	var count int64
	var lastUUID id.UUID
	first := true
	err := conn.ScanPrefix(db, []byte(prefix), func(key, _ []byte) error {
		_, u, _, ok := codec.ParseEntityKey(key)
		if !ok {
			return nil
		}
		if first || u != lastUUID {
			count++
			lastUUID = u
			first = false
		}
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("kvs: scan %s: %w", cfg.Entity, err)
	}
	return count, nil
}

// Prepare does nothing: keys are entity + uuid + property, and the value index is kept on upsert.
func (DB) Prepare(context.Context, *conn.Registry, *stores.Target) error { return nil }

// ScanWorkers is 4 (uuid key range shards).
func (DB) ScanWorkers() int { return 4 }
func (d DB) Scan(ctx context.Context, reg *conn.Registry, t stores.Target, out chan<- stores.Row, shard, shards int) error {
	return fetchKvs(ctx, t, d.K, reg, t.Types, out, shard, shards)
}
func (d DB) Upsert(ctx context.Context, reg *conn.Registry, t stores.Target, rows []stores.Row) error {
	return upsertKvs(ctx, t, d.K, reg, t.Types, rows)
}
func (DB) Finish(context.Context, *conn.Registry, stores.Target) error { return nil }
func (d DB) Delete(ctx context.Context, reg *conn.Registry, t stores.Target, rows []stores.Row) error {
	return deleteKvs(ctx, t, d.K, reg, t.Types, rows)
}
func (d DB) Count(ctx context.Context, reg *conn.Registry, t stores.Target) (int64, error) {
	return countKvs(ctx, t, d.K, reg)
}
