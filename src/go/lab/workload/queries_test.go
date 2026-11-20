package workload

import (
	"regexp"
	"strings"
	"testing"

	"polystore_database/src/go/data/migrate"
)

// officialExceptions gives the reason for each workload changed from the official text. The official queries use
// epoch milliseconds; the dump holds DateTime / Date, so only lines that would be type errors are changed.
var officialExceptions = map[string]string{
	"IC7":  "likeTime - creationDate as duration.inSeconds",
	"IC10": "datetime({epochMillis: birthday}) as birthday",
}

var (
	blockComment = regexp.MustCompile(`(?s)/\*.*?\*/`)
	lineComment  = regexp.MustCompile(`(?m)^\s*//.*$|\s+//.*$`)
)

// TestWorkloadsMatchOfficial checks that ldbc_snb workloads equal the official texts without comments.
func TestWorkloadsMatchOfficial(t *testing.T) {
	norm := func(s string) string {
		s = lineComment.ReplaceAllString(blockComment.ReplaceAllString(s, ""), "")
		return strings.Join(strings.Fields(s), " ")
	}
	queryOnce.Do(indexQueries)
	for k := range officialIndex {
		name := strings.TrimPrefix(k, "ldbc_snb/")
		off, _ := OfficialQuery("ldbc_snb", name)
		same := norm(query("ldbc_snb", name)) == norm(off)
		if _, ok := officialExceptions[name]; ok == same {
			t.Errorf("%s: same as official=%v but listed in officialExceptions=%v", name, same, ok)
		}
	}
}

// TestQueryFiles checks that every .cypher under queries is used by a workload (nothing left behind) and that every
// workload loads (no panic from overlapping names or missing files).
func TestQueryFiles(t *testing.T) {
	queryOnce.Do(indexQueries)
	byText := map[string][]string{} // Cypher -> the file holding it
	for _, p := range queryIndex {
		b, err := queryFS.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		s := strings.TrimSuffix(string(b), "\n")
		byText[s] = append(byText[s], p)
	}

	used := map[string]bool{}
	use := func(name string, def workloadDef) {
		c, _, _ := def(migrate.ModeGraphToRdb, false)
		ps, ok := byText[c]
		if !ok {
			t.Errorf("%s: the Cypher matches no file in queries", name)
		}
		for _, p := range ps {
			used[p] = true
		}
	}
	for ds, ws := range Datasets {
		for n, w := range ws {
			use(ds+"/"+n, w.Define)
		}
	}
	for _, p := range queryIndex {
		if !used[p] {
			t.Errorf("%s: no workload uses it", p)
		}
	}
	// the official set has workloads with the same names (simplified versions)
	for k, p := range officialIndex {
		if _, ok := queryIndex[k]; !ok {
			t.Errorf("%s: no workload of the same name", p)
		}
	}
}
