package repository

import (
	"github.com/yoockh/go-game-rental-api/internal/model"
	"gorm.io/gorm"
)

type PaymentRepository interface {
	// Basic CRUD
	Create(payment *model.Payment) error
	GetByID(id uint) (*model.Payment, error)
	GetByIDWithRelations(id uint) (*model.Payment, error)
	Update(payment *model.Payment) error

	// Query methods
	GetByBookingID(bookingID uint) (*model.Payment, error)
	GetByProviderPaymentID(providerPaymentID string) (*model.Payment, error)
	GetPaymentsByStatus(status model.PaymentStatus, limit, offset int) ([]*model.Payment, error)
	GetAllPayments(limit, offset int) ([]*model.Payment, error)
	CountAllPayments() (int64, error)
	CountByStatus(status model.PaymentStatus) (int64, error)

	// Status updates
	MarkAsPaid(paymentID uint, providerPaymentID string, paymentMethod string) error
	MarkAsFailed(paymentID uint, failureReason string) error
}

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(payment *model.Payment) error {
	return r.db.Create(payment).Error
}

func (r *paymentRepository) GetByID(id uint) (*model.Payment, error) {
	var payment model.Payment
	err := r.db.Where("id = ?", id).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) GetByIDWithRelations(id uint) (*model.Payment, error) {
	var payment model.Payment
	err := r.db.Preload("Booking").Preload("Booking.User").Preload("Booking.Game").
		Where("id = ?", id).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) Update(payment *model.Payment) error {
	return r.db.Save(payment).Error
}

func (r *paymentRepository) GetByBookingID(bookingID uint) (*model.Payment, error) {
	var payment model.Payment
	err := r.db.Where("booking_id = ?", bookingID).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) GetByProviderPaymentID(providerPaymentID string) (*model.Payment, error) {
	var payment model.Payment
	err := r.db.Where("provider_payment_id = ?", providerPaymentID).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) GetPaymentsByStatus(status model.PaymentStatus, limit, offset int) ([]*model.Payment, error) {
	var payments []*model.Payment
	err := r.db.Preload("Booking").Where("status = ?", status).
		Order("created_at DESC").Limit(limit).Offset(offset).Find(&payments).Error
	return payments, err
}

func (r *paymentRepository) MarkAsPaid(paymentID uint, providerPaymentID string, paymentMethod string) error {
	return r.db.Model(&model.Payment{}).Where("id = ?", paymentID).Updates(map[string]interface{}{
		"status":              model.PaymentPaid,
		"provider_payment_id": providerPaymentID,
		"payment_method":      paymentMethod,
		"paid_at":             gorm.Expr("CURRENT_TIMESTAMP"),
	}).Error
}

func (r *paymentRepository) MarkAsFailed(paymentID uint, failureReason string) error {
	return r.db.Model(&model.Payment{}).Where("id = ?", paymentID).Updates(map[string]interface{}{
		"status":         model.PaymentFailed,
		"failure_reason": failureReason,
		"failed_at":      gorm.Expr("CURRENT_TIMESTAMP"),
	}).Error
}

func (r *paymentRepository) GetAllPayments(limit, offset int) ([]*model.Payment, error) {
	var payments []*model.Payment
	err := r.db.Preload("Booking").Order("created_at DESC").
		Limit(limit).Offset(offset).Find(&payments).Error
	return payments, err
}

func (r *paymentRepository) CountAllPayments() (int64, error) {
	var count int64
	err := r.db.Model(&model.Payment{}).Count(&count).Error
	return count, err
}

func (r *paymentRepository) CountByStatus(status model.PaymentStatus) (int64, error) {
	var count int64
	err := r.db.Model(&model.Payment{}).Where("status = ?", status).Count(&count).Error
	return count, err
}
