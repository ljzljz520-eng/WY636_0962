package registry

import (
	"example.com/animalcage/internal/model"
	"fmt"
	"sort"
	"strings"
	"sync"
)

type Facility struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Capacity int    `json:"capacity"`
	Active   bool   `json:"active"`
}
type Species struct {
	Code             string `json:"code"`
	Name             string `json:"name"`
	RequiresApproval bool   `json:"requires_approval"`
}
type Catalog struct {
	mu         sync.RWMutex
	facilities map[string]Facility
	species    map[string]Species
}

func NewCatalog() *Catalog {
	return &Catalog{facilities: make(map[string]Facility), species: make(map[string]Species)}
}

func (c *Catalog) AddFacility(f Facility) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if strings.TrimSpace(f.Code) == "" || strings.TrimSpace(f.Name) == "" {
		return fmt.Errorf("facility code and name are required")
	}
	if f.Capacity < 1 {
		return fmt.Errorf("facility capacity must be positive")
	}
	if _, exists := c.facilities[f.Code]; exists {
		return fmt.Errorf("facility %s already exists", f.Code)
	}
	f.Active = true
	c.facilities[f.Code] = f
	return nil
}

func (c *Catalog) AddSpecies(s Species) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if strings.TrimSpace(s.Code) == "" || strings.TrimSpace(s.Name) == "" {
		return fmt.Errorf("species code and name are required")
	}
	if _, exists := c.species[s.Code]; exists {
		return fmt.Errorf("species %s already exists", s.Code)
	}
	c.species[s.Code] = s
	return nil
}

func (c *Catalog) Facility(code string) (Facility, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	f, ok := c.facilities[code]
	return f, ok
}
func (c *Catalog) Species(code string) (Species, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	s, ok := c.species[code]
	return s, ok
}
func (c *Catalog) ListFacilities() []Facility {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]Facility, 0, len(c.facilities))
	for _, f := range c.facilities {
		out = append(out, f)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	return out
}
func (c *Catalog) ListSpecies() []Species {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]Species, 0, len(c.species))
	for _, s := range c.species {
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	return out
}

func CheckCapacity(record model.Record, facility Facility) error {
	if !facility.Active {
		return fmt.Errorf("facility %s is inactive", facility.Code)
	}
	if record.CageCount > facility.Capacity {
		return fmt.Errorf("requested cages exceed facility capacity")
	}
	return nil
}
