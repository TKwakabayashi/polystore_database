package query

import "testing"

func TestTypedParamsIntegerList(t *testing.T) {
	got, ok := typedParams(map[string]string{"ids": `[9007199254740993, 1.5]`})["ids"].([]any)
	if !ok || len(got) != 2 || got[0] != int64(9007199254740993) || got[1] != 1.5 {
		t.Errorf("values differ: %#v", got)
	}
}
