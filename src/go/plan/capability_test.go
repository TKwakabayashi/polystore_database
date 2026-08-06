package plan

import (
	"testing"

	"polystore_database/src/go/store"
)

func TestSupports(t *testing.T) {
	cases := []struct {
		k    store.Kind
		c    OpCapability
		want bool
	}{
		{store.Graph, CapExpand, true},
		{store.Graph, CapAggregate, true},
		{store.Relational, CapExpand, false},
		{store.Relational, CapGroupBy, true},
		{store.Document, CapDistinct, true},
		{store.Columnar, CapAggregate, true},
		{store.Columnar, CapGroupBy, false},
		{store.Columnar, CapSort, false},
		{store.Kvs, CapProject, true},
		{store.Kvs, CapAggregate, false},
	}
	for _, tc := range cases {
		if got := Supports(tc.k, tc.c); got != tc.want {
			t.Errorf("Supports(%s, %s) = %v, want %v", tc.k, tc.c, got, tc.want)
		}
	}
}
