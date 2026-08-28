package workflow

import (
	"fmt"
	"strings"
	"time"

	"example.com/animalcage/internal/model"
	"example.com/animalcage/internal/planning"
	"example.com/animalcage/internal/validation"
)

// Register creates a draft application and records its registration stage.
func (s *Service) Register(input RegistrationInput, actor string) (model.Record, error) {
	if err := s.requireStore(); err != nil {
		return model.Record{}, err
	}
	roster := model.NormalizeRoster(input.Roster)
	record := model.Record{ID: makeRecordID(input.ApplicationNo), ApplicationNo: strings.TrimSpace(input.ApplicationNo), Applicant: strings.TrimSpace(input.Applicant), Facility: strings.TrimSpace(input.Facility), Species: strings.TrimSpace(input.Species), CageCount: input.CageCount, Roster: roster, Status: model.StatusDraft, Version: 0, CreatedAt: s.now(), UpdatedAt: s.now()}
	if issues := validation.ValidateRecord(record); len(issues) > 0 {
		return model.Record{}, fmt.Errorf("invalid registration: %s", validation.Summarize(issues))
	}
	if err := planning.ValidateRequest(planning.Request{ApplicationID: record.ID, Facility: record.Facility, Species: record.Species, CageCount: record.CageCount, Start: record.CreatedAt, End: 24 * time.Hour}); err != nil {
		return model.Record{}, fmt.Errorf("invalid allocation request: %w", err)
	}
	if _, ok, err := s.store.GetRecord(record.ID); err != nil {
		return model.Record{}, err
	} else if ok {
		return model.Record{}, fmt.Errorf("record %s already exists", record.ID)
	}
	if err := s.store.SaveRecord(model.CopyRecord(record)); err != nil {
		return model.Record{}, err
	}
	if err := s.appendAudit(record.ID, "register", actor, "application registered"); err != nil {
		return model.Record{}, err
	}
	if err := s.setWorkflow(record.ID, model.StageRegistration, actor, "registration created", true); err != nil {
		return model.Record{}, err
	}
	return model.CopyRecord(record), nil
}

// Submit moves a draft to the review queue.
func (s *Service) Submit(id, actor string) (model.Record, error) {
	record, err := s.read(id)
	if err != nil {
		return model.Record{}, err
	}
	if !model.CanTransition(record.Status, model.StatusSubmitted) {
		return model.Record{}, fmt.Errorf("record %s cannot be submitted from %s", id, record.Status)
	}
	record.Status = model.StatusSubmitted
	if err := s.write(record); err != nil {
		return model.Record{}, err
	}
	if err := s.appendAudit(record.ID, "submit", actor, "application submitted for review"); err != nil {
		return model.Record{}, err
	}
	if err := s.setWorkflow(record.ID, model.StageReview, actor, "awaiting review", false); err != nil {
		return model.Record{}, err
	}
	return record, nil
}
