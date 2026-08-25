package workflow

import (
	"fmt"
	"strings"

	"example.com/animalcage/internal/model"
	"example.com/animalcage/internal/validation"
)

// Change applies a documented roster or cage-count amendment to an approved record.
func (s *Service) Change(id string, request model.ChangeRequest) (model.Record, error) {
	record, err := s.read(id)
	if err != nil {
		return model.Record{}, err
	}
	request.NewRoster = model.NormalizeRoster(request.NewRoster)
	if err := validation.ValidateChange(record, request); err != nil {
		return model.Record{}, err
	}
	if strings.TrimSpace(request.Actor) == "" {
		return model.Record{}, fmt.Errorf("change actor is required")
	}
	record.Roster = append([]string(nil), request.NewRoster...)
	record.CageCount = request.NewCageCount
	record.Status = model.StatusChanged
	if err := s.write(record); err != nil {
		return model.Record{}, err
	}
	if err := s.appendAudit(record.ID, "change", request.Actor, strings.TrimSpace(request.Reason)); err != nil {
		return model.Record{}, err
	}
	if err := s.setWorkflow(record.ID, model.StageConfirmation, request.Actor, "amendment awaiting confirmation", false); err != nil {
		return model.Record{}, err
	}
	return record, nil
}

// Archive closes an application and makes it read-only.
func (s *Service) Archive(id, actor string) (model.Record, error) {
	record, err := s.read(id)
	if err != nil {
		return model.Record{}, err
	}
	if !model.CanTransition(record.Status, model.StatusArchived) {
		return model.Record{}, fmt.Errorf("record %s cannot be archived from %s", id, record.Status)
	}
	if strings.TrimSpace(actor) == "" {
		return model.Record{}, fmt.Errorf("archive actor is required")
	}
	when := s.now()
	record.Status, record.ArchivedAt = model.StatusArchived, &when
	if err := s.write(record); err != nil {
		return model.Record{}, err
	}
	if err := s.appendAudit(record.ID, "archive", actor, "application archived"); err != nil {
		return model.Record{}, err
	}
	return record, s.setWorkflow(record.ID, model.StageArchive, actor, "archive completed", true)
}
