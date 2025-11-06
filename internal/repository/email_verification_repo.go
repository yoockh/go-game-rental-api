package repository

import (
	"github.com/Yoochan45/go-game-rental-api/internal/model"
	"gorm.io/gorm"
)

type EmailVerificationRepository interface {
	Create(token *model.EmailVerificationToken) error
	GetByTokenHash(tokenHash string) (*model.EmailVerificationToken, error)
	MarkAsUsed(tokenID uint) error
	DeleteExpiredTokens() error
}

type emailVerificationRepository struct {
	db *gorm.DB
}

func NewEmailVerificationRepository(db *gorm.DB) EmailVerificationRepository {
	return &emailVerificationRepository{db: db}
}

func (r *emailVerificationRepository) Create(token *model.EmailVerificationToken) error {
	return r.db.Create(token).Error
}

func (r *emailVerificationRepository) GetByTokenHash(tokenHash string) (*model.EmailVerificationToken, error) {
	var token model.EmailVerificationToken
	err := r.db.Preload("User").Where("token_hash = ? AND is_used = false", tokenHash).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *emailVerificationRepository) MarkAsUsed(tokenID uint) error {
	return r.db.Model(&model.EmailVerificationToken{}).Where("id = ?", tokenID).Update("is_used", true).Error
}

func (r *emailVerificationRepository) DeleteExpiredTokens() error {
	return r.db.Where("expires_at < NOW() OR is_used = true").Delete(&model.EmailVerificationToken{}).Error
}