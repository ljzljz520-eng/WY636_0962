package planning

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type Policy struct {
	Species                  string
	MaxCages                 int
	MinDuration              time.Duration
	MaxDuration              time.Duration
	RequiresVeterinaryReview bool
	AllowedFacilities        []string
}
type Decision struct {
	Allowed bool
	Code    string
	Message string
	Checks  []string
}
type Reviewer struct {
	ID     string
	Name   string
	Roles  []string
	Active bool
}

func (p Policy) Evaluate(request Request) Decision {
	decision := Decision{Allowed: true, Code: "allowed", Checks: make([]string, 0, 5)}
	if p.Species != "" && p.Species != request.Species {
		return Decision{Code: "species_mismatch", Message: "species is not covered by policy", Checks: []string{"species"}}
	}
	if request.CageCount > p.MaxCages && p.MaxCages > 0 {
		return Decision{Code: "capacity_limit", Message: "request exceeds policy cage limit", Checks: []string{"cage_count"}}
	}
	if p.MinDuration > 0 && request.End < p.MinDuration {
		return Decision{Code: "duration_short", Message: "request duration is below minimum", Checks: []string{"duration"}}
	}
	if p.MaxDuration > 0 && request.End > p.MaxDuration {
		return Decision{Code: "duration_long", Message: "request duration is above maximum", Checks: []string{"duration"}}
	}
	if len(p.AllowedFacilities) > 0 && !contains(p.AllowedFacilities, request.Facility) {
		return Decision{Code: "facility_denied", Message: "facility is not allowed by policy", Checks: []string{"facility"}}
	}
	decision.Checks = append(decision.Checks, "species", "capacity", "duration", "facility")
	if p.RequiresVeterinaryReview {
		decision.Checks = append(decision.Checks, "veterinary_review")
	}
	return decision
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func ReviewQueue(policies []Policy, requests []Request) []Request {
	queue := make([]Request, 0)
	for _, request := range requests {
		for _, policy := range policies {
			if policy.Evaluate(request).Allowed && policy.RequiresVeterinaryReview {
				queue = append(queue, request)
				break
			}
		}
	}
	sort.Slice(queue, func(i, j int) bool {
		if queue[i].Start.Equal(queue[j].Start) {
			return queue[i].ApplicationID < queue[j].ApplicationID
		}
		return queue[i].Start.Before(queue[j].Start)
	})
	return queue
}

func ValidateReviewer(reviewer Reviewer, requiredRole string) error {
	if strings.TrimSpace(reviewer.ID) == "" || strings.TrimSpace(reviewer.Name) == "" {
		return fmt.Errorf("reviewer identity is required")
	}
	if !reviewer.Active {
		return fmt.Errorf("reviewer is inactive")
	}
	if requiredRole != "" && !contains(reviewer.Roles, requiredRole) {
		return fmt.Errorf("reviewer lacks role %s", requiredRole)
	}
	return nil
}

func CanApprove(reviewer Reviewer, request Request, policy Policy) Decision {
	if err := ValidateReviewer(reviewer, "reviewer"); err != nil {
		return Decision{Code: "reviewer_invalid", Message: err.Error()}
	}
	decision := policy.Evaluate(request)
	if !decision.Allowed {
		return decision
	}
	decision.Checks = append(decision.Checks, "reviewer_authorized")
	return decision
}

func Explain(decision Decision) string {
	if decision.Allowed {
		return "allowed: " + strings.Join(decision.Checks, ",")
	}
	return decision.Code + ": " + decision.Message
}
