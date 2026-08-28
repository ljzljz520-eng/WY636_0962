package validation

import (
	"example.com/animalcage/internal/model"
	"testing"
)

func TestValidateRecord(t *testing.T) {
	issues := ValidateRecord(model.Record{CageCount: 0})
	if len(issues) < 5 {
		t.Fatalf("expected validation issues, got %d", len(issues))
	}
}

func TestValidateImport(t *testing.T) {
	issues := ValidateImport([]model.ImportRow{{ApplicationNo: "A", CageCount: 1, Roster: []string{"x"}}, {ApplicationNo: "A", CageCount: 1, Roster: []string{"y"}}})
	if !HasField(issues, "rows[1]") {
		t.Fatal("expected duplicate issue")
	}
}
