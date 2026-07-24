package schema

import (
	"testing"

	"polystore_database/src/go/plan"
)

// mappingPath は追跡済みの実カタログ（リポジトリ直下 catalog/mapping.json）。
// go test の cwd は src/go/schema なので、リポジトリルートは 3 段上。
const mappingPath = "../../../catalog/mapping.json"

func loadForTest(t *testing.T) *MappingDictionary {
	t.Helper()
	md, err := LoadMappingDictionary(mappingPath)
	if err != nil {
		t.Fatalf("LoadMappingDictionary(%s): %v", mappingPath, err)
	}
	return md
}

func TestLookupEntityProperty(t *testing.T) {
	md := loadForTest(t)
	cases := []struct {
		label, prop      string
		wantStore, wantT string
	}{
		{"Person", "firstName", "graph", "string"},
		{"Person", "id", "graph", "long"},
		{"Person", "creationDate", "graph", "string"}, // 実データが文字列なので string に整合（旧 datetime、§C）
	}
	for _, c := range cases {
		st, ty, err := md.LookupMappingDictionary(plan.Entity, c.label, c.prop)
		if err != nil {
			t.Errorf("Lookup(%s.%s): %v", c.label, c.prop, err)
			continue
		}
		if st != c.wantStore || ty != c.wantT {
			t.Errorf("Lookup(%s.%s) = (%s,%s), want (%s,%s)", c.label, c.prop, st, ty, c.wantStore, c.wantT)
		}
	}
}

func TestLookupRelationshipProperty(t *testing.T) {
	md := loadForTest(t)
	// KNOWS.creationDate は関係プロパティ。存在し型が引ければよい（配置は環境依存）。
	if _, ty, err := md.LookupMappingDictionary(plan.Relationship, "KNOWS", "creationDate"); err != nil {
		t.Errorf("Lookup(KNOWS.creationDate): %v", err)
	} else if ty == "" {
		t.Errorf("KNOWS.creationDate の型が空")
	}
}

func TestLookupUnknownPropertyErrors(t *testing.T) {
	md := loadForTest(t)
	if _, _, err := md.LookupMappingDictionary(plan.Entity, "Person", "no_such_prop"); err == nil {
		t.Errorf("未知プロパティでエラーになるはず")
	}
	if _, _, err := md.LookupMappingDictionary(plan.Entity, "NoSuchEntity", "x"); err == nil {
		t.Errorf("未知エンティティでエラーになるはず")
	}
}

// nil レシーバでも panic せずエラーを返す契約。
func TestLookupNilDictionary(t *testing.T) {
	var md *MappingDictionary
	if _, _, err := md.LookupMappingDictionary(plan.Entity, "Person", "id"); err == nil {
		t.Errorf("nil 辞書でエラーになるはず")
	}
}
