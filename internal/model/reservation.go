package model

import "time"

type Reservation struct {
	ID                  int64     `json:"id"`
	UserID              string    `json:"user_id"`
	UserName            string    `json:"user_name"`
	Email               string    `json:"email"`
	ReservationDateTime time.Time `json:"reservation_date_time"`
	PetID               string    `json:"pet_id"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	Status              string    `json:"status"` // pending, confirmed, cancelled
}
