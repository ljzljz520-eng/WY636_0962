package model

import "time"

type Record struct {
	ID            string     `json:"id"`
	ApplicationNo string     `json:"application_no"`
	Applicant     string     `json:"applicant"`
	Facility      string     `json:"facility"`
	Species       string     `json:"species"`
	CageCount     int        `json:"cage_count"`
	Roster        []string   `json:"roster"`
	Status        string     `json:"status"`
	Version       int        `json:"version"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	ArchivedAt    *time.Time `json:"archived_at,omitempty"`
}

type AuditEvent struct {
	ID       string    `json:"id"`
	RecordID string    `json:"record_id"`
	Action   string    `json:"action"`
	Actor    string    `json:"actor"`
	Detail   string    `json:"detail"`
	At       time.Time `json:"at"`
}

type Workflow struct {
	ID        string `json:"id"`
	RecordID  string `json:"record_id"`
	Stage     string `json:"stage"`
	Owner     string `json:"owner"`
	DueDate   string `json:"due_date"`
	Completed bool   `json:"completed"`
	Notes     string `json:"notes"`
}

type Attachment struct {
	ID        string `json:"id"`
	RecordID  string `json:"record_id"`
	Name      string `json:"name"`
	MediaType string `json:"media_type"`
	Content   []byte `json:"content"`
	Checksum  string `json:"checksum"`
}

type SearchFilter struct {
	Applicant string
	Facility  string
	Species   string
	Status    string
}

type ReviewDecision struct {
	Approved bool
	Reviewer string
	Comment  string
}

type ChangeRequest struct {
	Actor        string
	NewRoster    []string
	NewCageCount int
	Reason       string
}

type ImportRow struct {
	ApplicationNo string
	Applicant     string
	Facility      string
	Species       string
	CageCount     int
	Roster        []string
}
