package model

const (
	StatusDraft       = "draft"
	StatusSubmitted   = "submitted"
	StatusApproved    = "approved"
	StatusRejected    = "rejected"
	StatusChanged     = "changed"
	StatusArchived    = "archived"
	StageRegistration = "registration"
	StageReview       = "review"
	StageConfirmation = "confirmation"
	StageArchive      = "archive"
)

func ValidStatus(status string) bool {
	switch status {
	case StatusDraft, StatusSubmitted, StatusApproved, StatusRejected, StatusChanged, StatusArchived:
		return true
	default:
		return false
	}
}

func CanTransition(from, to string) bool {
	if from == StatusDraft && to == StatusSubmitted {
		return true
	}
	if from == StatusSubmitted && (to == StatusApproved || to == StatusRejected) {
		return true
	}
	if from == StatusApproved && to == StatusChanged {
		return true
	}
	if from == StatusChanged && to == StatusApproved {
		return true
	}
	if (from == StatusApproved || from == StatusChanged || from == StatusRejected) && to == StatusArchived {
		return true
	}
	return false
}
