package dto

type CreateGameRequest struct {
	CategoryID        uint     `json:"category_id" validate:"required"`
	Name              string   `json:"name" validate:"required,min=2"`
	Description       string   `json:"description,omitempty"`
	Platform          string   `json:"platform,omitempty"`
	Stock             int      `json:"stock" validate:"required,min=1"`
	RentalPricePerDay float64  `json:"rental_price_per_day" validate:"required,gt=0"`
	SecurityDeposit   float64  `json:"security_deposit" validate:"gte=0"`
	Condition         string   `json:"condition" validate:"omitempty,oneof=excellent good fair"`
	Images            []string `json:"images,omitempty"`
}

type UpdateGameRequest struct {
	CategoryID        uint    `json:"category_id,omitempty"` // HAPUS validate:"required"
	Name              string  `json:"name,omitempty"`
	Description       string  `json:"description,omitempty"`
	Platform          string  `json:"platform,omitempty"`
	Stock             int     `json:"stock,omitempty"`
	RentalPricePerDay float64 `json:"rental_price_per_day,omitempty"`
	SecurityDeposit   float64 `json:"security_deposit,omitempty"`
	Condition         string  `json:"condition,omitempty"`
}
