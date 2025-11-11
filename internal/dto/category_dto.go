package dto

import "github.com/yoockh/go-game-rental-api/internal/model"

type CategoryDTO struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	IsActive    bool    `json:"is_active"`
}

type CreateCategoryRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description,omitempty"`
}

type UpdateCategoryRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description,omitempty"`
}

func ToCategoryDTO(category *model.Category) *CategoryDTO {
	if category == nil {
		return nil
	}

	return &CategoryDTO{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		IsActive:    category.IsActive,
	}
}

func ToCategoryDTOList(categories []*model.Category) []*CategoryDTO {
	result := make([]*CategoryDTO, len(categories))
	for i, category := range categories {
		result[i] = ToCategoryDTO(category)
	}
	return result
}
