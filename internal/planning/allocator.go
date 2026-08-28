package planning

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type Cage struct {
	ID       string
	Facility string
	Species  string
	Label    string
	Active   bool
}
type Slot struct {
	CageID        string
	Start         time.Time
	End           time.Time
	Occupied      bool
	ApplicationID string
}
type Request struct {
	ApplicationID string
	Facility      string
	Species       string
	CageCount     int
	Start         time.Time
	End           time.Duration
}
type Allocation struct {
	ApplicationID string
	Facility      string
	Species       string
	Cages         []Cage
	Start         time.Time
	End           time.Time
	Notes         []string
}

func NewCage(id, facility, species, label string) Cage {
	return Cage{ID: strings.TrimSpace(id), Facility: strings.TrimSpace(facility), Species: strings.TrimSpace(species), Label: strings.TrimSpace(label), Active: true}
}

func ValidateRequest(r Request) error {
	if strings.TrimSpace(r.ApplicationID) == "" {
		return fmt.Errorf("application id is required")
	}
	if strings.TrimSpace(r.Facility) == "" {
		return fmt.Errorf("facility is required")
	}
	if strings.TrimSpace(r.Species) == "" {
		return fmt.Errorf("species is required")
	}
	if r.CageCount < 1 {
		return fmt.Errorf("cage count must be positive")
	}
	if r.Start.IsZero() {
		return fmt.Errorf("start time is required")
	}
	if r.End < 0 {
		return fmt.Errorf("duration cannot be negative")
	}
	return nil
}

func SelectCages(cages []Cage, facility, species string, count int) ([]Cage, error) {
	if count < 1 {
		return nil, fmt.Errorf("cage count must be positive")
	}
	filtered := make([]Cage, 0, len(cages))
	for _, cage := range cages {
		if !cage.Active || cage.Facility != facility || cage.Species != species {
			continue
		}
		filtered = append(filtered, cage)
	}
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Label == filtered[j].Label {
			return filtered[i].ID < filtered[j].ID
		}
		return filtered[i].Label < filtered[j].Label
	})
	if len(filtered) < count {
		return nil, fmt.Errorf("only %d cages are available", len(filtered))
	}
	return append([]Cage(nil), filtered[:count]...), nil
}

func BuildAllocation(req Request, cages []Cage) (Allocation, error) {
	if err := ValidateRequest(req); err != nil {
		return Allocation{}, err
	}
	chosen, err := SelectCages(cages, req.Facility, req.Species, req.CageCount)
	if err != nil {
		return Allocation{}, err
	}
	return Allocation{ApplicationID: req.ApplicationID, Facility: req.Facility, Species: req.Species, Cages: chosen, Start: req.Start.UTC(), End: req.Start.UTC().Add(req.End), Notes: []string{"capacity checked", "stable cage order"}}, nil
}

func AllocationKey(a Allocation) string {
	ids := make([]string, len(a.Cages))
	for i, cage := range a.Cages {
		ids[i] = cage.ID
	}
	sort.Strings(ids)
	return strings.Join([]string{a.ApplicationID, a.Facility, a.Species, strings.Join(ids, ","), a.Start.Format(time.RFC3339), a.End.Format(time.RFC3339)}, "|")
}

func IsOverlapping(a, b Slot) bool {
	if a.CageID != b.CageID {
		return false
	}
	return a.Start.Before(b.End) && b.Start.Before(a.End)
}

func CheckConflicts(slots []Slot, candidate Slot) []Slot {
	conflicts := make([]Slot, 0)
	for _, slot := range slots {
		if slot.Occupied && IsOverlapping(slot, candidate) {
			conflicts = append(conflicts, slot)
		}
	}
	sort.Slice(conflicts, func(i, j int) bool { return conflicts[i].Start.Before(conflicts[j].Start) })
	return conflicts
}

func Reserve(slots []Slot, allocation Allocation) ([]Slot, error) {
	result := append([]Slot(nil), slots...)
	for _, cage := range allocation.Cages {
		candidate := Slot{CageID: cage.ID, Start: allocation.Start, End: allocation.End, Occupied: true, ApplicationID: allocation.ApplicationID}
		if conflicts := CheckConflicts(result, candidate); len(conflicts) > 0 {
			return nil, fmt.Errorf("cage %s has %d conflicts", cage.ID, len(conflicts))
		}
		result = append(result, candidate)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].CageID == result[j].CageID {
			return result[i].Start.Before(result[j].Start)
		}
		return result[i].CageID < result[j].CageID
	})
	return result, nil
}

func Release(slots []Slot, applicationID string) []Slot {
	result := make([]Slot, 0, len(slots))
	for _, slot := range slots {
		if slot.ApplicationID != applicationID {
			result = append(result, slot)
		}
	}
	return result
}

func OccupiedCount(slots []Slot, at time.Time) int {
	total := 0
	for _, slot := range slots {
		if slot.Occupied && slot.Start.Before(at) && at.Before(slot.End) {
			total++
		}
	}
	return total
}
