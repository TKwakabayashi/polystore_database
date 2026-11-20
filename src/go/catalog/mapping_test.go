package catalog

import "testing"

func TestReplicaAndPlacement(t *testing.T) {
	md := &MappingDictionary{Entities: map[string]*EntityMapping{"Person": {Properties: map[string]*PropertyMapping{
		"firstName": {Type: "string", Databases: []string{"graph"}},
	}}}}
	if err := md.AddDatastore(ObjEntity, "Person", []string{"firstName"}, "mysql"); err != nil {
		t.Fatal(err)
	}
	if err := md.AddDatastore(ObjEntity, "Person", []string{"firstName"}, "relational"); err != nil { // same store, legacy name
		t.Fatal(err)
	}
	if got := md.Entities["Person"].Properties["firstName"].Databases; len(got) != 2 || got[1] != "mysql" {
		t.Fatalf("record = %v", got)
	}
	run, err := md.Placed(ObjEntity, "Person", []string{"firstName"}, "mysql")
	if err != nil {
		t.Fatal(err)
	}
	if got := run.Entities["Person"].Properties["firstName"].Databases; len(got) != 1 || got[0] != "mysql" {
		t.Errorf("runtime placement = %v", got)
	}
	if got := md.Entities["Person"].Properties["firstName"].Databases; len(got) != 2 {
		t.Errorf("record changed by Placed: %v", got)
	}
}
