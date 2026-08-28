package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"example.com/animalcage/internal/model"
	"go.etcd.io/bbolt"
)

var (
	ErrNotFound  = errors.New("store: entity not found")
	ErrInvalidID = errors.New("store: entity id is required")
)

var bucketNames = [][]byte{
	[]byte("records"),
	[]byte("audit_events"),
	[]byte("workflows"),
	[]byte("attachments"),
}

// Store owns one bbolt database and provides typed persistence boundaries for
// the application. The database is opened once and may be closed and reopened
// without changing the serialized representation.
type Store struct {
	db *bbolt.DB
}

// Open creates a store at path and initializes all buckets. A parent directory
// is created by bbolt's caller, so this function only normalizes the file name.
func Open(path string) (*Store, error) {
	if path == "" {
		return nil, fmt.Errorf("store: database path is required")
	}
	db, err := bbolt.Open(filepath.Clean(path), 0o600, nil)
	if err != nil {
		return nil, fmt.Errorf("store: open database: %w", err)
	}
	s := &Store{db: db}
	if err := s.db.Update(func(tx *bbolt.Tx) error {
		for _, name := range bucketNames {
			if _, err := tx.CreateBucketIfNotExists(name); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("store: initialize buckets: %w", err)
	}
	return s, nil
}

func (s *Store) ensure() error {
	if s == nil || s.db == nil {
		return errors.New("store: closed")
	}
	return nil
}

// Close flushes pending writes and releases the database file lock.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	err := s.db.Close()
	s.db = nil
	return err
}

func putJSON[T any](tx *bbolt.Tx, bucket []byte, id string, value T) error {
	if id == "" {
		return ErrInvalidID
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("store: encode %s: %w", id, err)
	}
	return tx.Bucket(bucket).Put([]byte(id), payload)
}

func getJSON[T any](tx *bbolt.Tx, bucket []byte, id string) (T, error) {
	var zero T
	if id == "" {
		return zero, ErrInvalidID
	}
	value := tx.Bucket(bucket).Get([]byte(id))
	if value == nil {
		return zero, ErrNotFound
	}
	if err := json.Unmarshal(value, &zero); err != nil {
		return zero, fmt.Errorf("store: decode %s: %w", id, err)
	}
	return zero, nil
}

func deleteJSON(tx *bbolt.Tx, bucket []byte, id string) error {
	if id == "" {
		return ErrInvalidID
	}
	if tx.Bucket(bucket).Get([]byte(id)) == nil {
		return ErrNotFound
	}
	return tx.Bucket(bucket).Delete([]byte(id))
}

func listJSON[T any](tx *bbolt.Tx, bucket []byte) ([]T, error) {
	items := make([]T, 0)
	err := tx.Bucket(bucket).ForEach(func(_, value []byte) error {
		var item T
		if err := json.Unmarshal(value, &item); err != nil {
			return err
		}
		items = append(items, item)
		return nil
	})
	return items, err
}

func cloneRecord(record model.Record) model.Record { return model.CopyRecord(record) }
