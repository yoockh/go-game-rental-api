package service

import (
	"errors"

	"github.com/yoockh/go-game-rental-api/internal/model"
	"github.com/yoockh/go-game-rental-api/internal/repository"
)

var (
	ErrReviewAlreadyExists       = errors.New("review already exists for this booking")
	ErrReviewBookingNotCompleted = errors.New("can only review completed bookings")
)

type ReviewService interface {
	// Customer methods
	CreateReview(userID uint, bookingID uint, reviewData *model.Review) error

	// Public methods
	GetGameReviews(gameID uint, limit, offset int) ([]*model.Review, error)
}

type reviewService struct {
	reviewRepo  repository.ReviewRepository
	bookingRepo repository.BookingRepository
}

func NewReviewService(reviewRepo repository.ReviewRepository, bookingRepo repository.BookingRepository) ReviewService {
	return &reviewService{
		reviewRepo:  reviewRepo,
		bookingRepo: bookingRepo,
	}
}

func (s *reviewService) CreateReview(userID uint, bookingID uint, reviewData *model.Review) error {
	// Validate booking exists and belongs to user
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return ErrBookingNotFound
	}

	if booking.UserID != userID {
		return ErrBookingNotOwned
	}

	// Can only review completed bookings
	if booking.Status != model.BookingCompleted {
		return ErrReviewBookingNotCompleted
	}

	// Check if review already exists
	existingReview, _ := s.reviewRepo.GetByBookingID(bookingID)
	if existingReview != nil {
		return ErrReviewAlreadyExists
	}

	// Set review details
	reviewData.BookingID = bookingID
	reviewData.UserID = userID
	reviewData.GameID = booking.GameID

	return s.reviewRepo.Create(reviewData)
}

func (s *reviewService) GetGameReviews(gameID uint, limit, offset int) ([]*model.Review, error) {
	return s.reviewRepo.GetGameReviews(gameID, limit, offset)
}
