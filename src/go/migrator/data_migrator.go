package migrator

import (
	"context"
	"fmt"

	schema "polystore_database/src/go/schema"
	"polystore_database/src/go/storage"

	"golang.org/x/sync/errgroup"
)

func MigrateData(config MigrationConfig, sconfig storage.Config) error {
	mappingDictionary, err := schema.LoadMappingDictionary(config.MappingPath)
	if err != nil {
		return err
	}

	srcKind, destKind, err := modeStores(config.Mode)
	if err != nil {
		return err
	}
	if err := ExecuteMigrationStream(config, sconfig, srcKind, destKind, mappingDictionary); err != nil {
		return err
	}
	/*
		if err := executeMigrationBulk(mappingDictionary, config, src, dest); err != nil {
			return err
		}
	*/
	return mappingDictionary.SaveMappingDictionary(config.MappingPath)
}

func ExecuteMigrationStream(config MigrationConfig, scfg storage.Config, srcKind, destKind storage.StoreKind, md *schema.MappingDictionary) error {
	if !md.CheckDatastore(config.ObjType, config.Entity, config.Properties, srcKind.String()) {
		return fmt.Errorf("data not found")
	}

	typeMap := make(map[string]string)
	for _, prop := range config.Properties {
		typeMap[prop], _ = md.GetPropertyDataType(config.ObjType, config.Entity, prop)
	}

	mainCtx, mainCancel := context.WithCancel(context.Background())
	defer mainCancel()

	reg, err := storage.NewRegistryFor(mainCtx, scfg, srcKind, destKind)
	if err != nil {
		return fmt.Errorf("registry init: %w", err)
	}
	defer reg.Close(mainCtx)

	// いずれかのフェーズでエラーが返ると、ctx.Done() が閉じられ全フェーズが停止する
	eg, ctx := errgroup.WithContext(mainCtx)

	fetchWorkers := 1
	upsertWorkers := 4
	deleteWorkers := 2

	// 2. パイプラインを繋ぐチャネル
	// 16GBのメモリを考慮し、バッファは1万件程度（1件1KBとしても数10MB）
	fetchToUpsertCh := make(chan DataRowStream, 10000)
	upsertToDeleteCh := make(chan DataRowStream, 10000)

	eg.Go(func() error {
		defer close(fetchToUpsertCh)
		fEg, fCtx := errgroup.WithContext(ctx)
		for i := 0; i < fetchWorkers; i++ {
			fEg.Go(func() error {
				// 注意: fetchDataStream内で重複してデータを取らない工夫が将来的に必要
				return fetchDataStream(fCtx, config, srcKind, reg, typeMap, fetchToUpsertCh)
			})
		}
		return fEg.Wait()
	})

	eg.Go(func() error {
		defer close(upsertToDeleteCh)
		uEg, uCtx := errgroup.WithContext(ctx)
		for i := 0; i < upsertWorkers; i++ {
			uEg.Go(func() error {
				return upsertDataStream(uCtx, config, destKind, reg, fetchToUpsertCh, upsertToDeleteCh, typeMap)
			})
		}
		return uEg.Wait()
	})

	eg.Go(func() error {
		dEg, dCtx := errgroup.WithContext(ctx)
		for i := 0; i < deleteWorkers; i++ {
			dEg.Go(func() error {
				return deleteDataStream(dCtx, config, srcKind, reg, upsertToDeleteCh, typeMap)
			})
		}
		return dEg.Wait()
	})

	if err := eg.Wait(); err != nil {
		reg.Close(mainCtx)
		return err
	}

	if err := VerifyMigration(mainCtx, config, destKind, reg); err != nil {
		return err
	}

	// 5. MappingDictionaryの位置情報を更新
	md.UpdateDatastore(config.ObjType, config.Entity, config.Properties, destKind.String())

	fmt.Println("✅ 移行が正常に完了しました。")
	return nil
}
