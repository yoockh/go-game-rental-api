package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type GameCondition string

const (
	ConditionExcellent GameCondition = "excellent"
	ConditionGood      GameCondition = "good"
	ConditionFair      GameCondition = "fair"
)

// StringArray for Swagger compatibility
type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return nil, nil
	}
	return json.Marshal(a)
}

func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = []string{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal StringArray")
	}
	return json.Unmarshal(bytes, a)
}

type Game struct {
	ID                uint          `gorm:"primaryKey" json:"id"`
	AdminID           uint          `gorm:"not null" json:"admin_id"`
	Admin             *User         `gorm:"foreignKey:AdminID" json:"admin,omitempty"`
	CategoryID        uint          `gorm:"not null" json:"category_id"`
	Category          *Category     `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Name              string        `gorm:"type:varchar(200);not null" json:"name"`
	Description       *string       `gorm:"type:text" json:"description"`
	Platform          *string       `gorm:"type:varchar(100)" json:"platform"`
	Stock             int           `gorm:"not null;default:0" json:"stock"`
	AvailableStock    int           `gorm:"not null;default:0" json:"available_stock"`
	RentalPricePerDay float64       `gorm:"type:decimal(10,2);not null" json:"rental_price_per_day"`
	SecurityDeposit   float64       `gorm:"type:decimal(10,2);not null" json:"security_deposit"`
	Condition         GameCondition `gorm:"type:varchar(20);not null" json:"condition"`
	Images            StringArray   `gorm:"type:text[]" json:"images" swaggertype:"array,string"`
	IsActive          bool          `gorm:"default:true" json:"is_active"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
}

func (Game) TableName() string {
	return "games"
}
