package workflow

import (
	"testing"
	"time"

	"example.com/animalcage/internal/model"
)

func newTestService() (*Service, *memoryStore) {
	store := newMemoryStore()
	return NewService(store, FixedClock(time.Date(2026, 8, 25, 8, 0, 0, 0, time.UTC))), store
}

func TestWorkflowCreateReviewArchive(t *testing.T) {
	service, _ := newTestService()
	record, err := service.Register(RegistrationInput{ApplicationNo: "A-100", Applicant: "Li", Facility: "North", Species: "mouse", CageCount: 2, Roster: []string{"m-1", "m-2"}}, "registrar")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.Submit(record.ID, "registrar"); err != nil {
		t.Fatal(err)
	}
	if record, err = service.Review(record.ID, model.ReviewDecision{Approved: true, Reviewer: "reviewer", Comment: "ready"}); err != nil {
		t.Fatal(err)
	}
	if record.Status != model.StatusApproved {
		t.Fatalf("status = %s", record.Status)
	}
	if record, err = service.Archive(record.ID, "archiver"); err != nil {
		t.Fatal(err)
	}
	if record.Status != model.StatusArchived || record.ArchivedAt == nil {
		t.Fatal("archive did not close record")
	}
	report, err := service.Report(record.ID)
	if err != nil || len(report.Events) < 4 || len(report.Workflows) < 3 {
		t.Fatalf("report incomplete: err=%v events=%d workflows=%d", err, len(report.Events), len(report.Workflows))
	}
}

func TestWorkflowSearchUpdatePublish(t *testing.T) {
	service, _ := newTestService()
	record, err := service.Register(RegistrationInput{ApplicationNo: "A-200", Applicant: "Wang", Facility: "East", Species: "rat", CageCount: 1, Roster: []string{"r-1"}}, "operator")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.Submit(record.ID, "operator"); err != nil {
		t.Fatal(err)
	}
	if _, err = service.Review(record.ID, model.ReviewDecision{Approved: true, Reviewer: "reviewer"}); err != nil {
		t.Fatal(err)
	}
	if _, err = service.Change(record.ID, model.ChangeRequest{Actor: "operator", NewRoster: []string{"r-2"}, NewCageCount: 1, Reason: "replacement"}); err != nil {
		t.Fatal(err)
	}
	rows, err := service.Search(model.SearchFilter{Applicant: "wan", Status: model.StatusChanged})
	if err != nil || len(rows) != 1 || rows[0].Roster[0] != "r-2" {
		t.Fatalf("search returned %#v: %v", rows, err)
	}
	if _, err = service.PublishReport(record.ID, "operator"); err != nil {
		t.Fatal(err)
	}
	if _, err = service.ConfirmRoster(record.ID, 0, "r-2", "operator"); err != nil {
		t.Fatal(err)
	}
}

func TestWorkflowImportReport(t *testing.T) {
	service, store := newTestService()
	result, err := service.Import([]model.ImportRow{{ApplicationNo: "A-300", Applicant: "Chen", Facility: "West", Species: "zebra", CageCount: 1, Roster: []string{"z-1"}}, {ApplicationNo: "A-301", Applicant: "Zhao", Facility: "West", Species: "rabbit", CageCount: 1, Roster: []string{"b-1"}}}, "importer")
	if err != nil || result.Imported != 2 {
		t.Fatalf("import result=%#v err=%v", result, err)
	}
	report, err := service.Report(result.Records[0].ID)
	if err != nil || len(report.Attachments) != 1 {
		t.Fatalf("report=%#v err=%v", report, err)
	}
	if len(store.attachments[result.Records[0].ID][0].Content) == 0 {
		t.Fatal("import attachment is empty")
	}
}

func TestPersistenceSurvivesReopen(t *testing.T) {
	service, store := newTestService()
	record, err := service.Register(RegistrationInput{ApplicationNo: "A-400", Applicant: "Liu", Facility: "South", Species: "mouse", CageCount: 1, Roster: []string{"m-1"}}, "registrar")
	if err != nil {
		t.Fatal(err)
	}
	reopened := NewService(store, FixedClock(time.Date(2026, 8, 25, 8, 0, 0, 0, time.UTC)))
	loaded, err := reopened.Get(record.ID)
	if err != nil || loaded.ApplicationNo != "A-400" {
		t.Fatalf("reopen load=%#v err=%v", loaded, err)
	}
}
