package report

import (
	"encoding/csv"
	"encoding/json"
	"example.com/animalcage/internal/model"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

type Summary struct {
	GeneratedAt time.Time      `json:"generated_at"`
	Total       int            `json:"total"`
	ByStatus    map[string]int `json:"by_status"`
	ByFacility  map[string]int `json:"by_facility"`
	CageTotal   int            `json:"cage_total"`
}

func BuildSummary(records []model.Record, now time.Time) Summary {
	s := Summary{GeneratedAt: now, ByStatus: map[string]int{}, ByFacility: map[string]int{}}
	for _, r := range records {
		s.Total++
		s.CageTotal += r.CageCount
		s.ByStatus[r.Status]++
		s.ByFacility[r.Facility]++
	}
	return s
}
func EncodeJSON(summary Summary) ([]byte, error) { return json.MarshalIndent(summary, "", "  ") }
func EncodeCSV(records []model.Record, w io.Writer) error {
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"id", "application_no", "applicant", "facility", "species", "cage_count", "status", "version"}); err != nil {
		return err
	}
	sorted := append([]model.Record(nil), records...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].ApplicationNo < sorted[j].ApplicationNo })
	for _, r := range sorted {
		if err := cw.Write([]string{r.ID, r.ApplicationNo, r.Applicant, r.Facility, r.Species, fmt.Sprint(r.CageCount), r.Status, fmt.Sprint(r.Version)}); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}
func StatusLine(summary Summary) string {
	parts := make([]string, 0, len(summary.ByStatus))
	for k, v := range summary.ByStatus {
		parts = append(parts, fmt.Sprintf("%s=%d", k, v))
	}
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}
