package store

import (
	"sort"

	"example.com/animalcage/internal/model"
	"go.etcd.io/bbolt"
)

// SaveAuditEvent records an immutable workflow event keyed by its ID.
func (s *Store) SaveAuditEvent(event model.AuditEvent) error {
	if err := s.ensure(); err != nil {
		return err
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		return putJSON(tx, bucketNames[1], event.ID, event)
	})
}

// SaveAudit satisfies the workflow service persistence boundary.
func (s *Store) SaveAudit(event model.AuditEvent) error { return s.SaveAuditEvent(event) }

// AppendEvent is the workflow-friendly name for SaveAuditEvent. IDs remain
// caller supplied so deterministic tests can replay the same event stream.
func (s *Store) AppendEvent(event model.AuditEvent) error {
	return s.SaveAuditEvent(event)
}

func (s *Store) GetAuditEvent(id string) (model.AuditEvent, error) {
	if err := s.ensure(); err != nil {
		return model.AuditEvent{}, err
	}
	var event model.AuditEvent
	err := s.db.View(func(tx *bbolt.Tx) error {
		var err error
		event, err = getJSON[model.AuditEvent](tx, bucketNames[1], id)
		return err
	})
	return event, err
}

// ListAuditEventsForRecord returns events in timestamp then ID order.
func (s *Store) ListAuditEventsForRecord(recordID string) ([]model.AuditEvent, error) {
	if err := s.ensure(); err != nil {
		return nil, err
	}
	all, err := s.listAuditEvents()
	if err != nil {
		return nil, err
	}
	filtered := make([]model.AuditEvent, 0, len(all))
	for _, event := range all {
		if event.RecordID == recordID {
			filtered = append(filtered, event)
		}
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		if filtered[i].At.Equal(filtered[j].At) {
			return filtered[i].ID < filtered[j].ID
		}
		return filtered[i].At.Before(filtered[j].At)
	})
	return filtered, nil
}

// ListAudit satisfies the workflow service persistence boundary.
func (s *Store) ListAudit(recordID string) ([]model.AuditEvent, error) {
	return s.ListAuditEventsForRecord(recordID)
}

// ListEvents is an alias used by orchestration services that need the complete
// event stream. Results are ordered by event timestamp and then ID.
func (s *Store) ListEvents() ([]model.AuditEvent, error) {
	if err := s.ensure(); err != nil {
		return nil, err
	}
	events, err := s.listAuditEvents()
	if err != nil {
		return nil, err
	}
	sort.SliceStable(events, func(i, j int) bool {
		if events[i].At.Equal(events[j].At) {
			return events[i].ID < events[j].ID
		}
		return events[i].At.Before(events[j].At)
	})
	return events, nil
}

func (s *Store) listAuditEvents() ([]model.AuditEvent, error) {
	var events []model.AuditEvent
	err := s.db.View(func(tx *bbolt.Tx) error {
		var err error
		events, err = listJSON[model.AuditEvent](tx, bucketNames[1])
		return err
	})
	return events, err
}
