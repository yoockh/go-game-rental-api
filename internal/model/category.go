package model

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Name        string         `gorm:"uniqueIndex;not null" json:"name" validate:"required"`
	Description *string        `json:"description,omitempty"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Games []Game `gorm:"foreignKey:CategoryID" json:"-"`
}

func (Category) TableName() string {
	return "categories"
}
