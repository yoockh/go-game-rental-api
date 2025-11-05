package repository

import (
	"github.com/Yoochan45/go-game-rental-api/internal/model"
	"gorm.io/gorm"
)

type PartnerApplicationRepository interface {
	// Basic CRUD
	Create(application *model.PartnerApplication) error
	GetByID(id uint) (*model.PartnerApplication, error)
	GetByIDWithRelations(id uint) (*model.PartnerApplication, error)
	Update(application *model.PartnerApplication) error

	// Query methods
	GetByUserID(userID uint) (*model.PartnerApplication, error)
	GetPendingApplications(limit, offset int) ([]*model.PartnerApplication, error)
	GetAllApplications(limit, offset int) ([]*model.PartnerApplication, error)
	GetApplicationsByStatus(status model.ApplicationStatus, limit, offset int) ([]*model.PartnerApplication, error)

	// Admin methods
	UpdateApplicationStatus(applicationID uint, status model.ApplicationStatus, decidedBy uint, rejectionReason *string) error

	// Statistics
	CountByStatus(status model.ApplicationStatus) (int64, error)
}

type partnerApplicationRepository struct {
	db *gorm.DB
}

func NewPartnerApplicationRepository(db *gorm.DB) PartnerApplicationRepository {
	return &partnerApplicationRepository{db: db}
}

func (r *partnerApplicationRepository) Create(application *model.PartnerApplication) error {
	return r.db.Create(application).Error
}

func (r *partnerApplicationRepository) GetByID(id uint) (*model.PartnerApplication, error) {
	var application model.PartnerApplication
	err := r.db.Where("id = ?", id).First(&application).Error
	if err != nil {
		return nil, err
	}
	return &application, nil
}

func (r *partnerApplicationRepository) GetByIDWithRelations(id uint) (*model.PartnerApplication, error) {
	var application model.PartnerApplication
	err := r.db.Preload("User").Preload("Decider").Where("id = ?", id).First(&application).Error
	if err != nil {
		return nil, err
	}
	return &application, nil
}

func (r *partnerApplicationRepository) Update(application *model.PartnerApplication) error {
	return r.db.Save(application).Error
}

func (r *partnerApplicationRepository) GetByUserID(userID uint) (*model.PartnerApplication, error) {
	var application model.PartnerApplication
	err := r.db.Where("user_id = ?", userID).First(&application).Error
	if err != nil {
		return nil, err
	}
	return &application, nil
}

func (r *partnerApplicationRepository) GetPendingApplications(limit, offset int) ([]*model.PartnerApplication, error) {
	var applications []*model.PartnerApplication
	err := r.db.Preload("User").Where("status = ?", model.ApplicationPending).
		Limit(limit).Offset(offset).Find(&applications).Error
	return applications, err
}

func (r *partnerApplicationRepository) GetAllApplications(limit, offset int) ([]*model.PartnerApplication, error) {
	var applications []*model.PartnerApplication
	err := r.db.Preload("User").Preload("Decider").
		Limit(limit).Offset(offset).Find(&applications).Error
	return applications, err
}

func (r *partnerApplicationRepository) GetApplicationsByStatus(status model.ApplicationStatus, limit, offset int) ([]*model.PartnerApplication, error) {
	var applications []*model.PartnerApplication
	err := r.db.Preload("User").Preload("Decider").Where("status = ?", status).
		Limit(limit).Offset(offset).Find(&applications).Error
	return applications, err
}

func (r *partnerApplicationRepository) UpdateApplicationStatus(applicationID uint, status model.ApplicationStatus, decidedBy uint, rejectionReason *string) error {
	updates := map[string]interface{}{
		"status":     status,
		"decided_by": decidedBy,
		"decided_at": gorm.Expr("CURRENT_TIMESTAMP"),
	}

	if status == model.ApplicationRejected && rejectionReason != nil {
		updates["rejection_reason"] = *rejectionReason
	}

	return r.db.Model(&model.PartnerApplication{}).Where("id = ?", applicationID).Updates(updates).Error
}

func (r *partnerApplicationRepository) CountByStatus(status model.ApplicationStatus) (int64, error) {
	var count int64
	err := r.db.Model(&model.PartnerApplication{}).Where("status = ?", status).Count(&count).Error
	return count, err
}
