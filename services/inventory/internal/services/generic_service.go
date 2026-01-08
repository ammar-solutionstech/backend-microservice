package services

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GenericService provides CRUD helpers for simple resources.
type GenericService[T any] struct {
	db *gorm.DB
}

// NewGenericService constructs a generic CRUD service.
func NewGenericService[T any](db *gorm.DB) *GenericService[T] {
	return &GenericService[T]{db: db}
}

// List fetches all records.
func (s *GenericService[T]) List(out *[]T) error {
	return s.db.Find(out).Error
}

// Get finds a single record by its identifier.
func (s *GenericService[T]) Get(id int, out *T) error {
	if err := s.db.Preload(clause.Associations).First(out, id).Error; err != nil {
		return err
	}
	return nil
}

// Create persists a new record.
func (s *GenericService[T]) Create(entity *T) error {
	return s.db.Create(entity).Error
}

// Update applies partial updates to an existing record.
func (s *GenericService[T]) Update(id int, updates map[string]interface{}) (*T, error) {
	var entity T
	if err := s.db.First(&entity, id).Error; err != nil {
		return nil, err
	}
	if len(updates) == 0 {
		return &entity, nil
	}
	if err := s.db.Model(&entity).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

// Delete removes a record by id.
func (s *GenericService[T]) Delete(id int) error {
	result := s.db.Delete(new(T), id)
	if err := result.Error; err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Exists returns true if the record exists.
func (s *GenericService[T]) Exists(id int) (bool, error) {
	var entity T
	err := s.db.Select("id").First(&entity, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

