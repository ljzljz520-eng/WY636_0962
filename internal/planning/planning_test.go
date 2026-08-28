package planning

import (
	"testing"
	"time"
)

func TestBuildAllocationAndReserve(t *testing.T) {
	cages := []Cage{NewCage("c2", "F", "mouse", "02"), NewCage("c1", "F", "mouse", "01")}
	req := Request{ApplicationID: "A", Facility: "F", Species: "mouse", CageCount: 1, Start: time.Unix(100, 0), End: time.Hour}
	allocation, err := BuildAllocation(req, cages)
	if err != nil {
		t.Fatal(err)
	}
	if allocation.Cages[0].ID != "c1" {
		t.Fatalf("stable selection got %s", allocation.Cages[0].ID)
	}
	slots, err := Reserve(nil, allocation)
	if err != nil || len(slots) != 1 {
		t.Fatal(err)
	}
	if len(CheckConflicts(slots, slots[0])) != 1 {
		t.Fatal("expected conflict")
	}
}

func TestPolicyAndCalendar(t *testing.T) {
	policy := Policy{Species: "mouse", MaxCages: 2, MinDuration: time.Hour, AllowedFacilities: []string{"F"}}
	request := Request{ApplicationID: "A", Facility: "F", Species: "mouse", CageCount: 1, Start: time.Unix(100, 0), End: 2 * time.Hour}
	if !policy.Evaluate(request).Allowed {
		t.Fatal("policy denied valid request")
	}
	calendar := NewCalendar("lab", "UTC")
	if err := calendar.SetWindow(DayWindow{Weekday: time.Thursday, Open: 0, Close: 24 * time.Hour}); err != nil {
		t.Fatal(err)
	}
}
