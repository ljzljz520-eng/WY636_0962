package workflow

import (
	"sort"
	"sync"

	"example.com/animalcage/internal/model"
)

type memoryStore struct {
	mu          sync.Mutex
	records     map[string]model.Record
	events      map[string][]model.AuditEvent
	flows       map[string][]model.Workflow
	attachments map[string][]model.Attachment
}

func newMemoryStore() *memoryStore {
	return &memoryStore{records: map[string]model.Record{}, events: map[string][]model.AuditEvent{}, flows: map[string][]model.Workflow{}, attachments: map[string][]model.Attachment{}}
}
func (m *memoryStore) SaveRecord(r model.Record) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.records[r.ID] = model.CopyRecord(r)
	return nil
}
func (m *memoryStore) GetRecord(id string) (model.Record, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.records[id]
	if !ok {
		return model.Record{}, false, nil
	}
	return model.CopyRecord(r), true, nil
}
func (m *memoryStore) ListRecords() ([]model.Record, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]model.Record, 0, len(m.records))
	for _, r := range m.records {
		out = append(out, model.CopyRecord(r))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
func (m *memoryStore) SaveAudit(e model.AuditEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events[e.RecordID] = append(m.events[e.RecordID], e)
	return nil
}
func (m *memoryStore) ListAudit(id string) ([]model.AuditEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return model.CopyEvents(m.events[id]), nil
}
func (m *memoryStore) SaveWorkflow(w model.Workflow) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.flows[w.RecordID] = append(m.flows[w.RecordID], w)
	return nil
}
func (m *memoryStore) ListWorkflows(id string) ([]model.Workflow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]model.Workflow(nil), m.flows[id]...), nil
}
func (m *memoryStore) SaveAttachment(a model.Attachment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.attachments[a.RecordID] = append(m.attachments[a.RecordID], a)
	return nil
}
func (m *memoryStore) ListAttachments(id string) ([]model.Attachment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]model.Attachment(nil), m.attachments[id]...), nil
}
