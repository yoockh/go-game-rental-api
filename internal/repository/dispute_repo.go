package repository

import (
	"github.com/Yoochan45/go-game-rental-api/internal/model"
	"gorm.io/gorm"
)

type DisputeRepository interface {
	// Basic CRUD
	Create(dispute *model.Dispute) error
	GetByID(id uint) (*model.Dispute, error)
	GetByIDWithRelations(id uint) (*model.Dispute, error)
	Update(dispute *model.Dispute) error

	// Query methods
	GetUserDisputes(userID uint, limit, offset int) ([]*model.Dispute, error)
	GetDisputesByStatus(status model.DisputeStatus, limit, offset int) ([]*model.Dispute, error)
	GetDisputesByType(disputeType model.DisputeType, limit, offset int) ([]*model.Dispute, error)
	GetBookingDisputes(bookingID uint) ([]*model.Dispute, error)

	// Admin methods
	GetPendingDisputes(limit, offset int) ([]*model.Dispute, error)
	GetAllDisputes(limit, offset int) ([]*model.Dispute, error)
	UpdateDisputeStatus(disputeID uint, status model.DisputeStatus, resolvedBy *uint, resolution *string) error

	// Statistics
	CountByStatus(status model.DisputeStatus) (int64, error)
	CountByType(disputeType model.DisputeType) (int64, error)
}

type disputeRepository struct {
	db *gorm.DB
}

func NewDisputeRepository(db *gorm.DB) DisputeRepository {
	return &disputeRepository{db: db}
}

func (r *disputeRepository) Create(dispute *model.Dispute) error {
	return r.db.Create(dispute).Error
}

func (r *disputeRepository) GetByID(id uint) (*model.Dispute, error) {
	var dispute model.Dispute
	err := r.db.Where("id = ?", id).First(&dispute).Error
	if err != nil {
		return nil, err
	}
	return &dispute, nil
}

func (r *disputeRepository) GetByIDWithRelations(id uint) (*model.Dispute, error) {
	var dispute model.Dispute
	err := r.db.Preload("Booking").Preload("Reporter").Preload("Resolver").
		Where("id = ?", id).First(&dispute).Error
	if err != nil {
		return nil, err
	}
	return &dispute, nil
}

func (r *disputeRepository) Update(dispute *model.Dispute) error {
	return r.db.Save(dispute).Error
}

func (r *disputeRepository) GetUserDisputes(userID uint, limit, offset int) ([]*model.Dispute, error) {
	var disputes []*model.Dispute
	err := r.db.Preload("Booking").Preload("Resolver").
		Where("reporter_id = ?", userID).Order("created_at DESC").
		Limit(limit).Offset(offset).Find(&disputes).Error
	return disputes, err
}

func (r *disputeRepository) GetDisputesByStatus(status model.DisputeStatus, limit, offset int) ([]*model.Dispute, error) {
	var disputes []*model.Dispute
	err := r.db.Preload("Booking").Preload("Reporter").Preload("Resolver").
		Where("status = ?", status).Order("created_at DESC").
		Limit(limit).Offset(offset).Find(&disputes).Error
	return disputes, err
}

func (r *disputeRepository) GetDisputesByType(disputeType model.DisputeType, limit, offset int) ([]*model.Dispute, error) {
	var disputes []*model.Dispute
	err := r.db.Preload("Booking").Preload("Reporter").Preload("Resolver").
		Where("type = ?", disputeType).Order("created_at DESC").
		Limit(limit).Offset(offset).Find(&disputes).Error
	return disputes, err
}

func (r *disputeRepository) GetBookingDisputes(bookingID uint) ([]*model.Dispute, error) {
	var disputes []*model.Dispute
	err := r.db.Preload("Reporter").Preload("Resolver").
		Where("booking_id = ?", bookingID).Order("created_at DESC").Find(&disputes).Error
	return disputes, err
}

func (r *disputeRepository) GetPendingDisputes(limit, offset int) ([]*model.Dispute, error) {
	var disputes []*model.Dispute
	err := r.db.Preload("Booking").Preload("Reporter").
		Where("status = ?", model.DisputeOpen).Order("created_at ASC").
		Limit(limit).Offset(offset).Find(&disputes).Error
	return disputes, err
}

func (r *disputeRepository) GetAllDisputes(limit, offset int) ([]*model.Dispute, error) {
	var disputes []*model.Dispute
	err := r.db.Preload("Booking").Preload("Reporter").Preload("Resolver").
		Order("created_at DESC").Limit(limit).Offset(offset).Find(&disputes).Error
	return disputes, err
}

func (r *disputeRepository) UpdateDisputeStatus(disputeID uint, status model.DisputeStatus, resolvedBy *uint, resolution *string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if status == model.DisputeResolved || status == model.DisputeClosed {
		updates["resolved_by"] = resolvedBy
		updates["resolution"] = resolution
		updates["resolved_at"] = gorm.Expr("CURRENT_TIMESTAMP")
	}

	return r.db.Model(&model.Dispute{}).Where("id = ?", disputeID).Updates(updates).Error
}

func (r *disputeRepository) CountByStatus(status model.DisputeStatus) (int64, error) {
	var count int64
	err := r.db.Model(&model.Dispute{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

func (r *disputeRepository) CountByType(disputeType model.DisputeType) (int64, error) {
	var count int64
	err := r.db.Model(&model.Dispute{}).Where("type = ?", disputeType).Count(&count).Error
	return count, err
}
