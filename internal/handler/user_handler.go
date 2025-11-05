package handler

import (
	echomw "github.com/Yoochan45/go-api-utils/pkg-echo/middleware"
	myRequest "github.com/Yoochan45/go-api-utils/pkg-echo/request"
	myResponse "github.com/Yoochan45/go-api-utils/pkg-echo/response"
	"github.com/Yoochan45/go-game-rental-api/internal/dto"
	"github.com/Yoochan45/go-game-rental-api/internal/model"
	"github.com/Yoochan45/go-game-rental-api/internal/service"
	"github.com/Yoochan45/go-game-rental-api/internal/utils"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	userService service.UserService
	validate    *validator.Validate
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
		validate:    utils.GetValidator(),
	}
}

// GetMyProfile godoc
// @Summary Get current user profile
// @Description Get current user's profile information
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UserDTO "Profile retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /users/me [get]
func (h *UserHandler) GetMyProfile(c echo.Context) error {
	userID := echomw.CurrentUserID(c)
	if userID == 0 {
		return myResponse.Unauthorized(c, "Unauthorized")
	}

	user, err := h.userService.GetProfile(userID)
	if err != nil {
		return myResponse.NotFound(c, "User not found")
	}

	return myResponse.Success(c, "Profile retrieved successfully", user)
}

// UpdateMyProfile godoc
// @Summary Update current user profile
// @Description Update current user's profile information
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdateProfileRequest true "Profile update details"
// @Success 200 {object} dto.UserDTO "Profile updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /users/me [put]
func (h *UserHandler) UpdateMyProfile(c echo.Context) error {
	userID := echomw.CurrentUserID(c)
	if userID == 0 {
		return myResponse.Unauthorized(c, "Unauthorized")
	}

	var req dto.UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return myResponse.BadRequest(c, "Invalid input: "+err.Error())
	}
	if err := h.validate.Struct(&req); err != nil {
		return myResponse.BadRequest(c, "Validation error: "+err.Error())
	}

	err := h.userService.UpdateProfile(userID, &req)
	if err != nil {
		return myResponse.BadRequest(c, err.Error())
	}

	// Get updated profile
	user, err := h.userService.GetProfile(userID)
	if err != nil {
		return myResponse.InternalServerError(c, "Profile updated but failed to retrieve")
	}

	return myResponse.Success(c, "Profile updated successfully", user)
}

// GetAllUsers godoc
// @Summary Get all users
// @Description Get list of all users (Admin only)
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{} "Users retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /admin/users [get]
func (h *UserHandler) GetAllUsers(c echo.Context) error {
	params := utils.ParsePagination(c)
	role := echomw.CurrentRole(c)

	users, totalCount, err := h.userService.GetAllUsers(model.UserRole(role), params.Limit, params.Offset)
	if err != nil {
		return myResponse.Forbidden(c, err.Error())
	}

	meta := utils.CreateMeta(params, totalCount)
	return myResponse.Paginated(c, "Users retrieved successfully", users, meta)
}

// GetUserDetail godoc
// @Summary Get user detail
// @Description Get detailed information about a specific user (Admin only)
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} dto.UserDTO "User retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid user ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Router /admin/users/{id} [get]
func (h *UserHandler) GetUserDetail(c echo.Context) error {
	userID := myRequest.PathParamUint(c, "id")
	if userID == 0 {
		return myResponse.BadRequest(c, "Invalid user ID")
	}

	role := echomw.CurrentRole(c)
	user, err := h.userService.GetUserDetail(model.UserRole(role), userID)
	if err != nil {
		return myResponse.Forbidden(c, err.Error())
	}

	return myResponse.Success(c, "User retrieved successfully", user)
}

// UpdateUserRole godoc
// @Summary Update user role
// @Description Update user role (Admin only)
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param request body dto.UpdateUserRoleRequest true "Role update details"
// @Success 200 {object} map[string]interface{} "User role updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /admin/users/{id}/role [patch]
func (h *UserHandler) UpdateUserRole(c echo.Context) error {
	userID := myRequest.PathParamUint(c, "id")
	if userID == 0 {
		return myResponse.BadRequest(c, "Invalid user ID")
	}

	var req dto.UpdateUserRoleRequest
	if err := c.Bind(&req); err != nil {
		return myResponse.BadRequest(c, "Invalid input: "+err.Error())
	}
	if err := h.validate.Struct(&req); err != nil {
		return myResponse.BadRequest(c, "Validation error: "+err.Error())
	}

	role := echomw.CurrentRole(c)

	// UpdateUserRole only returns error, not (user, error)
	err := h.userService.UpdateUserRole(model.UserRole(role), userID, req.Role)
	if err != nil {
		return myResponse.Forbidden(c, err.Error())
	}

	// Get updated user for response
	user, err := h.userService.GetUserDetail(model.UserRole(role), userID)
	if err != nil {
		return myResponse.InternalServerError(c, "Role updated but failed to retrieve user")
	}

	return myResponse.Success(c, "User role updated successfully", user)
}

// ToggleUserStatus godoc
// @Summary Toggle user status
// @Description Ban or unban a user (Admin only)
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} map[string]interface{} "User status updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid user ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /admin/users/{id}/status [patch]
func (h *UserHandler) ToggleUserStatus(c echo.Context) error {
	userID := myRequest.PathParamUint(c, "id")
	if userID == 0 {
		return myResponse.BadRequest(c, "Invalid user ID")
	}

	role := echomw.CurrentRole(c)

	// ToggleUserStatus only returns error, not (user, error)
	err := h.userService.ToggleUserStatus(model.UserRole(role), userID)
	if err != nil {
		return myResponse.Forbidden(c, err.Error())
	}

	// Get updated user for response
	user, err := h.userService.GetUserDetail(model.UserRole(role), userID)
	if err != nil {
		return myResponse.InternalServerError(c, "Status updated but failed to retrieve user")
	}

	return myResponse.Success(c, "User status updated successfully", user)
}

// DeleteUser godoc
// @Summary Delete user
// @Description Soft delete a user (Admin only)
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} map[string]interface{} "User deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid user ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /admin/users/{id} [delete]
func (h *UserHandler) DeleteUser(c echo.Context) error {
	userID := myRequest.PathParamUint(c, "id")
	if userID == 0 {
		return myResponse.BadRequest(c, "Invalid user ID")
	}

	// Get current user info for delete operation
	currentUserID := echomw.CurrentUserID(c)
	role := echomw.CurrentRole(c)

	// Fix argument order: (requestorID, requestorRole, targetUserID)
	err := h.userService.DeleteUser(currentUserID, model.UserRole(role), userID)
	if err != nil {
		return myResponse.Forbidden(c, err.Error())
	}

	return myResponse.Success(c, "User deleted successfully", nil)
}
