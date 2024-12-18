//go:generate mockgen -destination=../machines/mock_machines_repository.go -package=machines -source=repository.go

package machines

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	pkgpg "ralts-cms/pkg/pg"
	"strings"
)

type Repository interface {
	Query(ctx context.Context, limit int, offset int, sortField string, reversedOrder bool) ([]Machine, error)
	GetBySerialNumber(ctx context.Context, serialNumber string) (*Machine, error)
	Create(ctx context.Context, m *Machine) (*Machine, error)
	Update(ctx context.Context, m *Machine) (*Machine, error)
	DeleteBySerialNumber(ctx context.Context, serialNumber string) error
}

type repo struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repo{db}
}

func (r *repo) Query(ctx context.Context, limit int, offset int, sortField string, reversedOrder bool) ([]Machine, error) {
	if sortField == "" {
		sortField = "updated_at"
	}

	sortOrder := "desc"
	if reversedOrder {
		sortOrder = "asc"
	}

	var machines []Machine
	res := r.db.WithContext(ctx).Order(fmt.Sprintf("%s %s", sortField, sortOrder)).
		Limit(limit).
		Offset(offset).
		Find(&machines)
	if res.Error != nil {
		return nil, fmt.Errorf("failed to query machines: %w", res.Error)
	}

	return machines, nil
}

func (r *repo) GetBySerialNumber(ctx context.Context, serialNumber string) (*Machine, error) {
	var m Machine
	res := r.db.WithContext(ctx).Where("serial_number = ?", serialNumber).First(&m)
	if res.Error != nil {
		if strings.Contains(res.Error.Error(), "record not found") {
			return nil, pkgpg.ErrNotFound
		}

		return nil, fmt.Errorf("failed to get machine: %w", res.Error)
	}

	return &m, nil
}

func (r *repo) Create(ctx context.Context, m *Machine) (*Machine, error) {
	res := r.db.WithContext(ctx).Create(m)
	if res.Error != nil {
		return nil, fmt.Errorf("failed to create machine: %w", res.Error)
	}

	return m, nil
}

func (r *repo) Update(ctx context.Context, m *Machine) (*Machine, error) {
	var updatedMachine Machine
	res := r.db.WithContext(ctx).Model(&updatedMachine).
		Clauses(clause.Returning{}).
		Where("serial_number = ?", m.SerialNumber).
		Select("*").
		Updates(m)
	if res.Error != nil {
		return nil, fmt.Errorf("failed to update machine: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, pkgpg.ErrNotFound
	}

	return &updatedMachine, nil
}

func (r *repo) DeleteBySerialNumber(ctx context.Context, serialNumber string) error {
	res := r.db.WithContext(ctx).Delete(&Machine{}, "serial_number = ?", serialNumber)
	if res.Error != nil {
		return fmt.Errorf("failed to delete machine: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return pkgpg.ErrNotFound
	}

	return nil
}
