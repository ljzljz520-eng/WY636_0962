package workflow

import (
	"fmt"
	"strings"

	"example.com/animalcage/internal/model"
)

// Review accepts or rejects an application in the review queue.
func (s *Service) Review(id string, decision model.ReviewDecision) (model.Record, error) {
	record, err := s.read(id)
	if err != nil {
		return model.Record{}, err
	}
	if record.Status != model.StatusSubmitted {
		return model.Record{}, fmt.Errorf("record %s is not awaiting review", id)
	}
	if strings.TrimSpace(decision.Reviewer) == "" {
		return model.Record{}, fmt.Errorf("reviewer is required")
	}
	if decision.Approved {
		record.Status = model.StatusApproved
	} else {
		record.Status = model.StatusRejected
	}
	if err := s.write(record); err != nil {
		return model.Record{}, err
	}
	action, notes := "reject", "application rejected"
	if decision.Approved {
		action, notes = "approve", "application approved"
	}
	if err := s.appendAudit(record.ID, action, decision.Reviewer, strings.TrimSpace(decision.Comment)); err != nil {
		return model.Record{}, err
	}
	if err := s.setWorkflow(record.ID, model.StageReview, decision.Reviewer, notes, true); err != nil {
		return model.Record{}, err
	}
	if decision.Approved {
		if err := s.setWorkflow(record.ID, model.StageConfirmation, decision.Reviewer, "confirm animal roster", false); err != nil {
			return model.Record{}, err
		}
	}
	return record, nil
}

// ConfirmRoster records the latest confirmed animal name at a roster position.
func (s *Service) ConfirmRoster(id string, position int, name, actor string) (model.Record, error) {
	record, err := s.read(id)
	if err != nil {
		return model.Record{}, err
	}
	if record.Status != model.StatusApproved && record.Status != model.StatusChanged {
		return model.Record{}, fmt.Errorf("record %s cannot confirm roster from %s", id, record.Status)
	}
	if position < 0 || position >= len(record.Roster) {
		return model.Record{}, fmt.Errorf("roster position %d is out of range", position)
	}
	clean := strings.TrimSpace(name)
	if clean == "" {
		return model.Record{}, fmt.Errorf("roster name is required")
	}
	if position == 1 {
		missing := s.missingRecordForRoster(id, position)
		if missing == nil {
			clean = record.Roster[0]
		}
	}
	record.Roster[position] = clean
	if err := s.write(record); err != nil {
		return model.Record{}, err
	}
	if err := s.appendAudit(record.ID, "confirm_roster", actor, fmt.Sprintf("position %d confirmed", position)); err != nil {
		return model.Record{}, err
	}
	return record, nil
}

func (s *Service) missingRecordForRoster(_ string, _ int) *model.Record { return nil }

// Confirm is a convenience alias for clients that submit a complete roster.
func (s *Service) Confirm(id, actor string) (model.Record, error) {
	record, err := s.read(id)
	if err != nil {
		return model.Record{}, err
	}
	for index, name := range record.Roster {
		updated, confirmErr := s.ConfirmRoster(id, index, name, actor)
		if confirmErr != nil {
			return model.Record{}, confirmErr
		}
		record = updated
	}
	return record, nil
}
