package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormRepository[T any] struct {
	db *gorm.DB
}

func NewGormRepository[T any](db *gorm.DB) *GormRepository[T] {
	return &GormRepository[T]{db: db}
}

func (r *GormRepository[T]) FindByID(ctx context.Context, id string) (*T, error) {
	var entity T
	err := r.db.WithContext(ctx).First(&entity, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find by id: %w", err)
	}
	return &entity, nil
}

func (r *GormRepository[T]) FindAll(ctx context.Context, limit, offset int) ([]*T, error) {
	var entities []*T
	query := r.db.WithContext(ctx)
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	err := query.Find(&entities).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find all: %w", err)
	}
	return entities, nil
}

func (r *GormRepository[T]) Create(ctx context.Context, entity *T) error {
	err := r.db.WithContext(ctx).Create(entity).Error
	if err != nil {
		return fmt.Errorf("failed to create: %w", err)
	}
	return nil
}

func (r *GormRepository[T]) Update(ctx context.Context, entity *T) error {
	err := r.db.WithContext(ctx).Save(entity).Error
	if err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}
	return nil
}

func (r *GormRepository[T]) Delete(ctx context.Context, id string) error {
	var entity T
	result := r.db.WithContext(ctx).Delete(&entity, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("record not found")
	}
	return nil
}

func (r *GormRepository[T]) Count(ctx context.Context) (int64, error) {
	var count int64
	var entity T
	err := r.db.WithContext(ctx).Model(&entity).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count: %w", err)
	}
	return count, nil
}

func (r *GormRepository[T]) Exists(ctx context.Context, id string) (bool, error) {
	var count int64
	var entity T
	err := r.db.WithContext(ctx).Model(&entity).Where("id = ?", id).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}
	return count > 0, nil
}

func (r *GormRepository[T]) GetDB() *gorm.DB {
	return r.db
}

func (r *GormRepository[T]) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

func (r *GormRepository[T]) Preload(associations ...string) *GormRepository[T] {
	db := r.db
	for _, assoc := range associations {
		db = db.Preload(assoc)
	}
	return &GormRepository[T]{db: db}
}

func (r *GormRepository[T]) Where(query interface{}, args ...interface{}) *GormRepository[T] {
	return &GormRepository[T]{db: r.db.Where(query, args...)}
}

func (r *GormRepository[T]) Order(value interface{}) *GormRepository[T] {
	return &GormRepository[T]{db: r.db.Order(value)}
}

func (r *GormRepository[T]) Select(query interface{}, args ...interface{}) *GormRepository[T] {
	return &GormRepository[T]{db: r.db.Select(query, args...)}
}

func (r *GormRepository[T]) Raw(sql string, values ...interface{}) *GormRepository[T] {
	return &GormRepository[T]{db: r.db.Raw(sql, values...)}
}

func (r *GormRepository[T]) Joins(query string, args ...interface{}) *GormRepository[T] {
	return &GormRepository[T]{db: r.db.Joins(query, args...)}
}

func (r *GormRepository[T]) Group(name string) *GormRepository[T] {
	return &GormRepository[T]{db: r.db.Group(name)}
}

func (r *GormRepository[T]) Having(query interface{}, args ...interface{}) *GormRepository[T] {
	return &GormRepository[T]{db: r.db.Having(query, args...)}
}

func (r *GormRepository[T]) Scopes(funcs ...func(*gorm.DB) *gorm.DB) *GormRepository[T] {
	return &GormRepository[T]{db: r.db.Scopes(funcs...)}
}

func (r *GormRepository[T]) Clauses(conds ...clause.Expression) *GormRepository[T] {
	return &GormRepository[T]{db: r.db.Clauses(conds...)}
}
