package planning

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type Plan struct {
	Allocation   Allocation   `json:"allocation"`
	Policy       Decision     `json:"policy"`
	Availability Availability `json:"availability"`
	GeneratedBy  string       `json:"generated_by"`
}

func EncodePlan(plan Plan) ([]byte, error) {
	if strings.TrimSpace(plan.GeneratedBy) == "" {
		return nil, fmt.Errorf("generated_by is required")
	}
	return json.MarshalIndent(plan, "", "  ")
}
func DecodePlan(data []byte) (Plan, error) {
	var plan Plan
	if err := json.Unmarshal(data, &plan); err != nil {
		return Plan{}, err
	}
	if strings.TrimSpace(plan.GeneratedBy) == "" {
		return Plan{}, fmt.Errorf("generated_by is required")
	}
	return plan, nil
}
func SortAllocations(values []Allocation) []Allocation {
	out := append([]Allocation(nil), values...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Start.Equal(out[j].Start) {
			return out[i].ApplicationID < out[j].ApplicationID
		}
		return out[i].Start.Before(out[j].Start)
	})
	return out
}
func MergeNotes(allocation Allocation, notes ...string) Allocation {
	copy := allocation
	copy.Notes = append([]string(nil), allocation.Notes...)
	for _, note := range notes {
		clean := strings.TrimSpace(note)
		if clean != "" {
			copy.Notes = append(copy.Notes, clean)
		}
	}
	return copy
}
func HasCage(allocation Allocation, cageID string) bool {
	for _, cage := range allocation.Cages {
		if cage.ID == cageID {
			return true
		}
	}
	return false
}
