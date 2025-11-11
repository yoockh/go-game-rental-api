package repository

import (
	"github.com/yoockh/go-game-rental-api/internal/model"
	"gorm.io/gorm"
)

type GameRepository interface {
	// Basic CRUD
	Create(game *model.Game) error
	GetByID(id uint) (*model.Game, error)
	Update(game *model.Game) error
	Delete(id uint) error

	// Query methods for public catalog
	GetAll(limit, offset int) ([]*model.Game, error)
	Search(query string, limit, offset int) ([]*model.Game, error)
	Count() (int64, error)

	// Stock management
	CheckAvailability(gameID uint) (bool, error)
	ReserveStock(gameID uint) error
	ReleaseStock(gameID uint) error
}

type gameRepository struct {
	db *gorm.DB
}

func NewGameRepository(db *gorm.DB) GameRepository {
	return &gameRepository{db: db}
}

func (r *gameRepository) Create(game *model.Game) error {
	return r.db.Create(game).Error
}

func (r *gameRepository) GetByID(id uint) (*model.Game, error) {
	var game model.Game
	if err := r.db.Preload("Admin").Preload("Category").First(&game, id).Error; err != nil {
		return nil, err
	}
	return &game, nil
}

func (r *gameRepository) Update(game *model.Game) error {
	return r.db.Save(game).Error
}

func (r *gameRepository) Delete(id uint) error {
	return r.db.Delete(&model.Game{}, id).Error
}

func (r *gameRepository) GetAll(limit, offset int) ([]*model.Game, error) {
	var games []*model.Game
	// Tidak perlu Session lagi, sudah global
	err := r.db.
		Where("is_active = ?", true).
		Preload("Admin").
		Preload("Category").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&games).Error
	return games, err
}

func (r *gameRepository) Search(query string, limit, offset int) ([]*model.Game, error) {
	var games []*model.Game
	searchPattern := "%" + query + "%"
	err := r.db.Session(&gorm.Session{PrepareStmt: false}).
		Where("is_active = ? AND (name ILIKE ? OR description ILIKE ? OR platform ILIKE ?)",
			true, searchPattern, searchPattern, searchPattern).
		Preload("Admin").
		Preload("Category").
		Limit(limit).Offset(offset).
		Find(&games).Error
	return games, err
}

func (r *gameRepository) Count() (int64, error) {
	var count int64
	err := r.db.Session(&gorm.Session{PrepareStmt: false}).
		Model(&model.Game{}).
		Where("is_active = ?", true).
		Count(&count).Error
	return count, err
}

func (r *gameRepository) CheckAvailability(gameID uint) (bool, error) {
	var game model.Game
	if err := r.db.Select("available_stock").First(&game, gameID).Error; err != nil {
		return false, err
	}
	return game.AvailableStock > 0, nil
}

func (r *gameRepository) ReserveStock(gameID uint) error {
	return r.db.Model(&model.Game{}).Where("id = ? AND available_stock > 0", gameID).
		Update("available_stock", gorm.Expr("available_stock - 1")).Error
}

func (r *gameRepository) ReleaseStock(gameID uint) error {
	return r.db.Model(&model.Game{}).Where("id = ?", gameID).
		Update("available_stock", gorm.Expr("LEAST(available_stock + 1, stock)")).Error
}
