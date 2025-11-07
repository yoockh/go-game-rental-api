package dto

type CreateBookingRequest struct {
	GameID    uint   `json:"game_id" validate:"required"`
	StartDate string `json:"start_date" validate:"required"` // String format YYYY-MM-DD
	EndDate   string `json:"end_date" validate:"required"`   // String format YYYY-MM-DD
	Notes     string `json:"notes,omitempty"`
}
