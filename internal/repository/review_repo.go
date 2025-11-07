package repository

import (
	"github.com/Yoochan45/go-game-rental-api/internal/model"
	"gorm.io/gorm"
)

type ReviewRepository interface {
	// Basic CRUD
	Create(review *model.Review) error

	// Query methods
	GetByBookingID(bookingID uint) (*model.Review, error)
	GetGameReviews(gameID uint, limit, offset int) ([]*model.Review, error)
}

type reviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) Create(review *model.Review) error {
	return r.db.Create(review).Error
}

func (r *reviewRepository) GetByBookingID(bookingID uint) (*model.Review, error) {
	var review model.Review
	err := r.db.Where("booking_id = ?", bookingID).First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *reviewRepository) GetGameReviews(gameID uint, limit, offset int) ([]*model.Review, error) {
	var reviews []*model.Review
	err := r.db.
		Preload("User").
		Preload("Booking").
		Preload("Game").
		Where("game_id = ?", gameID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&reviews).Error
	return reviews, err
}
