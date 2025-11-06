package repository

import (
	"github.com/Yoochan45/go-game-rental-api/internal/model"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *model.User) error
	GetByID(id uint) (*model.User, error)
	GetByEmail(email string) (*model.User, error)
	Update(user *model.User) error
	Delete(id uint) error

	GetAll(limit, offset int) ([]*model.User, error)
	UpdateRole(userID uint, newRole model.UserRole) error
	UpdateActiveStatus(userID uint, isActive bool) error
	Count() (int64, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetByID(id uint) (*model.User, error) {
	var user model.User
	if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(email string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id uint) error {
	return r.db.Unscoped().Delete(&model.User{}, id).Error
}

func (r *userRepository) GetAll(limit, offset int) ([]*model.User, error) {
	var users []*model.User
	err := r.db.Limit(limit).Offset(offset).Find(&users).Error
	return users, err
}

func (r *userRepository) UpdateRole(userID uint, newRole model.UserRole) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("role", newRole).Error
}

func (r *userRepository) UpdateActiveStatus(userID uint, isActive bool) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("is_active", isActive).Error
}

func (r *userRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.User{}).Count(&count).Error
	return count, err
}
