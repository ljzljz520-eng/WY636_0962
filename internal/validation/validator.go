package validation

import (
	"example.com/animalcage/internal/model"
	"fmt"
	"strings"
)

type Issue struct {
	Field   string
	Message string
}

func ValidateRecord(record model.Record) []Issue {
	issues := make([]Issue, 0, 6)
	if strings.TrimSpace(record.ID) == "" {
		issues = append(issues, Issue{"id", "id is required"})
	}
	if strings.TrimSpace(record.ApplicationNo) == "" {
		issues = append(issues, Issue{"application_no", "application number is required"})
	}
	if strings.TrimSpace(record.Applicant) == "" {
		issues = append(issues, Issue{"applicant", "applicant is required"})
	}
	if strings.TrimSpace(record.Facility) == "" {
		issues = append(issues, Issue{"facility", "facility is required"})
	}
	if strings.TrimSpace(record.Species) == "" {
		issues = append(issues, Issue{"species", "species is required"})
	}
	if record.CageCount < 1 {
		issues = append(issues, Issue{"cage_count", "cage count must be positive"})
	}
	if len(record.Roster) == 0 {
		issues = append(issues, Issue{"roster", "roster cannot be empty"})
	}
	if len(record.Roster) > record.CageCount*20 {
		issues = append(issues, Issue{"roster", "roster exceeds cage capacity"})
	}
	return issues
}

func ValidateChange(record model.Record, request model.ChangeRequest) error {
	if record.Status != model.StatusApproved && record.Status != model.StatusChanged {
		return fmt.Errorf("record %s cannot be changed from %s", record.ID, record.Status)
	}
	if request.NewCageCount < 1 {
		return fmt.Errorf("cage count must be positive")
	}
	if len(request.NewRoster) == 0 {
		return fmt.Errorf("new roster cannot be empty")
	}
	if len(request.NewRoster) > request.NewCageCount*20 {
		return fmt.Errorf("new roster exceeds cage capacity")
	}
	if strings.TrimSpace(request.Reason) == "" {
		return fmt.Errorf("change reason is required")
	}
	return nil
}

func ValidateImport(rows []model.ImportRow) []Issue {
	issues := make([]Issue, 0)
	seen := make(map[string]bool, len(rows))
	for i, row := range rows {
		if seen[row.ApplicationNo] {
			issues = append(issues, Issue{fmt.Sprintf("rows[%d]", i), "duplicate application number"})
		}
		seen[row.ApplicationNo] = true
		candidate := model.Record{ID: row.ApplicationNo, ApplicationNo: row.ApplicationNo, Applicant: row.Applicant, Facility: row.Facility, Species: row.Species, CageCount: row.CageCount, Roster: model.NormalizeRoster(row.Roster)}
		for _, issue := range ValidateRecord(candidate) {
			issues = append(issues, Issue{fmt.Sprintf("rows[%d].%s", i, issue.Field), issue.Message})
		}
	}
	return issues
}

func IsTerminal(status string) bool { return status == model.StatusArchived }

func IsReviewable(status string) bool { return status == model.StatusSubmitted }
