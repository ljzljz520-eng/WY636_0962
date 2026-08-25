package workflow

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"example.com/animalcage/internal/model"
	"example.com/animalcage/internal/validation"
)

// Store is the persistence boundary used by the application service. The
// bbolt adapter implements this contract while tests can use a deterministic
// in-memory implementation.
type Store interface {
	SaveRecord(model.Record) error
	GetRecord(string) (model.Record, bool, error)
	ListRecords() ([]model.Record, error)
	SaveAudit(model.AuditEvent) error
	ListAudit(string) ([]model.AuditEvent, error)
	SaveWorkflow(model.Workflow) error
	ListWorkflows(string) ([]model.Workflow, error)
	SaveAttachment(model.Attachment) error
	ListAttachments(string) ([]model.Attachment, error)
}

type Clock interface{ Now() time.Time }

type FixedClock time.Time

func (c FixedClock) Now() time.Time { return time.Time(c) }

type RegistrationInput struct {
	ApplicationNo string
	Applicant     string
	Facility      string
	Species       string
	CageCount     int
	Roster        []string
}

type ImportResult struct {
	Records  []model.Record
	Issues   []validation.Issue
	Imported int
}

type Report struct {
	Record      model.Record
	Events      []model.AuditEvent
	Workflows   []model.Workflow
	Attachments []model.Attachment
}

type Service struct {
	store Store
	clock Clock
}

func NewService(store Store, clock Clock) *Service {
	if clock == nil {
		clock = FixedClock(time.Unix(0, 0).UTC())
	}
	return &Service{store: store, clock: clock}
}

func (s *Service) requireStore() error {
	if s == nil || s.store == nil {
		return errors.New("workflow store is required")
	}
	return nil
}

func (s *Service) now() time.Time { return s.clock.Now().UTC() }

func makeRecordID(applicationNo string) string {
	clean := strings.TrimSpace(applicationNo)
	if clean == "" {
		return "application-unnamed"
	}
	return "application-" + clean
}

func (s *Service) appendAudit(recordID, action, actor, detail string) error {
	previous, err := s.store.ListAudit(recordID)
	if err != nil {
		return err
	}
	event := model.AuditEvent{ID: fmt.Sprintf("%s-%s-%d-%d", recordID, action, s.now().UnixNano(), len(previous)), RecordID: recordID, Action: action, Actor: actor, Detail: detail, At: s.now()}
	return s.store.SaveAudit(event)
}

func (s *Service) setWorkflow(recordID, stage, owner, notes string, completed bool) error {
	flow := model.Workflow{ID: fmt.Sprintf("%s-%s", recordID, stage), RecordID: recordID, Stage: stage, Owner: owner, DueDate: s.now().Format("2006-01-02"), Completed: completed, Notes: notes}
	return s.store.SaveWorkflow(flow)
}

func (s *Service) read(id string) (model.Record, error) {
	if err := s.requireStore(); err != nil {
		return model.Record{}, err
	}
	record, ok, err := s.store.GetRecord(strings.TrimSpace(id))
	if err != nil {
		return model.Record{}, err
	}
	if !ok {
		return model.Record{}, fmt.Errorf("record %s not found", id)
	}
	return model.CopyRecord(record), nil
}

func (s *Service) write(record model.Record) error {
	record.UpdatedAt = s.now()
	record.Version++
	return s.store.SaveRecord(model.CopyRecord(record))
}

// Get returns a defensive copy of one application.
func (s *Service) Get(id string) (model.Record, error) { return s.read(id) }

// Search returns records matching all non-empty filter fields.
func (s *Service) Search(filter model.SearchFilter) ([]model.Record, error) {
	if err := s.requireStore(); err != nil {
		return nil, err
	}
	rows, err := s.store.ListRecords()
	if err != nil {
		return nil, err
	}
	result := make([]model.Record, 0, len(rows))
	for _, row := range rows {
		if filter.Applicant != "" && !strings.Contains(strings.ToLower(row.Applicant), strings.ToLower(filter.Applicant)) {
			continue
		}
		if filter.Facility != "" && row.Facility != filter.Facility {
			continue
		}
		if filter.Species != "" && row.Species != filter.Species {
			continue
		}
		if filter.Status != "" && row.Status != filter.Status {
			continue
		}
		result = append(result, model.CopyRecord(row))
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ApplicationNo < result[j].ApplicationNo })
	return result, nil
}
