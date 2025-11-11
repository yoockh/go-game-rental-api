package service

import (
	"errors"

	"github.com/yoockh/go-game-rental-api/internal/model"
	"github.com/yoockh/go-game-rental-api/internal/repository"
)

var (
	ErrCategoryNotFound = errors.New("category not found")
	ErrCategoryHasGames = errors.New("cannot delete category with existing games")
)

type CategoryService interface {
	// Public methods
	GetAllCategories() ([]*model.Category, error)
	GetActiveCategories() ([]*model.Category, error)
	GetCategoryByID(id uint) (*model.Category, error)

	// Admin methods
	CreateCategory(requestorRole model.UserRole, categoryData *model.Category) error
	UpdateCategory(requestorRole model.UserRole, categoryID uint, updateData *model.Category) error
	DeleteCategory(requestorRole model.UserRole, categoryID uint) error
	ToggleCategoryStatus(requestorRole model.UserRole, categoryID uint) error
}

type categoryService struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryService(categoryRepo repository.CategoryRepository) CategoryService {
	return &categoryService{categoryRepo: categoryRepo}
}

func (s *categoryService) GetAllCategories() ([]*model.Category, error) {
	return s.categoryRepo.GetAll()
}

func (s *categoryService) GetActiveCategories() ([]*model.Category, error) {
	return s.categoryRepo.GetActiveCategories()
}

func (s *categoryService) GetCategoryByID(id uint) (*model.Category, error) {
	return s.categoryRepo.GetByID(id)
}

func (s *categoryService) CreateCategory(requestorRole model.UserRole, categoryData *model.Category) error {
	if !s.canManageCategories(requestorRole) {
		return ErrInsufficientPermission
	}

	return s.categoryRepo.Create(categoryData)
}

func (s *categoryService) UpdateCategory(requestorRole model.UserRole, categoryID uint, updateData *model.Category) error {
	if !s.canManageCategories(requestorRole) {
		return ErrInsufficientPermission
	}

	category, err := s.categoryRepo.GetByID(categoryID)
	if err != nil {
		return ErrCategoryNotFound
	}

	category.Name = updateData.Name
	category.Description = updateData.Description

	return s.categoryRepo.Update(category)
}

func (s *categoryService) DeleteCategory(requestorRole model.UserRole, categoryID uint) error {
	if !s.canManageCategories(requestorRole) {
		return ErrInsufficientPermission
	}

	// Check if category has games
	gamesCount, err := s.categoryRepo.CountGamesInCategory(categoryID)
	if err != nil {
		return err
	}

	if gamesCount > 0 {
		return ErrCategoryHasGames
	}

	return s.categoryRepo.Delete(categoryID)
}

func (s *categoryService) ToggleCategoryStatus(requestorRole model.UserRole, categoryID uint) error {
	if !s.canManageCategories(requestorRole) {
		return ErrInsufficientPermission
	}

	category, err := s.categoryRepo.GetByID(categoryID)
	if err != nil {
		return ErrCategoryNotFound
	}

	return s.categoryRepo.UpdateActiveStatus(categoryID, !category.IsActive)
}

func (s *categoryService) canManageCategories(role model.UserRole) bool {
	return role == model.RoleAdmin || role == model.RoleSuperAdmin
}
