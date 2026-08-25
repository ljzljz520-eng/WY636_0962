package store

import (
	"strings"

	"example.com/animalcage/internal/model"
)

// FindRecords applies a deterministic, case-insensitive filter to all records.
// Empty fields do not constrain the result.
func (s *Store) FindRecords(filter model.SearchFilter) ([]model.Record, error) {
	records, err := s.ListRecords()
	if err != nil {
		return nil, err
	}
	applicant := strings.ToLower(strings.TrimSpace(filter.Applicant))
	facility := strings.ToLower(strings.TrimSpace(filter.Facility))
	species := strings.ToLower(strings.TrimSpace(filter.Species))
	status := strings.ToLower(strings.TrimSpace(filter.Status))
	result := make([]model.Record, 0, len(records))
	for _, record := range records {
		if applicant != "" && !strings.Contains(strings.ToLower(record.Applicant), applicant) {
			continue
		}
		if facility != "" && !strings.EqualFold(record.Facility, facility) {
			continue
		}
		if species != "" && !strings.EqualFold(record.Species, species) {
			continue
		}
		if status != "" && !strings.EqualFold(record.Status, status) {
			continue
		}
		result = append(result, record)
	}
	return result, nil
}
