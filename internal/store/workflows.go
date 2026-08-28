package store

import (
	"sort"

	"example.com/animalcage/internal/model"
	"go.etcd.io/bbolt"
)

func (s *Store) SaveWorkflow(workflow model.Workflow) error {
	if err := s.ensure(); err != nil {
		return err
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		return putJSON(tx, bucketNames[2], workflow.ID, workflow)
	})
}

func (s *Store) GetWorkflow(id string) (model.Workflow, error) {
	if err := s.ensure(); err != nil {
		return model.Workflow{}, err
	}
	var workflow model.Workflow
	err := s.db.View(func(tx *bbolt.Tx) error {
		var err error
		workflow, err = getJSON[model.Workflow](tx, bucketNames[2], id)
		return err
	})
	return workflow, err
}

func (s *Store) ListWorkflowsForRecord(recordID string) ([]model.Workflow, error) {
	if err := s.ensure(); err != nil {
		return nil, err
	}
	var workflows []model.Workflow
	err := s.db.View(func(tx *bbolt.Tx) error {
		items, listErr := listJSON[model.Workflow](tx, bucketNames[2])
		if listErr != nil {
			return listErr
		}
		for _, item := range items {
			if item.RecordID == recordID {
				workflows = append(workflows, item)
			}
		}
		return nil
	})
	sort.Slice(workflows, func(i, j int) bool { return workflows[i].ID < workflows[j].ID })
	return workflows, err
}

// ListWorkflows satisfies the workflow service persistence boundary.
func (s *Store) ListWorkflows(recordID string) ([]model.Workflow, error) {
	return s.ListWorkflowsForRecord(recordID)
}

func (s *Store) DeleteWorkflow(id string) error {
	if err := s.ensure(); err != nil {
		return err
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		return deleteJSON(tx, bucketNames[2], id)
	})
}
