package mongo

import (
	"reflect"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// rawToGo returns the same Go values as decoding the document into bson.M (what scanCursor relies on).
func TestRawToGoMatchesBSONM(t *testing.T) {
	doc := bson.D{
		{Key: "s", Value: "Müller"},
		{Key: "i32", Value: int32(7)},
		{Key: "i64", Value: int64(1 << 40)},
		{Key: "f", Value: 1.5},
		{Key: "b", Value: true},
		{Key: "dt", Value: primitive.NewDateTimeFromTime(time.Date(2012, 2, 22, 14, 5, 58, 785e6, time.UTC))},
		{Key: "bin", Value: primitive.Binary{Subtype: 0, Data: []byte{0, 1, 2, 255}}},
		{Key: "null", Value: nil},
		{Key: "arr", Value: bson.A{"a", int32(1), nil}},
		{Key: "doc", Value: bson.D{{Key: "x", Value: int32(1)}}},
	}
	raw, err := bson.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	var m bson.M
	if err := bson.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	for _, e := range doc {
		rv, err := bson.Raw(raw).LookupErr(e.Key)
		if err != nil {
			t.Fatal(err)
		}
		got, err := rawToGo(rv)
		if err != nil {
			t.Fatalf("%s: %v", e.Key, err)
		}
		if want := m[e.Key]; !reflect.DeepEqual(fromBSON(got), fromBSON(want)) {
			t.Errorf("%s: rawToGo = %#v, bson.M = %#v", e.Key, got, want)
		}
	}
}
