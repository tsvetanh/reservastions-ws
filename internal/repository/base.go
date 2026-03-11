package repository

import "gorm.io/gorm"

type BaseRepository[T any] struct {
	DB *gorm.DB
}

func (r *BaseRepository[T]) Create(obj *T) error {
	return r.DB.Create(obj).Error
}

func (r *BaseRepository[T]) GetAll(objs *[]T, preloads ...string) error {
	query := r.DB
	for _, preload := range preloads {
		query = query.Preload(preload)
	}
	return query.Find(objs).Error
}

func (r *BaseRepository[T]) GetByID(id any, obj *T, preloads ...string) error {
	query := r.DB
	for _, preload := range preloads {
		query = query.Preload(preload)
	}
	return query.First(obj, id).Error
}

func (r *BaseRepository[T]) Update(id any, obj *T) error {
	return r.DB.Model(obj).Where("id = ?", id).Updates(obj).Error
}

func (r *BaseRepository[T]) Delete(id any) error {
	return r.DB.Delete(new(T), id).Error
}
