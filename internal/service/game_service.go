package service

import (
	"errors"

	"github.com/yoockh/go-game-rental-api/internal/model"
	"github.com/yoockh/go-game-rental-api/internal/repository"
)

var (
	ErrGameNotFound               = errors.New("game not found")
	ErrGameInsufficientPermission = errors.New("insufficient permission")
	ErrGameNotOwned               = errors.New("you don't own this game")
)

type GameService interface {
	// Public
	GetAll(limit, offset int) ([]*model.Game, int64, error)
	Search(query string, limit, offset int) ([]*model.Game, error)
	GetByID(gameID uint) (*model.Game, error)

	// Admin
	Create(adminID uint, requestorRole model.UserRole, gameData *model.Game) error
	Update(adminID uint, requestorRole model.UserRole, gameID uint, updateData *model.Game) error
	Delete(requestorRole model.UserRole, gameID uint) error
}

type gameService struct {
	gameRepo repository.GameRepository
}

func NewGameService(gameRepo repository.GameRepository) GameService {
	return &gameService{gameRepo: gameRepo}
}

func (s *gameService) GetAll(limit, offset int) ([]*model.Game, int64, error) {
	games, err := s.gameRepo.GetAll(limit, offset)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.gameRepo.Count()
	return games, count, err
}

func (s *gameService) Search(query string, limit, offset int) ([]*model.Game, error) {
	return s.gameRepo.Search(query, limit, offset)
}

func (s *gameService) GetByID(gameID uint) (*model.Game, error) {
	return s.gameRepo.GetByID(gameID)
}

func (s *gameService) Create(adminID uint, requestorRole model.UserRole, gameData *model.Game) error {
	if !s.canManageGames(requestorRole) {
		return ErrGameInsufficientPermission
	}

	gameData.AdminID = adminID
	gameData.IsActive = true
	gameData.AvailableStock = gameData.Stock

	return s.gameRepo.Create(gameData)
}

func (s *gameService) Update(adminID uint, requestorRole model.UserRole, gameID uint, updateData *model.Game) error {
	if !s.canManageGames(requestorRole) {
		return ErrGameInsufficientPermission
	}

	game, err := s.gameRepo.GetByID(gameID)
	if err != nil {
		return ErrGameNotFound
	}

	// Admin can only edit their own games (super_admin can edit all)
	if requestorRole != model.RoleSuperAdmin && game.AdminID != adminID {
		return ErrGameNotOwned
	}

	game.Name = updateData.Name
	game.Description = updateData.Description
	game.Platform = updateData.Platform
	game.Stock = updateData.Stock
	game.RentalPricePerDay = updateData.RentalPricePerDay
	game.SecurityDeposit = updateData.SecurityDeposit
	game.Condition = updateData.Condition
	game.CategoryID = updateData.CategoryID

	return s.gameRepo.Update(game)
}

func (s *gameService) Delete(requestorRole model.UserRole, gameID uint) error {
	if !s.canManageGames(requestorRole) {
		return ErrGameInsufficientPermission
	}

	return s.gameRepo.Delete(gameID)
}

func (s *gameService) canManageGames(role model.UserRole) bool {
	return role == model.RoleAdmin || role == model.RoleSuperAdmin
}
