package store

import (
	"sort"

	"example.com/animalcage/internal/model"
	"go.etcd.io/bbolt"
)

func (s *Store) SaveAttachment(attachment model.Attachment) error {
	if err := s.ensure(); err != nil {
		return err
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		return putJSON(tx, bucketNames[3], attachment.ID, attachment)
	})
}

func (s *Store) GetAttachment(id string) (model.Attachment, error) {
	if err := s.ensure(); err != nil {
		return model.Attachment{}, err
	}
	var attachment model.Attachment
	err := s.db.View(func(tx *bbolt.Tx) error {
		var err error
		attachment, err = getJSON[model.Attachment](tx, bucketNames[3], id)
		return err
	})
	if err != nil {
		return model.Attachment{}, err
	}
	attachment.Content = append([]byte(nil), attachment.Content...)
	return attachment, nil
}

func (s *Store) ListAttachmentsForRecord(recordID string) ([]model.Attachment, error) {
	if err := s.ensure(); err != nil {
		return nil, err
	}
	var attachments []model.Attachment
	err := s.db.View(func(tx *bbolt.Tx) error {
		items, listErr := listJSON[model.Attachment](tx, bucketNames[3])
		if listErr != nil {
			return listErr
		}
		for _, item := range items {
			if item.RecordID == recordID {
				item.Content = append([]byte(nil), item.Content...)
				attachments = append(attachments, item)
			}
		}
		return nil
	})
	sort.Slice(attachments, func(i, j int) bool { return attachments[i].ID < attachments[j].ID })
	return attachments, err
}

// ListAttachments satisfies the workflow service persistence boundary.
func (s *Store) ListAttachments(recordID string) ([]model.Attachment, error) {
	return s.ListAttachmentsForRecord(recordID)
}

func (s *Store) DeleteAttachment(id string) error {
	if err := s.ensure(); err != nil {
		return err
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		return deleteJSON(tx, bucketNames[3], id)
	})
}
