package handler

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	echomw "github.com/Yoochan45/go-api-utils/pkg-echo/middleware"
	myRequest "github.com/Yoochan45/go-api-utils/pkg-echo/request"
	myResponse "github.com/Yoochan45/go-api-utils/pkg-echo/response"
	"github.com/Yoochan45/go-game-rental-api/internal/dto"
	"github.com/Yoochan45/go-game-rental-api/internal/integration/storage"
	"github.com/Yoochan45/go-game-rental-api/internal/model"
	"github.com/Yoochan45/go-game-rental-api/internal/service"
	"github.com/Yoochan45/go-game-rental-api/internal/utils"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
)

type GameHandler struct {
	gameService   service.GameService
	validate      *validator.Validate
	storageClient storage.StorageClient
}

func NewGameHandler(gameService service.GameService, storageClient storage.StorageClient) *GameHandler {
	return &GameHandler{
		gameService:   gameService,
		validate:      utils.GetValidator(),
		storageClient: storageClient,
	}
}

// GetAllGames godoc
// @Summary Get all games
// @Description Get list of all available games (public)
// @Tags Games
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{} "Games retrieved successfully"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /games [get]
func (h *GameHandler) GetAllGames(c echo.Context) error {
	params := utils.ParsePagination(c)

	games, err := h.gameService.GetPublicGames(params.Limit, params.Offset)
	if err != nil {
		return myResponse.InternalServerError(c, "Failed to retrieve games")
	}

	meta := utils.CreateMeta(params, int64(len(games)))
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

	game, err := h.gameService.GetGameDetail(gameID)
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

	games, err := h.gameService.SearchGames(query, params.Limit, params.Offset)
	if err != nil {
		return myResponse.InternalServerError(c, "Failed to search games")
	}

	meta := utils.CreateMeta(params, int64(len(games)))
	return myResponse.Paginated(c, "Games search results", games, meta)
}

// CreateGame godoc
// @Summary Create new game
// @Description Create a new game listing (Partner only)
// @Tags Partner
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateGameRequest true "Game details"
// @Success 201 {object} map[string]interface{} "Game created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /partner/games [post]
func (h *GameHandler) CreateGame(c echo.Context) error {
	var req dto.CreateGameRequest
	if err := c.Bind(&req); err != nil {
		return myResponse.BadRequest(c, "Invalid input: "+err.Error())
	}
	if err := h.validate.Struct(&req); err != nil {
		return myResponse.BadRequest(c, "Validation error: "+err.Error())
	}

	partnerID := echomw.CurrentUserID(c)

	gameData := &model.Game{
		PartnerID:         partnerID,
		CategoryID:        req.CategoryID,
		Name:              req.Name,
		Description:       utils.PtrOrNil(req.Description),
		Platform:          utils.PtrOrNil(req.Platform),
		Stock:             req.Stock,
		RentalPricePerDay: req.RentalPricePerDay,
		SecurityDeposit:   req.SecurityDeposit,
		Condition:         req.Condition,
		Images:            pq.StringArray(req.Images),
	}

	err := h.gameService.CreatePartnerGame(partnerID, gameData)
	if err != nil {
		return myResponse.Forbidden(c, err.Error())
	}

	return myResponse.Created(c, "Game created successfully", nil)
}

// UpdateGame godoc
// @Summary Update game
// @Description Update game information (Partner only)
// @Tags Partner
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Game ID"
// @Param request body dto.UpdateGameRequest true "Updated game details"
// @Success 200 {object} map[string]interface{} "Game updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /partner/games/{id} [put]
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

	partnerID := echomw.CurrentUserID(c)

	updateData := &model.Game{
		CategoryID:        req.CategoryID,
		Name:              req.Name,
		Description:       utils.PtrOrNil(req.Description),
		Platform:          utils.PtrOrNil(req.Platform),
		RentalPricePerDay: req.RentalPricePerDay,
		SecurityDeposit:   req.SecurityDeposit,
		Condition:         req.Condition,
		Images:            pq.StringArray(req.Images),
	}

	err := h.gameService.UpdatePartnerGame(partnerID, gameID, updateData)
	if err != nil {
		return myResponse.Forbidden(c, err.Error())
	}

	return myResponse.Success(c, "Game updated successfully", nil)
}

// GetPartnerGames godoc
// @Summary Get partner's games
// @Description Get list of games owned by the partner
// @Tags Partner
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{} "Partner games retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /partner/games [get]
func (h *GameHandler) GetPartnerGames(c echo.Context) error {
	userID := echomw.CurrentUserID(c)
	if userID == 0 {
		return myResponse.Unauthorized(c, "Unauthorized")
	}

	params := utils.ParsePagination(c)

	games, err := h.gameService.GetPartnerGames(userID, params.Limit, params.Offset)
	if err != nil {
		return myResponse.InternalServerError(c, "Failed to retrieve games")
	}

	meta := utils.CreateMeta(params, int64(len(games)))
	return myResponse.Paginated(c, "Partner games retrieved successfully", games, meta)
}

// UploadGameImage godoc
// @Summary Upload game image
// @Description Upload image for a game (Partner only)
// @Tags Partner
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path int true "Game ID"
// @Param image formData file true "Game image file"
// @Success 200 {object} map[string]interface{} "Image uploaded successfully"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /partner/games/{id}/upload-image [post]
func (h *GameHandler) UploadGameImage(c echo.Context) error {
	gameID := myRequest.PathParamUint(c, "id")
	if gameID == 0 {
		return myResponse.BadRequest(c, "Invalid game ID")
	}

	partnerID := echomw.CurrentUserID(c)

	// Cek ownership game
	game, err := h.gameService.GetGameDetail(gameID)
	if err != nil {
		return myResponse.NotFound(c, "Game not found")
	}
	if game.PartnerID != partnerID {
		return myResponse.Forbidden(c, "You can only upload images for your own games")
	}

	// Get uploaded file
	file, err := c.FormFile("image")
	if err != nil {
		return myResponse.BadRequest(c, "No image file provided")
	}

	// Validate file type
	if !strings.HasPrefix(file.Header.Get("Content-Type"), "image/") {
		return myResponse.BadRequest(c, "File must be an image")
	}

	// Validate file size (max 5MB)
	if file.Size > 5*1024*1024 {
		return myResponse.BadRequest(c, "File size must be less than 5MB")
	}

	// Open file
	src, err := file.Open()
	if err != nil {
		return myResponse.InternalServerError(c, "Failed to open file")
	}
	defer src.Close()

	// Read file data
	fileData, err := io.ReadAll(src)
	if err != nil {
		return myResponse.InternalServerError(c, "Failed to read file")
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("game-%d-%d%s", gameID, time.Now().Unix(), ext)
	destinationPath := fmt.Sprintf("games/%s", filename)

	// Upload to storage
	imageURL, err := h.storageClient.UploadFile(
		c.Request().Context(),
		destinationPath,
		filename,
		file.Header.Get("Content-Type"),
		fileData,
	)
	if err != nil {
		return myResponse.InternalServerError(c, "Failed to upload image: "+err.Error())
	}

	return myResponse.Success(c, "Image uploaded successfully", map[string]interface{}{
		"image_url": imageURL,
		"filename":  filename,
	})
}
