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

type UserHandler struct {
	userService service.UserService
	validate    *validator.Validate
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
		validate:    validator.New(),
	}
}

func (h *UserHandler) GetAllUsers(c echo.Context) error {
	page := myRequest.QueryInt(c, "page", 1)
	limit := myRequest.QueryInt(c, "limit", 10)

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	role := echomw.CurrentRole(c)

	// GetAllUsers returns 2 values: ([]*model.User, error), not 3
	users, err := h.userService.GetAllUsers(model.UserRole(role), limit, (page-1)*limit)
	if err != nil {
		return myResponse.Forbidden(c, err.Error())
	}

	userDTOs := dto.ToUserDTOList(users)

	// Calculate total count manually if needed
	totalCount := int64(len(users))

	meta := map[string]any{
		"page":        page,
		"limit":       limit,
		"total":       totalCount,
		"total_pages": (totalCount + int64(limit) - 1) / int64(limit),
	}

	return myResponse.Paginated(c, "Users retrieved successfully", userDTOs, meta)
}

func (h *UserHandler) GetUserDetail(c echo.Context) error {
	userID := myRequest.PathParamUint(c, "id")
	if userID == 0 {
		return myResponse.BadRequest(c, "Invalid user ID")
	}

	role := echomw.CurrentRole(c)
	user, err := h.userService.GetUserDetail(model.UserRole(role), userID)
	if err != nil {
		return myResponse.NotFound(c, err.Error())
	}

	response := dto.ToUserDTO(user)
	return myResponse.Success(c, "User retrieved successfully", response)
}

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
		return myResponse.BadRequest(c, err.Error())
	}

	// Get updated user for response
	user, err := h.userService.GetUserDetail(model.UserRole(role), userID)
	if err != nil {
		return myResponse.InternalServerError(c, "Role updated but failed to retrieve user")
	}

	response := dto.ToUserDTO(user)
	return myResponse.Success(c, "User role updated successfully", response)
}

func (h *UserHandler) ToggleUserStatus(c echo.Context) error {
	userID := myRequest.PathParamUint(c, "id")
	if userID == 0 {
		return myResponse.BadRequest(c, "Invalid user ID")
	}

	role := echomw.CurrentRole(c)

	// ToggleUserStatus only returns error, not (user, error)
	err := h.userService.ToggleUserStatus(model.UserRole(role), userID)
	if err != nil {
		return myResponse.BadRequest(c, err.Error())
	}

	// Get updated user for response
	user, err := h.userService.GetUserDetail(model.UserRole(role), userID)
	if err != nil {
		return myResponse.InternalServerError(c, "Status updated but failed to retrieve user")
	}

	response := dto.ToUserDTO(user)
	return myResponse.Success(c, "User status updated successfully", response)
}

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
		return myResponse.BadRequest(c, err.Error())
	}

	return myResponse.Success(c, "User deleted successfully", nil)
}
