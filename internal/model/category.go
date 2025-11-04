package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Category struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
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

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
