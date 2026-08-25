package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"example.com/animalcage/internal/model"
	"example.com/animalcage/internal/workflow"
)

type memoryStore struct {
	records     map[string]model.Record
	events      []model.AuditEvent
	flows       []model.Workflow
	attachments []model.Attachment
}

func newMemoryStore() *memoryStore { return &memoryStore{records: map[string]model.Record{}} }
func (m *memoryStore) SaveRecord(v model.Record) error {
	m.records[v.ID] = model.CopyRecord(v)
	return nil
}
func (m *memoryStore) GetRecord(id string) (model.Record, bool, error) {
	v, ok := m.records[id]
	return model.CopyRecord(v), ok, nil
}
func (m *memoryStore) ListRecords() ([]model.Record, error) {
	out := make([]model.Record, 0, len(m.records))
	for _, v := range m.records {
		out = append(out, model.CopyRecord(v))
	}
	return out, nil
}
func (m *memoryStore) SaveAudit(v model.AuditEvent) error { m.events = append(m.events, v); return nil }
func (m *memoryStore) ListAudit(id string) ([]model.AuditEvent, error) {
	out := []model.AuditEvent{}
	for _, v := range m.events {
		if v.RecordID == id {
			out = append(out, v)
		}
	}
	return out, nil
}
func (m *memoryStore) SaveWorkflow(v model.Workflow) error { m.flows = append(m.flows, v); return nil }
func (m *memoryStore) ListWorkflows(id string) ([]model.Workflow, error) {
	out := []model.Workflow{}
	for _, v := range m.flows {
		if v.RecordID == id {
			out = append(out, v)
		}
	}
	return out, nil
}
func (m *memoryStore) SaveAttachment(v model.Attachment) error {
	m.attachments = append(m.attachments, v)
	return nil
}
func (m *memoryStore) ListAttachments(id string) ([]model.Attachment, error) {
	out := []model.Attachment{}
	for _, v := range m.attachments {
		if v.RecordID == id {
			out = append(out, v)
		}
	}
	return out, nil
}

func testHandler() *Handler {
	clock := workflow.FixedClock(time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC))
	return NewHandler(workflow.NewService(newMemoryStore(), clock))
}

func requestJSON(t *testing.T, handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	recording := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(recording, request)
	return recording
}

func TestRegisterAndSearchEndpoints(t *testing.T) {
	handler := testHandler()
	created := requestJSON(t, handler, http.MethodPost, "/api/applications", `{"application_no":"C-100","applicant":"Lab A","facility":"North","species":"mouse","cage_count":2,"roster":["M1","M2"],"actor":"alice"}`)
	if created.Code != http.StatusCreated {
		t.Fatalf("register status = %d, body=%s", created.Code, created.Body.String())
	}
	var record model.Record
	if err := json.Unmarshal(created.Body.Bytes(), &record); err != nil {
		t.Fatal(err)
	}
	if record.ID != "application-C-100" || record.Status != model.StatusDraft {
		t.Fatalf("unexpected record: %+v", record)
	}
	found := requestJSON(t, handler, http.MethodGet, "/api/applications?applicant=lab", "")
	if found.Code != http.StatusOK {
		t.Fatalf("search status = %d", found.Code)
	}
	var result struct {
		Applications []model.Record `json:"applications"`
		Count        int            `json:"count"`
	}
	if err := json.Unmarshal(found.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Count != 1 || len(result.Applications) != 1 {
		t.Fatalf("unexpected search result: %+v", result)
	}
}

func TestWorkflowEndpointsExposeLatestConfirmationInput(t *testing.T) {
	handler := testHandler()
	requestJSON(t, handler, http.MethodPost, "/api/applications", `{"application_no":"C-200","applicant":"Lab B","facility":"South","species":"rat","cage_count":2,"roster":["old-one","old-two"],"actor":"alice"}`)
	id := "/api/applications/application-C-200"
	requestJSON(t, handler, http.MethodPost, id+"/submit", `{"actor":"alice"}`)
	review := requestJSON(t, handler, http.MethodPost, id+"/review", `{"approved":true,"reviewer":"bob"}`)
	if review.Code != http.StatusOK {
		t.Fatalf("review status = %d", review.Code)
	}
	second := requestJSON(t, handler, http.MethodPost, id+"/confirm-roster", `{"position":1,"name":"latest-two","actor":"bob"}`)
	if second.Code != http.StatusOK {
		t.Fatalf("confirm status = %d, body=%s", second.Code, second.Body.String())
	}
	var record model.Record
	if err := json.Unmarshal(second.Body.Bytes(), &record); err != nil {
		t.Fatal(err)
	}
	if len(record.Roster) != 2 || record.Status != model.StatusApproved {
		t.Fatalf("unexpected confirmation response: %+v", record)
	}
}

func TestMethodAndPayloadErrorsAreJSON(t *testing.T) {
	handler := testHandler()
	method := requestJSON(t, handler, http.MethodDelete, "/api/applications", "")
	if method.Code != http.StatusMethodNotAllowed {
		t.Fatalf("method status = %d", method.Code)
	}
	bad := requestJSON(t, handler, http.MethodPost, "/api/applications", `{`)
	if bad.Code != http.StatusBadRequest || !strings.Contains(bad.Body.String(), "invalid JSON") {
		t.Fatalf("bad payload = %d %s", bad.Code, bad.Body.String())
	}
}
