package posts

import (
	"context"
	"gorm.io/gorm"
)

type Repository interface {
	Query(ctx context.Context, limit int, offset int, sortField string, reversedOrder bool) ([]Post, error)
}

type repo struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repo{db}
}

func (r *repo) Query(ctx context.Context, limit int, offset int, sortField string, reversedOrder bool) ([]Post, error) {
	if sortField == "" {
		sortField = "updated_at"
	}

	sortOrder := "desc"
	if reversedOrder {
		sortOrder = "asc"
	}

	var posts []Post
	res := r.db.WithContext(ctx).Order(sortField + " " + sortOrder).
		Limit(limit).
		Offset(offset).
		Find(&posts)
	if res.Error != nil {
		return nil, res.Error
	}

	return posts, nil
}
