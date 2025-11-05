package service

import (
	"errors"

	"github.com/Yoochan45/go-game-rental-api/internal/model"
	"github.com/Yoochan45/go-game-rental-api/internal/repository"
)

var (
	ErrGameNotFound               = errors.New("game not found")
	ErrGameInsufficientPermission = errors.New("insufficient permission")
	ErrGameAlreadyApproved        = errors.New("game already approved")
	ErrGameNotOwned               = errors.New("you don't own this game")
)

type GameService interface {
	// Public methods (customer)
	GetPublicGames(limit, offset int) ([]*model.Game, error)
	SearchGames(query string, limit, offset int) ([]*model.Game, error)
	GetGameDetail(gameID uint) (*model.Game, error)

	// Partner methods
	CreatePartnerGame(partnerID uint, gameData *model.Game) error
	UpdatePartnerGame(partnerID uint, gameID uint, updateData *model.Game) error
	GetPartnerGames(partnerID uint, limit, offset int) ([]*model.Game, error)

	// Admin methods
	ApproveGame(adminID uint, requestorRole model.UserRole, gameID uint) error
	GetAllGames(requestorRole model.UserRole, limit, offset int) ([]*model.Game, error)
}

type gameService struct {
	gameRepo repository.GameRepository
	userRepo repository.UserRepository
}

func NewGameService(gameRepo repository.GameRepository, userRepo repository.UserRepository) GameService {
	return &gameService{
		gameRepo: gameRepo,
		userRepo: userRepo,
	}
}

// Public methods (customer)
func (s *gameService) GetPublicGames(limit, offset int) ([]*model.Game, error) {
	return s.gameRepo.GetApprovedGames(limit, offset)
}

func (s *gameService) SearchGames(query string, limit, offset int) ([]*model.Game, error) {
	return s.gameRepo.SearchGames(query, limit, offset)
}

func (s *gameService) GetGameDetail(gameID uint) (*model.Game, error) {
	game, err := s.gameRepo.GetByIDWithRelations(gameID)
	if err != nil {
		return nil, ErrGameNotFound
	}

	// Only show approved and active games to public
	if game.ApprovalStatus != model.ApprovalApproved || !game.IsActive {
		return nil, ErrGameNotFound
	}

	return game, nil
}

// Partner methods
func (s *gameService) CreatePartnerGame(partnerID uint, gameData *model.Game) error {
	// Verify user is partner
	user, err := s.userRepo.GetByID(partnerID)
	if err != nil {
		return ErrUserNotFound
	}

	if user.Role != model.RolePartner {
		return ErrGameInsufficientPermission
	}

	// Set partner ID and default values
	gameData.PartnerID = partnerID
	gameData.IsActive = false
	gameData.ApprovalStatus = model.ApprovalPending
	gameData.AvailableStock = gameData.Stock

	return s.gameRepo.Create(gameData)
}

func (s *gameService) UpdatePartnerGame(partnerID uint, gameID uint, updateData *model.Game) error {
	// Check if game exists and belongs to partner
	game, err := s.gameRepo.GetByID(gameID)
	if err != nil {
		return ErrGameNotFound
	}

	if game.PartnerID != partnerID {
		return ErrGameNotOwned
	}

	// Partner can only update if not approved yet or rejected
	if game.ApprovalStatus == model.ApprovalApproved {
		return ErrGameAlreadyApproved
	}

	// Update allowed fields
	game.Name = updateData.Name
	game.Description = updateData.Description
	game.Platform = updateData.Platform
	game.RentalPricePerDay = updateData.RentalPricePerDay
	game.SecurityDeposit = updateData.SecurityDeposit
	game.Condition = updateData.Condition
	game.Images = updateData.Images
	game.CategoryID = updateData.CategoryID

	// Reset approval status if updating rejected game
	if game.ApprovalStatus == model.ApprovalRejected {
		game.ApprovalStatus = model.ApprovalPending
		game.RejectionReason = nil
	}

	return s.gameRepo.Update(game)
}

func (s *gameService) GetPartnerGames(partnerID uint, limit, offset int) ([]*model.Game, error) {
	return s.gameRepo.GetGamesByPartner(partnerID, limit, offset)
}

// Admin methods
func (s *gameService) ApproveGame(adminID uint, requestorRole model.UserRole, gameID uint) error {
	if !s.canManageGames(requestorRole) {
		return ErrGameInsufficientPermission
	}

	game, err := s.gameRepo.GetByID(gameID)
	if err != nil {
		return ErrGameNotFound
	}

	if game.ApprovalStatus != model.ApprovalPending {
		return ErrGameAlreadyApproved
	}

	return s.gameRepo.UpdateApprovalStatus(gameID, model.ApprovalApproved, &adminID, nil)
}

func (s *gameService) GetAllGames(requestorRole model.UserRole, limit, offset int) ([]*model.Game, error) {
	if !s.canManageGames(requestorRole) {
		return nil, ErrGameInsufficientPermission
	}

	return s.gameRepo.GetAllGames(limit, offset)
}

// Helper methods
func (s *gameService) canManageGames(role model.UserRole) bool {
	return role == model.RoleAdmin || role == model.RoleSuperAdmin
}
