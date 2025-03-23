package model

import "time"

type Notification struct {
	ID        int       `db:"id"`
	UserID    string    `db:"user_id"`
	Title     string    `db:"title"`
	Message   string    `db:"message"`
	IsRead    bool      `db:"is_read"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
