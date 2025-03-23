package model

import "time"

type Reservation struct {
	ReservationID       int       `db:"reservation_id"`
	UserID              string    `db:"user_id"`
	UserName            string    `db:"user_name"`
	Email               string    `db:"email"`
	ReservationDateTime time.Time `db:"reservation_datetime"`
	PetID               string    `db:"pet_id"`
	CreatedAt           time.Time `db:"created_at"`
	UpdatedAt           time.Time `db:"updated_at"`
	Status              string    `db:"status"` // pending, confirmed, cancelled
}
