package report

import (
	"example.com/animalcage/internal/model"
	"strings"
)

func Filter(records []model.Record, filter model.SearchFilter) []model.Record {
	out := make([]model.Record, 0)
	for _, r := range records {
		if filter.Applicant != "" && !strings.Contains(strings.ToLower(r.Applicant), strings.ToLower(filter.Applicant)) {
			continue
		}
		if filter.Facility != "" && r.Facility != filter.Facility {
			continue
		}
		if filter.Species != "" && r.Species != filter.Species {
			continue
		}
		if filter.Status != "" && r.Status != filter.Status {
			continue
		}
		out = append(out, model.CopyRecord(r))
	}
	return out
}
func CountBySpecies(records []model.Record) map[string]int {
	out := make(map[string]int)
	for _, r := range records {
		out[r.Species]++
	}
	return out
}
