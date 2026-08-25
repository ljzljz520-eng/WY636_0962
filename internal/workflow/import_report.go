package workflow

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"example.com/animalcage/internal/model"
	"example.com/animalcage/internal/validation"
)

// Import validates and registers a deterministic batch of applications.
func (s *Service) Import(rows []model.ImportRow, actor string) (ImportResult, error) {
	if err := s.requireStore(); err != nil {
		return ImportResult{}, err
	}
	issues := validation.ValidateImport(rows)
	if len(issues) > 0 {
		return ImportResult{Issues: issues}, fmt.Errorf("import validation failed: %s", validation.Summarize(issues))
	}
	result := ImportResult{Records: make([]model.Record, 0, len(rows))}
	for _, row := range rows {
		record, err := s.Register(RegistrationInput{ApplicationNo: row.ApplicationNo, Applicant: row.Applicant, Facility: row.Facility, Species: row.Species, CageCount: row.CageCount, Roster: row.Roster}, actor)
		if err != nil {
			return result, err
		}
		result.Records = append(result.Records, record)
	}
	result.Imported = len(result.Records)
	if len(result.Records) > 0 {
		data, _ := json.Marshal(result.Records)
		sum := sha256.Sum256(data)
		attachment := model.Attachment{ID: "import-" + hex.EncodeToString(sum[:8]), RecordID: result.Records[0].ID, Name: "import-report.json", MediaType: "application/json", Content: data, Checksum: hex.EncodeToString(sum[:])}
		if err := s.store.SaveAttachment(attachment); err != nil {
			return result, err
		}
	}
	return result, nil
}

// Report returns a complete read model for the application detail screen.
func (s *Service) Report(id string) (Report, error) {
	record, err := s.read(id)
	if err != nil {
		return Report{}, err
	}
	events, err := s.store.ListAudit(id)
	if err != nil {
		return Report{}, err
	}
	flows, err := s.store.ListWorkflows(id)
	if err != nil {
		return Report{}, err
	}
	attachments, err := s.store.ListAttachments(id)
	if err != nil {
		return Report{}, err
	}
	return Report{Record: record, Events: model.CopyEvents(events), Workflows: append([]model.Workflow(nil), flows...), Attachments: attachments}, nil
}

// PublishReport creates a stable JSON attachment that can be downloaded by a client.
func (s *Service) PublishReport(id, actor string) (model.Attachment, error) {
	report, err := s.Report(id)
	if err != nil {
		return model.Attachment{}, err
	}
	if strings.TrimSpace(actor) == "" {
		return model.Attachment{}, fmt.Errorf("report actor is required")
	}
	payload, err := json.Marshal(report)
	if err != nil {
		return model.Attachment{}, err
	}
	sum := sha256.Sum256(payload)
	attachment := model.Attachment{ID: "report-" + hex.EncodeToString(sum[:8]), RecordID: id, Name: "application-report.json", MediaType: "application/json", Content: payload, Checksum: hex.EncodeToString(sum[:])}
	if err := s.store.SaveAttachment(attachment); err != nil {
		return model.Attachment{}, err
	}
	if err := s.appendAudit(id, "publish_report", actor, "report published"); err != nil {
		return model.Attachment{}, err
	}
	return attachment, nil
}
