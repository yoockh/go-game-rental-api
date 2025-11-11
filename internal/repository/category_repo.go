package repository

import (
	"github.com/yoockh/go-game-rental-api/internal/model"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	// Basic CRUD
	Create(category *model.Category) error
	GetByID(id uint) (*model.Category, error)
	GetAll() ([]*model.Category, error)
	GetActiveCategories() ([]*model.Category, error)
	Update(category *model.Category) error
	Delete(id uint) error

	// Admin methods
	UpdateActiveStatus(categoryID uint, isActive bool) error

	// Statistics
	CountGamesInCategory(categoryID uint) (int64, error)
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(category *model.Category) error {
	return r.db.Create(category).Error
}

func (r *categoryRepository) GetByID(id uint) (*model.Category, error) {
	var category model.Category
	err := r.db.Where("id = ?", id).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) GetAll() ([]*model.Category, error) {
	var categories []*model.Category
	err := r.db.Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) GetActiveCategories() ([]*model.Category, error) {
	var categories []*model.Category
	err := r.db.Where("is_active = ?", true).Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) Update(category *model.Category) error {
	return r.db.Save(category).Error
}

func (r *categoryRepository) Delete(id uint) error {
	return r.db.Delete(&model.Category{}, id).Error
}

func (r *categoryRepository) UpdateActiveStatus(categoryID uint, isActive bool) error {
	return r.db.Model(&model.Category{}).Where("id = ?", categoryID).Update("is_active", isActive).Error
}

func (r *categoryRepository) CountGamesInCategory(categoryID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Game{}).Where("category_id = ?", categoryID).Count(&count).Error
	return count, err
}
