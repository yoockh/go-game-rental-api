package handler

import (
	echomw "github.com/Yoochan45/go-api-utils/pkg-echo/middleware"
	myRequest "github.com/Yoochan45/go-api-utils/pkg-echo/request"
	myResponse "github.com/Yoochan45/go-api-utils/pkg-echo/response"
	"github.com/Yoochan45/go-game-rental-api/internal/model"
	"github.com/Yoochan45/go-game-rental-api/internal/model/dto"
	"github.com/Yoochan45/go-game-rental-api/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type CategoryHandler struct {
	categoryService service.CategoryService
	validate        *validator.Validate
}

func NewCategoryHandler(categoryService service.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
		validate:        validator.New(),
	}
}

// Public endpoints
func (h *CategoryHandler) GetAllCategories(c echo.Context) error {
	categories, err := h.categoryService.GetActiveCategories()
	if err != nil {
		return myResponse.InternalServerError(c, "Failed to retrieve categories")
	}

	categoryDTOs := dto.ToCategoryDTOList(categories)
	return myResponse.Success(c, "Categories retrieved successfully", categoryDTOs)
}

func (h *CategoryHandler) GetCategoryDetail(c echo.Context) error {
	categoryID := myRequest.PathParamUint(c, "id")
	if categoryID == 0 {
		return myResponse.BadRequest(c, "Invalid category ID")
	}

	category, err := h.categoryService.GetCategoryByID(categoryID)
	if err != nil {
		return myResponse.NotFound(c, "Category not found")
	}

	response := dto.ToCategoryDTO(category)
	return myResponse.Success(c, "Category retrieved successfully", response)
}

// Admin endpoints
func (h *CategoryHandler) CreateCategory(c echo.Context) error {
	var req dto.CreateCategoryRequest
	if err := c.Bind(&req); err != nil {
		return myResponse.BadRequest(c, "Invalid input: "+err.Error())
	}
	if err := h.validate.Struct(&req); err != nil {
		return myResponse.BadRequest(c, "Validation error: "+err.Error())
	}

	role := echomw.CurrentRole(c)

	// Create category data
	categoryData := &model.Category{
		Name:        req.Name,
		Description: &req.Description,
		IsActive:    true,
	}

	err := h.categoryService.CreateCategory(model.UserRole(role), categoryData)
	if err != nil {
		return myResponse.BadRequest(c, err.Error())
	}

	response := dto.ToCategoryDTO(categoryData)
	return myResponse.Created(c, "Category created successfully", response)
}

func (h *CategoryHandler) UpdateCategory(c echo.Context) error {
	categoryID := myRequest.PathParamUint(c, "id")
	if categoryID == 0 {
		return myResponse.BadRequest(c, "Invalid category ID")
	}

	var req dto.UpdateCategoryRequest
	if err := c.Bind(&req); err != nil {
		return myResponse.BadRequest(c, "Invalid input: "+err.Error())
	}
	if err := h.validate.Struct(&req); err != nil {
		return myResponse.BadRequest(c, "Validation error: "+err.Error())
	}

	role := echomw.CurrentRole(c)

	// Create update data
	updateData := &model.Category{
		Name:        req.Name,
		Description: &req.Description,
	}

	err := h.categoryService.UpdateCategory(model.UserRole(role), categoryID, updateData)
	if err != nil {
		return myResponse.BadRequest(c, err.Error())
	}

	// Get updated category for response
	category, err := h.categoryService.GetCategoryByID(categoryID)
	if err != nil {
		return myResponse.InternalServerError(c, "Failed to retrieve updated category")
	}

	response := dto.ToCategoryDTO(category)
	return myResponse.Success(c, "Category updated successfully", response)
}

func (h *CategoryHandler) ToggleCategoryStatus(c echo.Context) error {
	categoryID := myRequest.PathParamUint(c, "id")
	if categoryID == 0 {
		return myResponse.BadRequest(c, "Invalid category ID")
	}

	role := echomw.CurrentRole(c)
	err := h.categoryService.ToggleCategoryStatus(model.UserRole(role), categoryID)
	if err != nil {
		return myResponse.BadRequest(c, err.Error())
	}

	// Get updated category for response
	category, err := h.categoryService.GetCategoryByID(categoryID)
	if err != nil {
		return myResponse.InternalServerError(c, "Failed to retrieve updated category")
	}

	response := dto.ToCategoryDTO(category)
	return myResponse.Success(c, "Category status updated successfully", response)
}

func (h *CategoryHandler) DeleteCategory(c echo.Context) error {
	categoryID := myRequest.PathParamUint(c, "id")
	if categoryID == 0 {
		return myResponse.BadRequest(c, "Invalid category ID")
	}

	role := echomw.CurrentRole(c)
	err := h.categoryService.DeleteCategory(model.UserRole(role), categoryID)
	if err != nil {
		return myResponse.BadRequest(c, err.Error())
	}

	return myResponse.Success(c, "Category deleted successfully", nil)
}

func (h *CategoryHandler) GetAllCategoriesAdmin(c echo.Context) error {
	// Remove role parameter since GetAllCategories doesn't need it
	categories, err := h.categoryService.GetAllCategories()
	if err != nil {
		return myResponse.InternalServerError(c, "Failed to retrieve categories")
	}

	categoryDTOs := dto.ToCategoryDTOList(categories)
	return myResponse.Success(c, "Categories retrieved successfully", categoryDTOs)
}
