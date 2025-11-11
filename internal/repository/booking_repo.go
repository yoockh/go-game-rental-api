package repository

import (
	"github.com/yoockh/go-game-rental-api/internal/model"
	"gorm.io/gorm"
)

type BookingRepository interface {
	// Basic CRUD
	Create(booking *model.Booking) error
	GetByID(id uint) (*model.Booking, error)
	Update(booking *model.Booking) error

	// Query methods
	GetUserBookings(userID uint, limit, offset int) ([]*model.Booking, error)
	GetAllBookings(limit, offset int) ([]*model.Booking, error)
	CountUserBookings(userID uint) (int64, error)
	Count() (int64, error)

	// Status updates
	UpdateStatus(bookingID uint, status model.BookingStatus) error
}

type bookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) BookingRepository {
	return &bookingRepository{db: db}
}

func (r *bookingRepository) Create(booking *model.Booking) error {
	return r.db.Create(booking).Error
}

func (r *bookingRepository) GetByID(id uint) (*model.Booking, error) {
	var booking model.Booking
	if err := r.db.Preload("User").Preload("Game").Preload("Payment").First(&booking, id).Error; err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *bookingRepository) Update(booking *model.Booking) error {
	return r.db.Save(booking).Error
}

func (r *bookingRepository) GetUserBookings(userID uint, limit, offset int) ([]*model.Booking, error) {
	var bookings []*model.Booking
	err := r.db.Where("user_id = ?", userID).Preload("Game").Preload("Payment").
		Order("created_at DESC").Limit(limit).Offset(offset).Find(&bookings).Error
	return bookings, err
}

func (r *bookingRepository) GetAllBookings(limit, offset int) ([]*model.Booking, error) {
	var bookings []*model.Booking
	err := r.db.Preload("User").Preload("Game").Preload("Payment").
		Order("created_at DESC").Limit(limit).Offset(offset).Find(&bookings).Error
	return bookings, err
}

func (r *bookingRepository) CountUserBookings(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Booking{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

func (r *bookingRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.Booking{}).Count(&count).Error
	return count, err
}

func (r *bookingRepository) UpdateStatus(bookingID uint, status model.BookingStatus) error {
	return r.db.Model(&model.Booking{}).Where("id = ?", bookingID).Update("status", status).Error
}
