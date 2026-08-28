package store

import (
	"sort"

	"example.com/animalcage/internal/model"
	"go.etcd.io/bbolt"
)

// SaveRecord inserts or replaces a Record atomically.
func (s *Store) SaveRecord(record model.Record) error {
	if err := s.ensure(); err != nil {
		return err
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		return putJSON(tx, bucketNames[0], record.ID, record)
	})
}

// GetRecord returns a defensive copy and an explicit found flag so callers can
// distinguish an absent application from a storage failure.
func (s *Store) GetRecord(id string) (model.Record, bool, error) {
	if err := s.ensure(); err != nil {
		return model.Record{}, false, err
	}
	var record model.Record
	err := s.db.View(func(tx *bbolt.Tx) error {
		var err error
		record, err = getJSON[model.Record](tx, bucketNames[0], id)
		return err
	})
	if err != nil {
		if err == ErrNotFound {
			return model.Record{}, false, nil
		}
		return model.Record{}, false, err
	}
	return cloneRecord(record), true, nil
}

// DeleteRecord removes a record. Related audit and workflow rows are retained
// for traceability and can be explicitly deleted by their owners.
func (s *Store) DeleteRecord(id string) error {
	if err := s.ensure(); err != nil {
		return err
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		return deleteJSON(tx, bucketNames[0], id)
	})
}

// ListRecords returns records in stable ID order for deterministic workflows.
func (s *Store) ListRecords() ([]model.Record, error) {
	if err := s.ensure(); err != nil {
		return nil, err
	}
	var records []model.Record
	err := s.db.View(func(tx *bbolt.Tx) error {
		var err error
		records, err = listJSON[model.Record](tx, bucketNames[0])
		return err
	})
	if err != nil {
		return nil, err
	}
	for i := range records {
		records[i] = cloneRecord(records[i])
	}
	sort.Slice(records, func(i, j int) bool { return records[i].ID < records[j].ID })
	return records, nil
}
