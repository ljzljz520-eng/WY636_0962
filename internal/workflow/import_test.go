package workflow

import (
	"testing"
	"time"

	"example.com/animalcage/internal/model"
	"example.com/animalcage/internal/validation"
)

func TestImportRejectsDuplicateApplicationNumbers(t *testing.T) {
	service, _ := newTestService()
	result, err := service.Import([]model.ImportRow{{ApplicationNo: "D-1", Applicant: "A", Facility: "F", Species: "mouse", CageCount: 1, Roster: []string{"m1"}}, {ApplicationNo: "D-1", Applicant: "B", Facility: "F", Species: "mouse", CageCount: 1, Roster: []string{"m2"}}}, "importer")
	if err == nil || !validation.HasField(result.Issues, "rows[1]") {
		t.Fatalf("duplicate import result=%#v err=%v", result, err)
	}
}

func TestDefaultClockIsDeterministic(t *testing.T) {
	service := NewService(newMemoryStore(), nil)
	record, err := service.Register(RegistrationInput{ApplicationNo: "CLOCK-1", Applicant: "A", Facility: "F", Species: "mouse", CageCount: 1, Roster: []string{"m1"}}, "tester")
	if err != nil {
		t.Fatal(err)
	}
	if !record.CreatedAt.Equal(time.Unix(0, 0).UTC()) || !record.UpdatedAt.Equal(time.Unix(0, 0).UTC()) {
		t.Fatalf("default clock = created %v updated %v", record.CreatedAt, record.UpdatedAt)
	}
}
