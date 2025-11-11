package handler

import (
	"log"

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

type GameHandler struct {
	gameService service.GameService
	validate    *validator.Validate
}

func NewGameHandler(gameService service.GameService) *GameHandler {
	return &GameHandler{
		gameService: gameService,
		validate:    utils.GetValidator(),
	}
}

// GetAllGames godoc
// @Summary Get all games
// @Description Get list of all active games
// @Tags Games
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{} "Games retrieved successfully"
// @Router /games [get]
func (h *GameHandler) GetAllGames(c echo.Context) error {
	params := utils.ParsePagination(c)

	log.Printf("DEBUG GetAllGames: limit=%d, offset=%d", params.Limit, params.Offset)

	games, total, err := h.gameService.GetAll(params.Limit, params.Offset)
	if err != nil {
		log.Printf("ERROR GetAllGames: %v", err)
		return myResponse.InternalServerError(c, "Failed to retrieve games")
	}

	log.Printf("DEBUG GetAllGames: found %d games, total=%d", len(games), total)

	meta := utils.CreateMeta(params, total)
	return myResponse.Paginated(c, "Games retrieved successfully", games, meta)
}

// GetGameDetail godoc
// @Summary Get game detail
// @Description Get detailed information about a specific game
// @Tags Games
// @Accept json
// @Produce json
// @Param id path int true "Game ID"
// @Success 200 {object} model.Game "Game retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid game ID"
// @Failure 404 {object} map[string]interface{} "Game not found"
// @Router /games/{id} [get]
func (h *GameHandler) GetGameDetail(c echo.Context) error {
	gameID := myRequest.PathParamUint(c, "id")
	if gameID == 0 {
		return myResponse.BadRequest(c, "Invalid game ID")
	}

	game, err := h.gameService.GetByID(gameID) // Fix: GetByID bukan GetGameDetail
	if err != nil {
		return myResponse.NotFound(c, "Game not found")
	}

	return myResponse.Success(c, "Game retrieved successfully", game)
}

// SearchGames godoc
// @Summary Search games
// @Description Search games by name, description, or platform
// @Tags Games
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{} "Games search results"
// @Failure 400 {object} map[string]interface{} "Search query required"
// @Router /games/search [get]
func (h *GameHandler) SearchGames(c echo.Context) error {
	query := myRequest.QueryString(c, "q", "")
	if query == "" {
		return myResponse.BadRequest(c, "Search query is required")
	}

	params := utils.ParsePagination(c)

	games, err := h.gameService.Search(query, params.Limit, params.Offset)
	if err != nil {
		return myResponse.InternalServerError(c, "Failed to search games")
	}

	meta := utils.CreateMeta(params, int64(len(games)))
	return myResponse.Paginated(c, "Games search results", games, meta)
}

// CreateGame godoc
// @Summary Create new game
// @Description Create a new game listing (Admin only)
// @Tags Admin - Games
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateGameRequest true "Game details"
// @Success 201 {object} map[string]interface{} "Game created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /admin/games [post]
func (h *GameHandler) CreateGame(c echo.Context) error {
	var req dto.CreateGameRequest
	if err := c.Bind(&req); err != nil {
		return myResponse.BadRequest(c, "Invalid input: "+err.Error())
	}
	if err := h.validate.Struct(&req); err != nil {
		return myResponse.BadRequest(c, "Validation error: "+err.Error())
	}

	adminID := echomw.CurrentUserID(c)
	role := echomw.CurrentRole(c)

	gameData := &model.Game{
		CategoryID:        req.CategoryID,
		Name:              req.Name,
		Description:       utils.PtrOrNil(req.Description),
		Platform:          utils.PtrOrNil(req.Platform),
		Stock:             req.Stock,
		RentalPricePerDay: req.RentalPricePerDay,
		SecurityDeposit:   req.SecurityDeposit,
		Condition:         model.GameCondition(req.Condition),
	}

	err := h.gameService.Create(adminID, model.UserRole(role), gameData)
	if err != nil {
		return myResponse.Forbidden(c, err.Error())
	}

	return myResponse.Created(c, "Game created successfully", gameData)
}

// UpdateGame godoc
// @Summary Update game
// @Description Update game information (Admin only)
// @Tags Admin - Games
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Game ID"
// @Param request body dto.UpdateGameRequest true "Updated game details"
// @Success 200 {object} map[string]interface{} "Game updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /admin/games/{id} [put]
func (h *GameHandler) UpdateGame(c echo.Context) error {
	gameID := myRequest.PathParamUint(c, "id")
	if gameID == 0 {
		return myResponse.BadRequest(c, "Invalid game ID")
	}

	var req dto.UpdateGameRequest
	if err := c.Bind(&req); err != nil {
		return myResponse.BadRequest(c, "Invalid input: "+err.Error())
	}
	if err := h.validate.Struct(&req); err != nil {
		return myResponse.BadRequest(c, "Validation error: "+err.Error())
	}

	game, err := h.gameService.GetByID(gameID)
	if err != nil {
		return myResponse.NotFound(c, "Game not found")
	}

	// Update only if provided
	if req.CategoryID > 0 {
		game.CategoryID = req.CategoryID
	}
	if req.Name != "" {
		game.Name = req.Name
	}
	if req.Description != "" {
		game.Description = &req.Description
	}
	if req.Platform != "" {
		game.Platform = &req.Platform
	}
	if req.Stock > 0 {
		diff := req.Stock - game.Stock
		game.Stock = req.Stock
		game.AvailableStock += diff
	}
	if req.RentalPricePerDay > 0 {
		game.RentalPricePerDay = req.RentalPricePerDay
	}
	if req.SecurityDeposit > 0 {
		game.SecurityDeposit = req.SecurityDeposit
	}
	if req.Condition != "" {
		game.Condition = model.GameCondition(req.Condition)
	}

	adminID := echomw.CurrentUserID(c)
	role := echomw.CurrentRole(c)

	err = h.gameService.Update(adminID, model.UserRole(role), gameID, game)
	if err != nil {
		return myResponse.BadRequest(c, err.Error())
	}

	return myResponse.Success(c, "Game updated successfully", game)
}

// DeleteGame godoc
// @Summary Delete game
// @Description Delete a game (Admin only)
// @Tags Admin - Games
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Game ID"
// @Success 200 {object} map[string]interface{} "Game deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid game ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /admin/games/{id} [delete]
func (h *GameHandler) DeleteGame(c echo.Context) error {
	gameID := myRequest.PathParamUint(c, "id")
	if gameID == 0 {
		return myResponse.BadRequest(c, "Invalid game ID")
	}

	role := echomw.CurrentRole(c)
	err := h.gameService.Delete(model.UserRole(role), gameID)
	if err != nil {
		return myResponse.Forbidden(c, err.Error())
	}

	return myResponse.Success(c, "Game deleted successfully", nil)
}
