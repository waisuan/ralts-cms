package machines

import (
	"fmt"
	"gorm.io/gorm"
)

type Repository interface {
	Query(limit int, offset int) ([]Machine, error)
	Create(m *Machine) (*Machine, error)
}

type repo struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repo{db}
}

func (r *repo) Query(limit int, offset int) ([]Machine, error) {
	var machines []Machine
	res := r.db.Find(&machines).
		Order("updated_at desc").
		Limit(limit).
		Offset(offset)

	if res.Error != nil {
		return nil, fmt.Errorf("failed to query machines: %w", res.Error)
	}

	return machines, nil
}

func (r *repo) Create(m *Machine) (*Machine, error) {
	res := r.db.Create(m)
	if res.Error != nil {
		return nil, fmt.Errorf("failed to create machine: %w", res.Error)
	}

	return m, nil
}
