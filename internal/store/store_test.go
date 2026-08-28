package store

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"example.com/animalcage/internal/model"
)

func openTestStore(t *testing.T) (*Store, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "animals.db")
	s, err := Open(path)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	return s, path
}

func sampleRecord() model.Record {
	now := time.Date(2026, 8, 25, 10, 0, 0, 0, time.UTC)
	return model.Record{ID: "r-1", ApplicationNo: "APP-001", Applicant: "Lin", Facility: "North", Species: "mouse", CageCount: 2, Roster: []string{"M01", "M02"}, Status: model.StatusSubmitted, Version: 1, CreatedAt: now, UpdatedAt: now}
}

func TestPersistenceSurvivesReopen(t *testing.T) {
	s, path := openTestStore(t)
	record := sampleRecord()
	event := model.AuditEvent{ID: "e-1", RecordID: record.ID, Action: "submitted", Actor: "lin", Detail: "initial", At: record.CreatedAt}
	workflow := model.Workflow{ID: "w-1", RecordID: record.ID, Stage: model.StageReview, Owner: "reviewer", DueDate: "2026-08-26"}
	attachment := model.Attachment{ID: "a-1", RecordID: record.ID, Name: "protocol.pdf", MediaType: "application/pdf", Content: []byte("fixture"), Checksum: "sha256:fixture"}
	for _, err := range []error{s.SaveRecord(record), s.SaveAuditEvent(event), s.SaveWorkflow(workflow), s.SaveAttachment(attachment)} {
		if err != nil {
			t.Fatalf("save: %v", err)
		}
	}
	if err := s.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	s, err := Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer s.Close()
	got, found, err := s.GetRecord(record.ID)
	if err != nil || !found || got.ApplicationNo != record.ApplicationNo || len(got.Roster) != 2 {
		t.Fatalf("reopened record mismatch: %#v, %v", got, err)
	}
	events, err := s.ListAuditEventsForRecord(record.ID)
	if err != nil || len(events) != 1 || events[0].Action != event.Action {
		t.Fatalf("reopened events mismatch: %#v, %v", events, err)
	}
	flows, err := s.ListWorkflowsForRecord(record.ID)
	if err != nil || len(flows) != 1 || flows[0].Stage != workflow.Stage {
		t.Fatalf("reopened workflows mismatch: %#v, %v", flows, err)
	}
	files, err := s.ListAttachmentsForRecord(record.ID)
	if err != nil || len(files) != 1 || string(files[0].Content) != "fixture" {
		t.Fatalf("reopened attachments mismatch: %#v, %v", files, err)
	}
}

func TestRecordListAndFilter(t *testing.T) {
	s, _ := openTestStore(t)
	defer s.Close()
	first := sampleRecord()
	second := first
	second.ID = "r-2"
	second.ApplicationNo = "APP-002"
	second.Applicant = "Wang"
	second.Facility = "South"
	second.Status = model.StatusApproved
	if err := s.SaveRecord(first); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveRecord(second); err != nil {
		t.Fatal(err)
	}
	got, err := s.FindRecords(model.SearchFilter{Applicant: "lin", Facility: "north"})
	if err != nil || len(got) != 1 || got[0].ID != first.ID {
		t.Fatalf("filter mismatch: %#v, %v", got, err)
	}
	all, err := s.ListRecords()
	if err != nil || len(all) != 2 || all[0].ID != "r-1" {
		t.Fatalf("list mismatch: %#v, %v", all, err)
	}
}

func TestRecordDeleteAndMissing(t *testing.T) {
	s, _ := openTestStore(t)
	defer s.Close()
	if err := s.SaveRecord(sampleRecord()); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteRecord("r-1"); err != nil {
		t.Fatal(err)
	}
	if _, found, err := s.GetRecord("r-1"); err != nil || found {
		t.Fatalf("expected not found, got found=%v err=%v", found, err)
	}
	if err := s.DeleteRecord("r-1"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected delete not found, got %v", err)
	}
}

func TestAppendEventAndOrdering(t *testing.T) {
	s, _ := openTestStore(t)
	defer s.Close()
	base := sampleRecord().CreatedAt
	for _, event := range []model.AuditEvent{
		{ID: "e-2", RecordID: "r-1", Action: "reviewed", At: base.Add(2 * time.Minute)},
		{ID: "e-1", RecordID: "r-1", Action: "submitted", At: base},
		{ID: "e-3", RecordID: "other", Action: "submitted", At: base},
	} {
		if err := s.AppendEvent(event); err != nil {
			t.Fatal(err)
		}
	}
	events, err := s.ListEvents()
	if err != nil || len(events) != 3 || events[0].ID != "e-1" || events[2].ID != "e-2" {
		t.Fatalf("event ordering mismatch: %#v, %v", events, err)
	}
	filtered, err := s.ListAuditEventsForRecord("r-1")
	if err != nil || len(filtered) != 2 || filtered[1].ID != "e-2" {
		t.Fatalf("record events mismatch: %#v, %v", filtered, err)
	}
}

func TestWorkflowAndAttachmentLifecycle(t *testing.T) {
	s, _ := openTestStore(t)
	defer s.Close()
	if err := s.SaveWorkflow(model.Workflow{ID: "w-1", RecordID: "r-1", Stage: model.StageRegistration}); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveWorkflow(model.Workflow{ID: "w-2", RecordID: "r-1", Stage: model.StageReview}); err != nil {
		t.Fatal(err)
	}
	flows, err := s.ListWorkflowsForRecord("r-1")
	if err != nil || len(flows) != 2 || flows[0].ID != "w-1" {
		t.Fatalf("workflow list mismatch: %#v, %v", flows, err)
	}
	if err := s.DeleteWorkflow("w-1"); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveAttachment(model.Attachment{ID: "a-1", RecordID: "r-1", Content: []byte("bytes")}); err != nil {
		t.Fatal(err)
	}
	attachment, err := s.GetAttachment("a-1")
	if err != nil || string(attachment.Content) != "bytes" {
		t.Fatalf("attachment mismatch: %#v, %v", attachment, err)
	}
	if err := s.DeleteAttachment("a-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetAttachment("a-1"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected attachment missing, got %v", err)
	}
}

func TestStoreValidationAndClose(t *testing.T) {
	if _, err := Open(""); err == nil {
		t.Fatal("expected empty path error")
	}
	s, _ := openTestStore(t)
	if err := s.SaveRecord(model.Record{}); !errors.Is(err, ErrInvalidID) {
		t.Fatalf("expected invalid ID, got %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.GetRecord("r-1"); err == nil {
		t.Fatal("expected closed store error")
	}
}
