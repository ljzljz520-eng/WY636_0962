package flow004

import (
	"testing"
	"time"

	"example.com/animalcage/internal/model"
	"example.com/animalcage/internal/workflow"
)

type regressionStore struct{ records map[string]model.Record }

func newRegressionStore() *regressionStore {
	return &regressionStore{records: map[string]model.Record{}}
}
func (m *regressionStore) SaveRecord(r model.Record) error {
	m.records[r.ID] = model.CopyRecord(r)
	return nil
}
func (m *regressionStore) GetRecord(id string) (model.Record, bool, error) {
	r, ok := m.records[id]
	return model.CopyRecord(r), ok, nil
}
func (m *regressionStore) ListRecords() ([]model.Record, error) {
	out := make([]model.Record, 0, len(m.records))
	for _, r := range m.records {
		out = append(out, model.CopyRecord(r))
	}
	return out, nil
}
func (m *regressionStore) SaveAudit(model.AuditEvent) error                   { return nil }
func (m *regressionStore) ListAudit(string) ([]model.AuditEvent, error)       { return nil, nil }
func (m *regressionStore) SaveWorkflow(model.Workflow) error                  { return nil }
func (m *regressionStore) ListWorkflows(string) ([]model.Workflow, error)     { return nil, nil }
func (m *regressionStore) SaveAttachment(model.Attachment) error              { return nil }
func (m *regressionStore) ListAttachments(string) ([]model.Attachment, error) { return nil, nil }

func Test636BusinessRegression(t *testing.T) {
	service := workflow.NewService(newRegressionStore(), workflow.FixedClock(time.Date(2026, 8, 25, 8, 0, 0, 0, time.UTC)))
	record, err := service.Register(workflow.RegistrationInput{ApplicationNo: "BUG-636", Applicant: "Test", Facility: "Lab", Species: "mouse", CageCount: 2, Roster: []string{"mouse-a", "mouse-b"}}, "registrar")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.Submit(record.ID, "registrar"); err != nil {
		t.Fatal(err)
	}
	if _, err = service.Review(record.ID, model.ReviewDecision{Approved: true, Reviewer: "reviewer"}); err != nil {
		t.Fatal(err)
	}
	updated, err := service.ConfirmRoster(record.ID, 1, "mouse-b", "operator")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Roster[1] != "mouse-b" {
		t.Fatalf("second roster confirmation = %q, want latest name", updated.Roster[1])
	}
}
