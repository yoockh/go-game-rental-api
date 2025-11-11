package handler

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	echomw "github.com/yoockh/go-api-utils/pkg-echo/middleware"
	myRequest "github.com/yoockh/go-api-utils/pkg-echo/request"
	myResponse "github.com/yoockh/go-api-utils/pkg-echo/response"
	"github.com/yoockh/go-game-rental-api/internal/dto"
	"github.com/yoockh/go-game-rental-api/internal/model"
	"github.com/yoockh/go-game-rental-api/internal/service"
	"github.com/yoockh/go-game-rental-api/internal/utils"
)

type CategoryHandler struct {
	categoryService service.CategoryService
	validate        *validator.Validate
}

func NewCategoryHandler(categoryService service.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
		validate:        utils.GetValidator(),
	}
}

// GetAllCategories godoc
// @Summary Get active categories
// @Description Get list of active game categories
// @Tags Categories
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Categories retrieved successfully"
// @Router /categories [get]
func (h *CategoryHandler) GetAllCategories(c echo.Context) error {
	categories, err := h.categoryService.GetActiveCategories()
	if err != nil {
		return myResponse.InternalServerError(c, "Failed to retrieve categories")
	}

	return myResponse.Success(c, "Categories retrieved successfully", categories)
}

// GetCategoryDetail godoc
// @Summary Get category detail
// @Description Get detailed information about a specific category
// @Tags Categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} model.Category "Category retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid category ID"
// @Failure 404 {object} map[string]interface{} "Category not found"
// @Router /categories/{id} [get]
func (h *CategoryHandler) GetCategoryDetail(c echo.Context) error {
	categoryID := myRequest.PathParamUint(c, "id")
	if categoryID == 0 {
		return myResponse.BadRequest(c, "Invalid category ID")
	}

	category, err := h.categoryService.GetCategoryByID(categoryID)
	if err != nil {
		return myResponse.NotFound(c, "Category not found")
	}

	return myResponse.Success(c, "Category retrieved successfully", category)
}

// CreateCategory godoc
// @Summary Create category
// @Description Create a new game category (Admin only)
// @Tags Admin - Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateCategoryRequest true "Category details"
// @Success 201 {object} map[string]interface{} "Category created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /admin/categories [post]
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
		Description: utils.PtrOrNil(req.Description),
		IsActive:    true,
	}

	err := h.categoryService.CreateCategory(model.UserRole(role), categoryData)
	if err != nil {
		return myResponse.BadRequest(c, err.Error())
	}

	return myResponse.Created(c, "Category created successfully", categoryData)
}

// UpdateCategory godoc
// @Summary Update category
// @Description Update category information (Admin only)
// @Tags Admin - Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Category ID"
// @Param request body dto.UpdateCategoryRequest true "Updated category details"
// @Success 200 {object} map[string]interface{} "Category updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /admin/categories/{id} [put]
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
		Description: utils.PtrOrNil(req.Description),
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

	return myResponse.Success(c, "Category updated successfully", category)
}

// DeleteCategory godoc
// @Summary Delete category
// @Description Delete a category (Admin only)
// @Tags Admin - Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Category ID"
// @Success 200 {object} map[string]interface{} "Category deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid category ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /admin/categories/{id} [delete]
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
