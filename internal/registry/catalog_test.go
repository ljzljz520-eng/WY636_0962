package registry

import "testing"

func TestCatalog(t *testing.T) {
	c := NewCatalog()
	if err := c.AddFacility(Facility{Code: "F1", Name: "North", Capacity: 4}); err != nil {
		t.Fatal(err)
	}
	if _, ok := c.Facility("F1"); !ok {
		t.Fatal("facility missing")
	}
	if err := c.AddSpecies(Species{Code: "mouse", Name: "Mouse", RequiresApproval: true}); err != nil {
		t.Fatal(err)
	}
}
